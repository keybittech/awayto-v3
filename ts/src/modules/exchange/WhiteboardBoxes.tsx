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

  const handleMouseDown = useCallback((e: React.MouseEvent | React.TouchEvent, boxId: string) => {
    e.preventDefault();
    setDraggingId(boxId);
  }, []);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  const setCoordinates = (e: MouseEvent | React.Touch, draggingId: string) => {
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
      <WhiteboardBox sx={{ position: 'absolute' }} key={box.id} {...box} onTouchEnd={handleMouseUp}>
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
      </WhiteboardBox>
    ))}
  </>
}
