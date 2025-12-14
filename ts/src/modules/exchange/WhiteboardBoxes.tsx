import React, { useCallback, useState, useEffect } from 'react';

import IconButton from '@mui/material/IconButton';

import CloseIcon from '@mui/icons-material/Close';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';

import { DraggableBoxData, getRelativeCoordinates, targets } from 'awayto/hooks';

import WhiteboardBox from './WhiteboardBox';

interface WhiteboardBoxesProps extends IComponent {
  boxes: DraggableBoxData[];
  setBoxes: React.Dispatch<React.SetStateAction<DraggableBoxData[]>>;
  whiteboardRef: HTMLCanvasElement | null;
  didUpdate: (newBoxes: DraggableBoxData[]) => void;
}

export default function WhiteboardBoxes({ boxes, setBoxes, whiteboardRef, didUpdate }: WhiteboardBoxesProps): React.JSX.Element {

  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [resizingState, setResizingState] = useState<{
    id: string;
    startX: number;
    startY: number;
    startWidth: number;
    startHeight: number;
  } | null>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent | React.TouchEvent, boxId: string) => {
    e.preventDefault();
    setDraggingId(boxId);
    setActiveId(boxId);
  }, []);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
    setResizingState(null);
  }, []);

  const setCoordinates = (e: MouseEvent | React.Touch, draggingId: string) => {
    if (!whiteboardRef) return;

    const { x, y } = getRelativeCoordinates(e, whiteboardRef);

    const boxElem = document.getElementById(draggingId);
    const boxWidth = boxElem?.offsetWidth || 0;
    const boxHeight = boxElem?.offsetHeight || 0;

    let newX = Math.max(x, 0);
    if (boxWidth > 0) {
      newX = Math.min(newX, whiteboardRef.offsetWidth - boxWidth);
    }

    let newY = Math.max(y, 0);
    if (boxHeight > 0) {
      newY = Math.min(newY, whiteboardRef.offsetHeight - boxHeight);
    }

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

  const handleResizeStart = useCallback((e: React.MouseEvent | React.TouchEvent, boxId: string) => {
    if (!whiteboardRef) return;

    const eventData = 'touches' in e ? e.touches[0] : e.nativeEvent;

    const { x, y } = getRelativeCoordinates(eventData, whiteboardRef);
    const boxElem = document.getElementById(boxId); // Use the ID passed to the DOM

    if (boxElem) {
      setResizingState({
        id: boxId,
        startX: x,
        startY: y,
        startWidth: boxElem.offsetWidth,
        startHeight: boxElem.offsetHeight
      });
      setActiveId(boxId);
    }
  }, [whiteboardRef]);

  const setDimensions = (e: MouseEvent | React.Touch) => {
    if (!whiteboardRef || !resizingState) return;

    const { x: mouseX, y: mouseY } = getRelativeCoordinates(e, whiteboardRef);

    const deltaX = mouseX - resizingState.startX;
    const deltaY = mouseY - resizingState.startY;

    setBoxes(prev => {
      const updated = prev.map(box => {
        if (box.id !== resizingState.id) return box;

        let newWidth = resizingState.startWidth + deltaX;
        let newHeight = resizingState.startHeight + deltaY;

        newWidth = Math.max(newWidth, 80);
        newHeight = Math.max(newHeight, 80);

        if (box.x + newWidth > whiteboardRef.offsetWidth) {
          newWidth = whiteboardRef.offsetWidth - box.x;
        }
        if (box.y + newHeight > whiteboardRef.offsetHeight) {
          newHeight = whiteboardRef.offsetHeight - box.y;
        }

        return { ...box, width: newWidth, height: newHeight };
      });
      didUpdate(updated);
      return updated;
    });
  }

  useEffect(() => {
    if (draggingId || resizingState) {
      document.body.style.userSelect = 'none';
      const handleGlobalMouseMove = (e: MouseEvent) => {
        if (draggingId) setCoordinates(e, draggingId);
        if (resizingState) setDimensions(e);
      };
      const handleGlobalTouchMove = (e: TouchEvent) => {
        if (draggingId) setCoordinates(e.touches[0], draggingId);
        if (resizingState) setDimensions(e.touches[0]);
      };

      const handleGlobalMouseUp = () => {
        document.body.style.userSelect = 'auto';
        setDraggingId(null);
        setResizingState(null);
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
  }, [draggingId, resizingState, whiteboardRef, setBoxes, didUpdate]);

  return <>
    {boxes.map(box => (
      <WhiteboardBox
        {...box}
        sx={{ position: 'absolute' }}
        key={box.id}
        zIndex={activeId === box.id ? 1002 : 1001}
        onTouchEnd={handleMouseUp}
        onResizeStart={handleResizeStart}
      >
        <DragIndicatorIcon
          {...targets(`whiteboard boxes drag box ${box.id}`, `press and hold the mouse to drag this box over the whiteboard`)}
          sx={{
            color: 'black',
            position: 'absolute',
            top: '8px',
            left: '8px',
            cursor: draggingId === box.id ? "grabbing" : "grab",
          }}
          fontSize="small"
          onMouseDown={e => handleMouseDown(e, box.id)}
          onMouseUp={handleMouseUp}
          onTouchStart={e => handleMouseDown(e, box.id)}
        />
        <IconButton
          {...targets(`whiteboard boxes close box ${box.id}`, `close whiteboard text box`)}
          size="small"
          sx={{
            color: 'black',
            position: 'absolute',
            top: 0,
            right: '8px',
            mt: '4px'
          }}
          onClick={e => {
            e.preventDefault();
            setBoxes(prevBoxes => {
              const bid = prevBoxes.find(b => b.id == box.id);
              if (bid) {
                bid.text = "";
              }
              didUpdate(prevBoxes);
              return prevBoxes;
            });
          }}
        >
          <CloseIcon fontSize="small" />
        </IconButton>
      </WhiteboardBox>
    ))}
  </>
}
