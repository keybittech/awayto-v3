import React, { useCallback, useEffect, useRef, useState } from 'react';

import 'react-pdf/dist/esm/Page/TextLayer.css';

import { pdfjs } from 'react-pdf';
import * as _ from 'pdfjs-dist';

pdfjs.GlobalWorkerOptions.workerSrc = '/app/pdf.worker.min.mjs';

import { Document, Page } from 'react-pdf';

import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';

import { IFile, IWhiteboard, SocketActions, DraggableBoxData, useWebSocketSubscribe, useFileContents, useUtil, getRelativeCoordinates, generateLightBgColor } from 'awayto/hooks';
import WhiteboardOptionsMenu from './WhiteboardOptionsMenu';
import WhiteboardBoxes from './WhiteboardBoxes';

interface WhiteboardProps extends IComponent {
  topicId: string;
  optionsMenu: React.JSX.Element;
  chatOpen: boolean;
  chatBox: React.JSX.Element;
  callBox: React.JSX.Element;
  setSharedFile: React.Dispatch<React.SetStateAction<IFile | undefined>>;
  sharedFile: IFile | undefined;
}

export default function Whiteboard({ chatOpen, chatBox, callBox, optionsMenu, sharedFile, setSharedFile, topicId }: WhiteboardProps): React.JSX.Element {
  if (!topicId) return <></>;

  const whiteboardRef = useRef<HTMLCanvasElement>(null);
  const fileDisplayRef = useRef<HTMLCanvasElement>(null);
  const fileScroller = useRef<HTMLDivElement>(null);
  const scrollTimeout = useRef<number | null>(null);
  const boxDidUpdate = useRef<boolean>(false);
  const isDrawing = useRef<boolean>(false);
  const lastTouchPoint = useRef<{ x: number, y: number } | null>(null);

  const whiteboard = useRef<IWhiteboard>({
    boxes: [],
    lines: [],
    settings: {
      highlight: false,
      position: [0, 0],
      scale: 1,
      stroke: generateLightBgColor()
    }
  });

  const { openConfirm } = useUtil();

  const [canvasPointerEvents, setCanvasPointerEvents] = useState('none');
  const [zoom, setZoom] = useState(1);
  const [currentFile, setCurrentFile] = useState<IFile | null | undefined>(sharedFile);
  const { fileContents, getFileContents } = useFileContents();

  const [active, setActive] = useState(false);
  const [numPages, setNumPages] = useState(0);
  const [pageNumber, setPageNumber] = useState(1);
  const [fileToggle, setFileToggle] = useState(false);
  const [selectedText, setSelectedText] = useState<Record<string, string>>({});
  const [_, setBoards] = useState<Record<string, Partial<IWhiteboard>>>({});
  const [boxes, setBoxes] = useState<DraggableBoxData[]>([]);

  const {
    connectionId,
    userList,
    storeMessage: storeWhiteboardMessage,
    sendMessage: sendWhiteboardMessage
  } = useWebSocketSubscribe<IWhiteboard>(topicId, ({ sender, action, payload }) => {
    setBoards(b => {
      const senderBoard = b[sender] || {};
      const { settings, ...p } = payload;
      const board = { ...senderBoard, ...p, settings: { ...senderBoard.settings, ...settings } };
      if (SocketActions.SET_POSITION === action) {
        const [left, top] = board.settings?.position || [];
        fileScroller.current?.scrollTo({ left, top });
      } else if (SocketActions.SET_SCALE === action) {
        whiteboard.current.settings.scale = board.settings?.scale || 1;
        setZoom(whiteboard.current.settings.scale);
      } else if (SocketActions.SET_PAGE === action) {
        whiteboard.current.settings.page = board.settings?.page || 1;
        setPageNumber(whiteboard.current.settings.page);
      } else if (SocketActions.DRAW_LINES === action) {
        if (connectionId !== sender) {
          handleLines(payload.lines, board.settings?.stroke, board.settings?.highlight);
        }
        if (!payload.lines?.length && whiteboardRef.current) {
          whiteboardRef.current.width = whiteboardRef.current.width;
        }
      } else if (SocketActions.SHARE_FILE === action) {
        setNumPages(0);
        setPageNumber(1);
        if (!board.sharedFile) {
          setSharedFile(undefined);
          setCurrentFile(undefined);
          getFileContents({});
        } else {
          const fileDetails = { name: board.sharedFile?.name, mimeType: board.sharedFile?.mimeType, uuid: board.sharedFile?.uuid };
          if (connectionId !== sender) {
            for (const user of Object.values(userList)) {
              if (user?.cids.includes(sender)) {
                openConfirm({
                  isConfirming: true,
                  confirmEffect: `${user.name} wants to share a file`,
                  confirmAction: () => {
                    setCurrentFile(board.sharedFile);
                    getFileContents(fileDetails).catch(console.error);
                  }
                });
              }
            }
          } else {
            getFileContents(fileDetails).catch(console.error);
          }
        }
      } else if (SocketActions.SET_SELECTED_TEXT === action) {
        if (payload.selectedText?.length) {
          setSelectedText({ [sender]: payload.selectedText });
        }
      } else if (SocketActions.SET_BOX === action) {
        const pb = (payload.boxes || []);
        if (connectionId !== sender) {
          whiteboard.current.boxes = pb;
          setBoxes(pb);
        }
        // Check for deletions
        const validBoxes = pb.filter(b => b && b.text != '');
        if (validBoxes.length != pb.length) {
          whiteboard.current.boxes = validBoxes;
          setBoxes(validBoxes);
        }

      } else if (SocketActions.CHANGE_SETTING === action) {
      }
      return { ...b, [sender]: board };
    });
  });

  const handleLines = (lines?: IWhiteboard['lines'], stroke?: string, highlight?: boolean) => {
    const draw = () => {
      const ctx = whiteboardRef.current?.getContext('2d');
      if (!ctx) return;

      lines?.forEach((line, i) => {
        if (i === 0) {
          ctx.beginPath();
          ctx.moveTo(line.startPoint.x, line.startPoint.y);
        }

        ctx.lineTo(line.endPoint.x, line.endPoint.y);
      });

      ctx.strokeStyle = stroke || 'black';
      if (highlight) {
        ctx.lineWidth = 10;
        ctx.globalAlpha = .33;
      } else {
        ctx.lineWidth = ctx.globalAlpha = 1;
      }

      ctx.stroke();
    }

    requestAnimationFrame(draw);
  };

  // Modified to support both mouse and touch events
  const handleMouseDown = useCallback((event: React.MouseEvent<HTMLCanvasElement>) => {
    event.preventDefault();
    const canvas = whiteboardRef.current;
    if (!canvas) return;

    isDrawing.current = true;
    const startPoint = getRelativeCoordinates(event, canvas);
    lastTouchPoint.current = startPoint;

    const onMouseMove = (e: MouseEvent) => {
      if (!isDrawing.current) return;

      const endPoint = getRelativeCoordinates(e, canvas);
      const newLine = {
        startPoint: { ...lastTouchPoint.current! },
        endPoint: { ...endPoint }
      };
      whiteboard.current = {
        ...whiteboard.current,
        lines: [
          ...whiteboard.current.lines,
          newLine
        ],
      }

      handleLines([newLine], whiteboard.current.settings.stroke, whiteboard.current.settings.highlight);

      // Update lastPoint
      lastTouchPoint.current = endPoint;
    };

    const onMouseUp = () => {
      isDrawing.current = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }, []);

  // New touch event handlers
  const handleTouchStart = useCallback((event: React.TouchEvent<HTMLCanvasElement>) => {
    event.preventDefault();
    const canvas = whiteboardRef.current;
    if (!canvas) return;

    isDrawing.current = true;
    const touch = event.touches[0];
    const rect = canvas.getBoundingClientRect();
    const startPoint = {
      x: touch.clientX - rect.left,
      y: touch.clientY - rect.top
    };
    lastTouchPoint.current = startPoint;
  }, []);

  const handleTouchMove = useCallback((event: React.TouchEvent<HTMLCanvasElement>) => {
    event.preventDefault();
    if (!isDrawing.current) return;

    const canvas = whiteboardRef.current;
    if (!canvas || !lastTouchPoint.current) return;

    const touch = event.touches[0];
    const rect = canvas.getBoundingClientRect();
    const endPoint = {
      x: touch.clientX - rect.left,
      y: touch.clientY - rect.top
    };

    const newLine = {
      startPoint: { ...lastTouchPoint.current },
      endPoint: { ...endPoint }
    };

    whiteboard.current = {
      ...whiteboard.current,
      lines: [
        ...whiteboard.current.lines,
        newLine
      ],
    };

    handleLines([newLine], whiteboard.current.settings.stroke, whiteboard.current.settings.highlight);
    lastTouchPoint.current = endPoint;
  }, []);

  const handleTouchEnd = useCallback(() => {
    isDrawing.current = false;
    lastTouchPoint.current = null;
  }, []);

  // New scroll handling for touch events
  const handleFileScrollTouch = useCallback((_: React.TouchEvent<HTMLDivElement>) => {
    if (scrollTimeout.current) {
      clearTimeout(scrollTimeout.current);
    }
    scrollTimeout.current = setTimeout(() => {
      if (!fileScroller.current) return;
      const position = [fileScroller.current.scrollLeft, fileScroller.current.scrollTop];
      whiteboard.current.settings.position = position;
      sendWhiteboardMessage(SocketActions.SET_POSITION, { settings: { position } });
    }, 150);
  }, []);

  const textRenderer = useCallback((textItem: { str: string }) => {
    let newText = textItem.str;
    if (!newText.length) return textItem.str;
    for (const connId of Object.keys(selectedText)) {
      for (const user of Object.values(userList)) {
        if (user.cids.includes(connId)) {
          newText = newText.replace(selectedText[connId], txt => `<span style="background-color:${user.color};opacity:.5;">${txt}</span>`);
        }
      }
    }
    return newText;
  }, [selectedText, userList]);

  const sendBatchedData = () => {
    if (whiteboard.current.lines.length > 0) {
      storeWhiteboardMessage(SocketActions.DRAW_LINES, {
        lines: whiteboard.current.lines,
        settings: {
          stroke: whiteboard.current.settings.stroke,
          highlight: whiteboard.current.settings.highlight
        }
      });
      whiteboard.current = { ...whiteboard.current, lines: [] };
    }
    if (boxDidUpdate.current) {
      boxDidUpdate.current = false;
      storeWhiteboardMessage(SocketActions.SET_BOX, {
        boxes: whiteboard.current.boxes
      });
    }
  };

  useEffect(() => {
    setTimeout(() => {
      sendWhiteboardMessage(SocketActions.CHANGE_SETTING, {
        settings: whiteboard.current.settings
      });
    }, 1000);
    const interval = setInterval(sendBatchedData, 150);
    return () => clearInterval(interval);
  }, [connectionId]);

  useEffect(() => {
    if ((!!sharedFile && sharedFile.uuid !== currentFile?.uuid) || (!sharedFile && !!currentFile)) {
      setCurrentFile(sharedFile);
      sendWhiteboardMessage(SocketActions.SHARE_FILE, { sharedFile });
    }
  }, [sharedFile, currentFile]);

  useEffect(() => {
    const scrollDiv = fileScroller.current;
    if (scrollDiv) {
      const onFileScroll = (_: Event) => {
        if (scrollTimeout.current) {
          clearTimeout(scrollTimeout.current);
        }
        scrollTimeout.current = setTimeout(() => {
          const position = [scrollDiv.scrollLeft, scrollDiv.scrollTop];
          whiteboard.current.settings.position = position;
          sendWhiteboardMessage(SocketActions.SET_POSITION, { settings: { position } });
        }, 150);
      };

      scrollDiv.addEventListener('scroll', onFileScroll);

      return () => {
        if (scrollTimeout.current) {
          clearTimeout(scrollTimeout.current);
        }
        scrollDiv.removeEventListener('scroll', onFileScroll);
      };
    }
  }, [fileScroller.current]);

  useEffect(() => {
    if (null !== whiteboardRef.current && !sharedFile) {
      whiteboardRef.current.width = window.screen.width;
      whiteboardRef.current.height = window.screen.height;
    }
  }, [whiteboardRef, sharedFile]);

  useEffect(() => {
    if (selectedText[connectionId]?.length) {
      sendWhiteboardMessage(SocketActions.SET_SELECTED_TEXT, { selectedText: selectedText[connectionId].replace('\n', ' ') });
    }
  }, [selectedText[connectionId], userList]);

  return <Grid size="grow">
    <WhiteboardOptionsMenu
      {...{
        whiteboard: whiteboard.current,
        whiteboardRef: whiteboardRef.current,
        sendWhiteboardMessage,
        storeWhiteboardMessage,
        onCanvasInputChanged(inputMethod) {
          setCanvasPointerEvents('auto');
          switch (inputMethod) {
            case 'addedBox':
            case 'panning':
              setCanvasPointerEvents('none');
              break;
            case 'penning':
            case 'brushing':
              const highlight = 'brushing' == inputMethod;
              whiteboard.current.settings.highlight = highlight;
              sendWhiteboardMessage(SocketActions.CHANGE_SETTING, { settings: { highlight } });
              break;
            default:
              break;
          }
        },
        pageNumber,
        numPages,
      }}
      onBoxAdded={box => {
        const newBoxes = [...boxes, box];
        whiteboard.current.boxes = newBoxes;
        boxDidUpdate.current = true;
        setBoxes(newBoxes);
      }}
    >
      {optionsMenu}
    </WhiteboardOptionsMenu>

    <Grid
      container
      sx={{
        height: {
          xs: `calc(100% - ${140 + (!!sharedFile ? 12 : -68)}px)`,
          sm: `calc(100% - ${80 - (!!sharedFile ? 0 : 15)}px)`,
        }
      }}
    >
      {chatBox}

      <Stack sx={{ width: chatOpen ? 'calc(100% - 390px)' : '100%', height: '100%' }}>

        {callBox}

        {/* General Canvas Background  */}
        <Box
          onClick={() => !active && setActive(true)}
          ref={fileScroller}
          onTouchMove={handleFileScrollTouch}
          sx={{
            backgroundColor: fileContents ? '#ccc' : 'white',
            flex: 1,
            overflow: 'scroll',
            position: 'relative',
            padding: '16px',
            height: '100%',
            WebkitOverflowScrolling: 'touch', // Improve iOS scrolling
          }}
        >
          <WhiteboardBoxes
            boxes={boxes}
            setBoxes={setBoxes}
            whiteboardRef={whiteboardRef.current}
            didUpdate={(newBoxes) => {
              whiteboard.current.boxes = newBoxes;
              boxDidUpdate.current = true;
            }}
          />

          <Box // The canvas
            sx={{
              position: 'absolute',
              zIndex: 1002,
              pointerEvents: canvasPointerEvents,
              touchAction: 'none', // Prevent browser handling of touch gestures
            }}
            ref={whiteboardRef}
            component='canvas'
            onMouseDown={handleMouseDown}
            onTouchStart={handleTouchStart}
            onTouchMove={handleTouchMove}
            onTouchEnd={handleTouchEnd}
          />

          {!fileContents ? <></> : <Document // File Viewer
            file={fileContents?.url}
            onLoadSuccess={({ numPages }) => setNumPages(numPages)}
          >
            <Page
              scale={zoom}
              canvasRef={fileDisplayRef}
              renderAnnotationLayer={false}
              pageNumber={pageNumber}
              customTextRenderer={textRenderer}
              onRenderSuccess={() => {
                if (fileDisplayRef.current && whiteboardRef.current) {
                  const tempCanvas = document.createElement('canvas');
                  const tempCtx = tempCanvas.getContext('2d');
                  tempCanvas.width = whiteboardRef.current.width;
                  tempCanvas.height = whiteboardRef.current.height;
                  tempCtx?.drawImage(whiteboardRef.current, 0, 0);

                  whiteboardRef.current.getContext('2d')?.drawImage(tempCanvas, 0, 0);

                  setFileToggle(!fileToggle);
                }
              }}
              onMouseUp={() => {
                const selection = window.getSelection();
                const selectionString = selection?.toString();
                if (selectionString && selectionString != "") {
                  setSelectedText({ [connectionId]: selectionString });
                }
              }}
            />
          </Document>}
        </Box>
      </Stack>
    </Grid>
  </Grid >
}
