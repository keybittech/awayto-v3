import React, { useCallback, useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import { DraggableBoxData, getRelativeCoordinates, targets } from 'awayto/hooks';
import { MathJax } from 'better-react-mathjax';

interface WhiteboardBoxesProps extends IComponent {
  boxes: DraggableBoxData[];
  setBoxes: React.Dispatch<React.SetStateAction<DraggableBoxData[]>>;
  whiteboardRef: HTMLCanvasElement | null;
  didUpdate: () => void;
}

export default function WhiteboardBoxes({ boxes, setBoxes, whiteboardRef, didUpdate }: WhiteboardBoxesProps): React.JSX.Element {

  // const [updated, setUpdated] = useState(false);

  const [draggingId, setDraggingId] = useState<number | null>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent, boxId: number) => {
    e.preventDefault();
    setDraggingId(boxId);
  }, []);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  useEffect(() => {
    if (draggingId) {
      const handleGlobalMouseMove = (e: MouseEvent) => {
        if (!whiteboardRef) return;

        const { x, y } = getRelativeCoordinates(e, whiteboardRef);

        let newX = Math.max(x, 0);
        let newY = Math.max(y, 0);

        setBoxes(prevBoxes =>
          prevBoxes.map(box =>
            box.id === draggingId
              ? { ...box, x: newX, y: newY }
              : box
          )
        );
        didUpdate();
      };

      const handleGlobalMouseUp = () => {
        setDraggingId(null);
      };

      window.addEventListener('mousemove', handleGlobalMouseMove);
      window.addEventListener('mouseup', handleGlobalMouseUp);

      return () => {
        window.removeEventListener('mousemove', handleGlobalMouseMove);
        window.removeEventListener('mouseup', handleGlobalMouseUp);
      };
    }
  }, [draggingId]);

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
          padding: '24px 24px 0',
          bgcolor: box.color,
          color: '#222',
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          fontWeight: "bold",
          cursor: draggingId === box.id ? "grabbing" : "grab",
          userSelect: "none",
          boxShadow: 2,
          borderRadius: 1
        }}
        onMouseDown={e => handleMouseDown(e, box.id)}
        onMouseUp={handleMouseUp}
      >
        <MathJax>\[ {box.text} \]</MathJax>
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
            setBoxes(prevBoxes => prevBoxes.filter(b => b.id !== box.id));
          }}
        >
          X
        </IconButton>
      </Box>
    ))}
  </>
}
