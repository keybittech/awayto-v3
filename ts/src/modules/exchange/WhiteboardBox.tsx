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

export default function WhiteboardBox({ sx, x, y, color, text, children, ...props }: DraggableBoxData & BoxProps): React.JSX.Element {

  const renderTextWithMath = (rawText: string) => {
    if (!rawText) return null;

    const regex = /(\$\$[\s\S]*?\$\$|\$[^$\n]*\$)/g;
    const parts = rawText.split(regex);

    return parts.map((part, index) => {
      if (part.startsWith('$$') && part.endsWith('$$')) {
        return (
          <BlockMath
            key={index}
            math={part.slice(2, -2)}
            renderError={(err) => <span style={{ color: 'red' }}>{err.message}</span>}
          />
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
    sx={{
      ...sx,
      zIndex: 1001,
      pointerEvents: 'auto',
      left: x,
      top: y,
      padding: '32px 14px 12px',
      bgcolor: color,
      color: '#222',
      width: 'fit-content',
      minWidth: '80px',
      // maxWidth: '400px',
      wordBreak: 'break-word',
      overflowWrap: 'anywhere',
      whiteSpace: 'pre-line',
      borderRadius: 1,
      touchAction: 'none',
      '& .katex-html': {
        display: 'none',
      },
    }}
  >
    {children}
    {renderTextWithMath(text)}
  </Box>
}
