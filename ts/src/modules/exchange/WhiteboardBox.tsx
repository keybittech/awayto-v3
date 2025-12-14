import React from 'react';
import { InlineMath, BlockMath } from 'react-katex';

import { BoxProps, styled } from '@mui/material';
import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

import { DraggableBoxData } from 'awayto/hooks';

const Code = styled('span')({
  backgroundColor: 'rgba(255, 255, 255, 0.15)',
  fontFamily: 'monospace',
  padding: '2px 4px',
  borderRadius: '4px',
  fontSize: '0.9em',
  color: '#fff',
});

export const MathHelpTooltip = () => (
  <Tooltip
    arrow
    placement="top-end"
    title={
      <Box sx={{ p: 0.5, maxWidth: 220, lineHeight: 1.6 }}>
        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1 }}>
          Math Formatting
        </Typography>

        <Box sx={{ mb: 1.5 }}>
          <div><strong>Inline:</strong> Wrap text in <Code>$</Code></div>
          <div style={{ opacity: 0.8, fontSize: '0.85em', marginTop: 4 }}>
            Type: <Code>$2+2$</Code>
          </div>
        </Box>

        <Box>
          <div><strong>Block:</strong> Wrap text in <Code>$$</Code></div>
          <div style={{ opacity: 0.8, fontSize: '0.85em', marginTop: 4 }}>
            Type: <Code>{`$$\\frac{1}{2}$$`}</Code>
          </div>
        </Box>
      </Box>
    }
  >
    <Typography
      variant="caption"
      sx={{
        cursor: 'pointer',
        textDecoration: 'underline',
        textDecorationStyle: 'dotted',
      }}
    >
      Katex support
    </Typography>
  </Tooltip>
);

export default function WhiteboardBox({ id, sx, x, y, width, height, zIndex, color, text, children, onResizeStart, ...props }: DraggableBoxData & BoxProps): React.JSX.Element {

  const renderTextWithMath = (rawText: string) => {
    if (!rawText) return null;

    const regex = /(\$\$[\s\S]*?\$\$|\$[^$\n]*\$)/g;
    const parts = rawText.split(regex);

    return parts.map((part, index) => {
      if (part.startsWith('$$') && part.endsWith('$$')) {
        return (
          <Box key={index} sx={{ overflowX: 'auto', overflowY: 'hidden', maxWidth: '100%' }}>
            <BlockMath
              key={index}
              math={part.slice(2, -2)}
              renderError={(err) => <span style={{ color: 'red' }}>{err.message}</span>}
            />
          </Box>
        );
      }
      else if (part.startsWith('$') && part.endsWith('$')) {
        return (
          <InlineMath
            key={index}
            math={part.slice(1, -1)}
            renderError={(err) => <span style={{ color: 'red' }}>{err.message}</span>}
          />
        );
      }
      else {
        if (!part) return null;
        return <span key={index}>{part}</span>;
      }
    });
  };

  return <Box
    {...props}
    id={id}
    style={{ // dynamic properties need to be inlined to prevent emotion style tags in <head>
      left: x,
      top: y,
      width: width ? `${width}px` : 'max-content',
      height: height ? `${height}px` : 'auto',
      zIndex: zIndex || 1001,
    }}
    sx={{
      ...sx,
      pointerEvents: 'auto',
      minWidth: '80px',
      minHeight: '80px',
      bgcolor: color,
      color: '#222',
      borderRadius: 1,
      touchAction: 'none',
      display: 'flex',
      flexDirection: 'column',
      overflow: 'hidden',
      padding: 0,
      '& .katex-html': {
        display: 'none',
      },
    }}
  >
    {children}
    <Box sx={{ height: '32px', width: '100%', flexShrink: 0 }} />
    <Box
      sx={{
        flex: 1,
        overflow: 'auto',
        padding: '0 14px 12px',
        width: '100%',
        wordBreak: 'break-word',
        overflowWrap: 'anywhere',
        whiteSpace: 'pre-wrap',
        '&::-webkit-scrollbar': { width: '6px', height: '6px' },
        '&::-webkit-scrollbar-track': { background: 'transparent' },
        '&::-webkit-scrollbar-thumb': { backgroundColor: 'rgba(0,0,0,0.2)', borderRadius: '4px' },
        '&::-webkit-scrollbar-thumb:hover': { backgroundColor: 'rgba(0,0,0,0.4)' },
      }}
    >
      {renderTextWithMath(text)}
    </Box>
    <Box // resize handle
      onMouseDown={(e) => {
        e.stopPropagation();
        onResizeStart(e, id);
      }}
      onTouchStart={(e) => {
        e.stopPropagation();
        onResizeStart(e, id);
      }}
      sx={{
        position: 'absolute',
        bottom: 0,
        right: 0,
        width: '15px',
        height: '15px',
        cursor: 'se-resize',
        zIndex: 1002,
        '&::after': {
          content: '""',
          position: 'absolute',
          bottom: '4px',
          right: '4px',
          width: '0',
          height: '0',
          borderStyle: 'solid',
          borderWidth: '0 0 10px 10px',
          borderColor: 'transparent transparent rgba(0,0,0,0.3) transparent',
        }
      }}
    />
  </Box>
}
