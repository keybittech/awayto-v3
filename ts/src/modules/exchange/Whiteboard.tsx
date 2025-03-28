
import React, { useCallback, useEffect, useRef, useState } from 'react';

import 'react-pdf/dist/esm/Page/TextLayer.css';

import { pdfjs } from 'react-pdf';
import * as _ from 'pdfjs-dist';

pdfjs.GlobalWorkerOptions.workerSrc = '/app/pdf.worker.min.mjs';

import { Document, Page } from 'react-pdf';

import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';

import { IFile, IWhiteboard, useWebSocketSubscribe, useFileContents, useUtil, SocketActions } from 'awayto/hooks';
import WhiteboardOptionsMenu from './WhiteboardOptionsMenu';
import WhiteboardBoxes, { DraggableBoxData } from './WhiteboardBoxes';

function getRelativeCoordinates(event: MouseEvent | React.MouseEvent<HTMLCanvasElement>, canvas: HTMLCanvasElement) {
  const rect = canvas.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const y = event.clientY - rect.top;
  return { x, y };
}

// onwhiteboard load use effect check fileDetails from modal close then do a confirm action to getFileContents ?

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
  const scrollTimeout = useRef<NodeJS.Timeout | null>(null);

  const whiteboard = useRef<IWhiteboard>({ lines: [], settings: { highlight: false, position: [0, 0] } });

  const { openConfirm } = useUtil();

  const [canvasPointerEvents, setCanvasPointerEvents] = useState('none');
  const [zoom, setZoom] = useState(1);
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
      const board = { ...b[sender], ...payload };
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
          handleLines(payload.lines, board.settings);
        }
      } else if (SocketActions.SHARE_FILE === action) {
        setNumPages(0);
        setPageNumber(1);
        if (!board.sharedFile) {
          setSharedFile(undefined);
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
      } else if (SocketActions.CHANGE_SETTING === action) {
      }
      return { ...b, [sender]: board };
    });
  });

  const handleLines = (lines?: IWhiteboard['lines'], settings?: IWhiteboard['settings']) => {
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

      ctx.strokeStyle = settings?.stroke || 'black';
      if (settings?.highlight) {
        ctx.lineWidth = 10;
        ctx.globalAlpha = .33;
      } else {
        ctx.lineWidth = ctx.globalAlpha = 1;
      }

      ctx.stroke();
    }

    requestAnimationFrame(draw);
  };

  const handleMouseDown = useCallback((event: React.MouseEvent<HTMLCanvasElement>) => {
    event.preventDefault();
    const canvas = whiteboardRef.current;
    if (!canvas) return;

    const startPoint = getRelativeCoordinates(event, canvas);
    let lastPoint = startPoint;

    const onMouseMove = (e: MouseEvent) => {
      const endPoint = getRelativeCoordinates(e, canvas);
      const newLine = {
        startPoint: { ...lastPoint },
        endPoint: { ...endPoint }
      };
      whiteboard.current = {
        ...whiteboard.current,
        lines: [
          ...whiteboard.current.lines,
          newLine
        ],
      }

      handleLines([newLine], whiteboard.current.settings);

      // Update lastPoint
      lastPoint = endPoint;
    };

    const onMouseUp = () => {
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }, []);

  const sendBatchedData = () => {
    if (whiteboard.current.lines.length > 0) {
      storeWhiteboardMessage(SocketActions.DRAW_LINES, { lines: whiteboard.current.lines });
      whiteboard.current = { ...whiteboard.current, lines: [] };
    }
  };

  const textRenderer = useCallback((textItem: { str: string }) => {
    let newText = textItem.str;
    for (const connId of Object.keys(selectedText)) {
      for (const user of Object.values(userList)) {
        if (user.cids.includes(connId)) {
          newText = newText.replace(selectedText[connId], txt => `<span style="background-color:${user.color};opacity:.3;">${txt}</span>`);
        }
      }
    }
    return newText;
  }, [selectedText, userList]);

  useEffect(() => {
    const interval = setInterval(sendBatchedData, 150);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    sendWhiteboardMessage(SocketActions.SHARE_FILE, { sharedFile });
  }, [sharedFile]);

  useEffect(() => {
    const scrollDiv = fileScroller.current;
    if (scrollDiv) {

      const onFileScroll = (_: Event) => {
        if (scrollTimeout.current) {
          clearTimeout(scrollTimeout.current);
        }
        scrollTimeout.current = setTimeout(() => {
          whiteboard.current.settings.position = [scrollDiv.scrollLeft, scrollDiv.scrollTop];
          sendWhiteboardMessage(SocketActions.SET_POSITION, whiteboard.current);
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
        setBoxes([...boxes, box]);
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
          sx={{
            backgroundColor: fileContents ? '#ccc' : 'white',
            flex: 1,
            overflow: 'scroll',
            position: 'relative',
            padding: '16px',
            height: '100%',
          }}
        >
          <WhiteboardBoxes boxes={boxes} setBoxes={setBoxes} whiteboardRef={whiteboardRef.current} />

          <Box // The canvas
            sx={{
              position: 'absolute',
              zIndex: 100,
              pointerEvents: canvasPointerEvents
            }}
            ref={whiteboardRef}
            component='canvas'
            onMouseDown={handleMouseDown}
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
                const selection = document.getSelection();
                if (selection && selection.toString() != "") {
                  setSelectedText({ [connectionId]: selection.toString() });
                }
              }}
            />
          </Document>}
        </Box>
      </Stack>
    </Grid>
  </Grid >
}
