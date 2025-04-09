import React, { useCallback, useState, useEffect, useMemo } from 'react';
import { MathJax } from 'better-react-mathjax';

import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';

import CloseIcon from '@mui/icons-material/Close';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';

import { DraggableBoxData, getRelativeCoordinates, targets } from 'awayto/hooks';

declare global {
  interface Window {
    MathJax: {
      options: Record<string, any>
    }
  }
}

window.MathJax = {
  options: {
    enableMenu: false
  },
}

interface WhiteboardBoxesProps extends IComponent {
  boxes: DraggableBoxData[];
  setBoxes: React.Dispatch<React.SetStateAction<DraggableBoxData[]>>;
  whiteboardRef: HTMLCanvasElement | null;
  didUpdate: (newBoxes: DraggableBoxData[]) => void;
}

export default function WhiteboardBoxes({ boxes, setBoxes, whiteboardRef, didUpdate }: WhiteboardBoxesProps): React.JSX.Element {

  const [draggingId, setDraggingId] = useState<number | null>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent | React.TouchEvent, boxId: number) => {
    e.preventDefault();
    setDraggingId(boxId);
  }, []);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  // MathJax is slow to drag around unless memoized
  const boxComponents = useMemo(() => {
    return boxes.reduce((m, d) => {
      return {
        ...m,
        [d.id]: <MathJax>\[ {d.text} \]</MathJax>
      }
    }, {} as Record<number, React.JSX.Element>)
  }, [boxes]);

  const setCoordinates = (e: MouseEvent | React.Touch, draggingId: number) => {
    if (!whiteboardRef) return;

    const { x, y } = getRelativeCoordinates(e, whiteboardRef);

    let newX = Math.max(x, 0);
    let newY = Math.max(y, 0);

    setBoxes(prevBoxes => {
      const updatedBoxes = prevBoxes.map(box =>
        box.id === draggingId
          ? { ...box, x: newX, y: newY }
          : box
      );
      didUpdate(updatedBoxes);
      return updatedBoxes;
    });
  }

  useEffect(() => {
    if (draggingId) {
      const handleGlobalMouseMove = (e: MouseEvent) => {
        setCoordinates(e, draggingId);
      };
      const handleGlobalTouchMove = (e: TouchEvent) => {
        setCoordinates(e.touches[0], draggingId);
      };

      const handleGlobalMouseUp = () => {
        setDraggingId(null);
      };

      window.addEventListener('touchmove', handleGlobalTouchMove);
      window.addEventListener('touchend', handleGlobalMouseUp);
      window.addEventListener('mousemove', handleGlobalMouseMove);
      window.addEventListener('mouseup', handleGlobalMouseUp);

      return () => {
        window.removeEventListener('touchmove', handleGlobalTouchMove);
        window.removeEventListener('touchend', handleGlobalMouseUp);
        window.removeEventListener('mousemove', handleGlobalMouseMove);
        window.removeEventListener('mouseup', handleGlobalMouseUp);
      };
    }
  }, [draggingId, whiteboardRef, setBoxes, didUpdate]);

  return <>
    {boxes.map(box => (
      <Box
        key={box.id}
        sx={{
          zIndex: 1001,
          pointerEvents: 'auto',
          position: "absolute",
          left: box.x,
          top: box.y,
          padding: '32px 14px 12px',
          bgcolor: box.color,
          color: '#222',
          width: 'max-content',
          minWidth: '80px',
          maxWidth: '400px',
          borderRadius: 1,
          touchAction: 'none' // Prevent default touch actions
        }}
        onTouchEnd={handleMouseUp}
      >
        <DragIndicatorIcon
          {...targets(`whiteboard boxes drag box ${box.id}`, `press and hold the mouse to drag this box over the whiteboard`)}
          sx={{
            color: 'black',
            position: 'absolute',
            top: '6px',
            left: '8px',
            cursor: draggingId === box.id ? "grabbing" : "grab",
          }}
          fontSize="small"
          onMouseDown={e => handleMouseDown(e, box.id)}
          onMouseUp={handleMouseUp}
          onTouchStart={e => handleMouseDown(e, box.id)}
        />
        {boxComponents[box.id]}
        <IconButton
          {...targets(`whiteboard boxes close box ${box.id}`, `close whiteboard text box`)}
          size="small"
          sx={{
            color: 'black',
            position: 'absolute',
            top: 0,
            right: '8px'
          }}
          onClick={e => {
            e.preventDefault();
            setBoxes(prevBoxes => {
              const newBoxes = prevBoxes.filter(b => b.id !== box.id);
              didUpdate(newBoxes);
              return newBoxes;
            });
          }}
        >
          <CloseIcon fontSize="small" />
        </IconButton>
      </Box>
    ))}
  </>
}
