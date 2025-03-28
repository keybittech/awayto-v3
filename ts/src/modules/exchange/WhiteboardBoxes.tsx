import React, { useCallback, useState, useEffect } from 'react';

import Box from '@mui/material/Box';

export interface DraggableBoxData {
  id: string;
  x: number;
  y: number;
  color: string;
  component: React.JSX.Element;
}

interface WhiteboardBoxesProps extends IComponent {
  boxes: DraggableBoxData[];
  setBoxes: React.Dispatch<React.SetStateAction<DraggableBoxData[]>>;
  whiteboardRef: HTMLCanvasElement | null;
}

export default function WhiteboardBoxes({ boxes, setBoxes, whiteboardRef }: WhiteboardBoxesProps): React.JSX.Element {

  // const thing = "\\lim_{x \\to c} f^*(x) = \\lim_{x \\to c} \\frac{f(x) - f(c)}{x - c} = f'(c)";

  const [draggingId, setDraggingId] = useState<string | null>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent, boxId: string) => {
    e.preventDefault();
    setDraggingId(boxId);
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!draggingId || !whiteboardRef) return;

    const rect = whiteboardRef.getBoundingClientRect();

    const newX = e.clientX - rect.left;
    const newY = e.clientY - rect.top;

    setBoxes(prevBoxes =>
      prevBoxes.map(box =>
        box.id === draggingId
          ? { ...box, x: newX, y: newY }
          : box
      )
    );
  }, [draggingId]);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  useEffect(() => {
    if (draggingId) {
      const handleGlobalMouseMove = (e: MouseEvent) => {
        if (!whiteboardRef) return;

        const rect = whiteboardRef.getBoundingClientRect();

        const newX = e.clientX - rect.left;
        const newY = e.clientY - rect.top;

        setBoxes(prevBoxes =>
          prevBoxes.map(box =>
            box.id === draggingId
              ? { ...box, x: newX, y: newY }
              : box
          )
        );
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

  return <Box sx={{ zIndex: 1001 }}>
    {boxes.map(box => (
      <Box
        key={box.id}
        sx={{
          pointerEvents: 'auto',
          position: "absolute",
          left: box.x,
          top: box.y,
          padding: '10px',
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
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
      >
        {box.component}
      </Box>
    ))}
  </Box>
}
