
/**
 * need to capture the idea where runons, etc. can be highlighted throughout a page after post processing
 */
import React, { useCallback, useEffect, useRef, useState } from 'react';

import 'react-pdf/dist/esm/Page/TextLayer.css';

import { pdfjs } from 'react-pdf';

pdfjs.GlobalWorkerOptions.workerSrc = '../static/js/pdf.worker.min.mjs'

import { Document, Page } from 'react-pdf';

import Box from '@mui/material/Box';

import { IFile, Whiteboard, useWebSocketSubscribe, useFileContents, useComponents, useUtil, SocketActions } from 'awayto/hooks';

function getRelativeCoordinates(event: MouseEvent | React.MouseEvent<HTMLCanvasElement>, canvas: HTMLCanvasElement) {
  const rect = canvas.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const y = event.clientY - rect.top;
  return { x, y };
}

// onwhiteboard load use effect check fileDetails from modal close then do a confirm action to getFileContents ?

declare global {
  interface IProps {
    optionsMenu?: JSX.Element;
    sharedFile?: IFile;
    topicId?: string;
    openFileSelect?: () => void;
  }
}

export default function Whiteboard({ optionsMenu, sharedFile, openFileSelect, topicId }: IProps): React.JSX.Element {
  if (!topicId) return <></>;

  const whiteboardRef = useRef<HTMLCanvasElement>(null);
  // const contextRef = useRef<CanvasRenderingContext2D | null>(null);
  const fileDisplayRef = useRef<HTMLCanvasElement>(null);
  const fileScroller = useRef<HTMLDivElement>(null);
  const scrollTimeout = useRef<NodeJS.Timeout | null>(null);

  const whiteboard = useRef<Whiteboard>({ lines: [], settings: { highlight: false, position: [0, 0] } });

  const { openConfirm } = useUtil();
  const { WhiteboardOptionsMenu } = useComponents();

  const [canvasPointerEvents, setCanvasPointerEvents] = useState('none');
  const [zoom, setZoom] = useState(1);
  const { fileContents, getFileContents } = useFileContents();

  const [active, setActive] = useState(false);
  const [numPages, setNumPages] = useState(0);
  const [pageNumber, setPageNumber] = useState(1);
  const [strokeColor, setStrokeColor] = useState('#aaaaaa');
  const [fileToggle, setFileToggle] = useState(false);
  const [boards, setBoards] = useState<Record<string, Partial<Whiteboard>>>({});

  const {
    connectionId,
    userList,
    storeMessage: storeWhiteboardMessage,
    sendMessage: sendWhiteboardMessage
  } = useWebSocketSubscribe<Whiteboard>(topicId, ({ sender, action, payload }) => {
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
      } else if (SocketActions.CHANGE_SETTING === action) {
      }
      return { ...b, [sender]: board };
    });
  });

  const handleLines = (lines?: Whiteboard['lines'], settings?: Whiteboard['settings']) => {
    const draw = () => {
      // const canvas = whiteboardRef.current;
      // if (!canvas) return;
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

  useEffect(() => {
    const interval = setInterval(sendBatchedData, 150);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (sharedFile && sharedFile.uuid !== fileContents?.uuid) {
      console.log({ SHARING: sharedFile });
      sendWhiteboardMessage(SocketActions.SHARE_FILE, { sharedFile });
    }
  }, [fileContents, sharedFile]);

  useEffect(() => {
    const scrollDiv = fileScroller.current;
    if (scrollDiv) {

      const onFileScroll = (e: Event) => {
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
      // contextRef.current = whiteboardRef.current.getContext('2d');
    }
  }, [whiteboardRef, sharedFile]);

  return <Box
    onClick={() => !active && setActive(true)}
    sx={{
      flex: 1,
      position: 'relative',
      display: 'flex',
      width: '100%'
    }}
  >

    {/* General Canvas Background  */}
    <Box
      ref={fileScroller}
      sx={{
        backgroundColor: fileContents ? '#ccc' : 'white',
        flex: 1,
        overflow: 'scroll',
        position: 'relative',
        padding: '16px'
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          zIndex: 100,
          pointerEvents: canvasPointerEvents
        }}
        ref={whiteboardRef}
        component='canvas'
        onMouseDown={handleMouseDown}
      />

      {!fileContents || !sharedFile ? <></> : <Document // File Viewer
        file={fileContents.url}
        onLoadSuccess={({ numPages }) => setNumPages(numPages)}
      >
        <Page
          scale={zoom}
          canvasRef={fileDisplayRef}
          renderAnnotationLayer={false}
          pageNumber={pageNumber}
          onRenderSuccess={() => {
            if (fileDisplayRef.current && whiteboardRef.current) {
              const tempCanvas = document.createElement('canvas');
              const tempCtx = tempCanvas.getContext('2d');
              tempCanvas.width = whiteboardRef.current.width;
              tempCanvas.height = whiteboardRef.current.height;
              tempCtx?.drawImage(whiteboardRef.current, 0, 0);

              const { width, height } = fileDisplayRef.current.getBoundingClientRect();
              whiteboardRef.current.width = width;
              whiteboardRef.current.height = height;

              whiteboardRef.current.getContext('2d')?.drawImage(tempCanvas, 0, 0);

              // fileDisplayRef.current.getContext('2d')?.translate(0, 0);
              setFileToggle(!fileToggle);
            }

          }}
        />
      </Document>}
    </Box>

    <WhiteboardOptionsMenu
      {...{
        whiteboard: whiteboard.current,
        strokeColor,
        setStrokeColor,
        pageNumber,
        numPages,
        setNumPages,
        scale: zoom,
        canvasPointerEvents,
        setCanvasPointerEvents,
        // contextRef: contextRef.current,
        whiteboardRef: whiteboardRef.current,
        sendWhiteboardMessage
      }}
    >
      {optionsMenu}
    </WhiteboardOptionsMenu>
  </Box >
}
