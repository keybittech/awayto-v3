import React, { useCallback, useState, useEffect } from 'react';

import Box from '@mui/material/Box';

interface DraggableBoxData {
  id: string;
  x: number;
  y: number;
  color: string;
  label: string;
}

interface WhiteboardBoxesProps extends IComponent {
  whiteboardRef: HTMLCanvasElement | null;
}

export default function WhiteboardBoxes({ whiteboardRef }: WhiteboardBoxesProps): React.JSX.Element {

  const [boxes, setBoxes] = useState<DraggableBoxData[]>([
    { id: '1', x: 50, y: 50, color: '#3f51b5', label: 'Box 1' },
    { id: '2', x: 200, y: 100, color: '#f50057', label: 'Box 2' }
  ]);

  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [mouseOffset, setMouseOffset] = useState({ x: 0, y: 0 });

  // const addBox = () => {
  //   const newId = `${boxes.length + 1}`;
  //   const colors = ['#3f51b5', '#f50057', '#009688', '#ff9800', '#9c27b0', '#2196f3'];
  //   const randomColor = colors[Math.floor(Math.random() * colors.length)];
  //
  //   setBoxes([
  //     ...boxes,
  //     {
  //       id: newId,
  //       x: Math.random() * 300,
  //       y: Math.random() * 200,
  //       color: randomColor,
  //       label: `Box ${newId}`
  //     }
  //   ]);
  // };

  const handleMouseDown = useCallback((e: React.MouseEvent, boxId: string) => {
    e.preventDefault();
    setDraggingId(boxId);

    const box = boxes.find(b => b.id === boxId);
    if (!box) return;

    setMouseOffset({
      x: e.clientX - box.x,
      y: e.clientY - box.y
    });
  }, [boxes]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!draggingId || !whiteboardRef) return;

    const rect = whiteboardRef.getBoundingClientRect();

    const newX = e.clientX - rect.left - mouseOffset.x;
    const newY = e.clientY - rect.top - mouseOffset.y;

    setBoxes(prevBoxes =>
      prevBoxes.map(box =>
        box.id === draggingId
          ? { ...box, x: newX, y: newY }
          : box
      )
    );
  }, [draggingId, mouseOffset]);

  const handleMouseUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  useEffect(() => {
    if (draggingId) {
      const handleGlobalMouseMove = (e: MouseEvent) => {
        if (!whiteboardRef) return;

        const rect = whiteboardRef.getBoundingClientRect();

        const newX = e.clientX - rect.left - mouseOffset.x;
        const newY = e.clientY - rect.top - mouseOffset.y;

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
  }, [draggingId, mouseOffset]);

  return <>
    {boxes.map(box => (
      <Box
        key={box.id}
        sx={{
          position: "absolute",
          left: box.x,
          top: box.y,
          width: "100px",
          height: "100px",
          bgcolor: box.color,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          color: "white",
          fontWeight: "bold",
          cursor: draggingId === box.id ? "grabbing" : "grab",
          userSelect: "none",
          boxShadow: 2,
          borderRadius: 1
        }}
        onMouseDown={(e) => handleMouseDown(e, box.id)}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
      >
        {box.label}
      </Box>
    ))}
  </>
}
