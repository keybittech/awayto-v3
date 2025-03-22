import React, { useEffect } from 'react';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';
import Divider from '@mui/material/Divider';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';

import EditIcon from '@mui/icons-material/Edit';
import BrushIcon from '@mui/icons-material/Brush';
import LayersClearIcon from '@mui/icons-material/LayersClear';
import PanToolIcon from '@mui/icons-material/PanTool';
import ZoomInIcon from '@mui/icons-material/ZoomIn';
import MenuBookIcon from '@mui/icons-material/MenuBook';
import AddIcon from '@mui/icons-material/Add';
import RemoveIcon from '@mui/icons-material/Remove';
import KeyboardArrowLeftIcon from '@mui/icons-material/KeyboardArrowLeft';
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight';
import KeyboardDoubleArrowLeftIcon from '@mui/icons-material/KeyboardDoubleArrowLeft';
import KeyboardDoubleArrowRightIcon from '@mui/icons-material/KeyboardDoubleArrowRight';
import InsertPageBreak from '@mui/icons-material/InsertPageBreak';

import { SocketActions, useDebounce, IWhiteboard, targets } from 'awayto/hooks';

const scales = [.1, .25, .5, .8, 1, 1.25, 1.5, 2, 2.5, 3, 4];

interface WhiteboardOptionsMenuProps extends IComponent {
  whiteboard: IWhiteboard;
  whiteboardRef: HTMLCanvasElement | null;
  canvasPointerEvents: string;
  strokeColor: string;
  setStrokeColor: React.Dispatch<React.SetStateAction<string>>;
  numPages: number;
  pageNumber: number;
  scale: number;
  setCanvasPointerEvents: React.Dispatch<React.SetStateAction<string>>;
  sendWhiteboardMessage: (action: SocketActions, payload?: Partial<IWhiteboard> | undefined) => void;
}

export function WhiteboardOptionsMenu({
  children,
  whiteboardRef,
  whiteboard,
  strokeColor,
  setStrokeColor,
  pageNumber,
  numPages,
  scale,
  canvasPointerEvents,
  setCanvasPointerEvents,
  sendWhiteboardMessage,
}: WhiteboardOptionsMenuProps): React.JSX.Element {

  const panning = 'none' === canvasPointerEvents;
  const penning = !whiteboard.settings.highlight;

  const setDrawStyle = (hl: boolean) => {
    whiteboard.settings.highlight = hl;
    sendWhiteboardMessage(SocketActions.CHANGE_SETTING, { settings: { highlight: hl } });
    setPointerEvents('auto');
  };

  const setPointerEvents = (style: string) => {
    setCanvasPointerEvents(style);
    // handleMenuClose();
  }

  const setScale = (inc: boolean | number) => {
    const scale = whiteboard.settings.scale || 1;
    const nextScale = 'number' === typeof inc ?
      inc :
      inc ? scales[Math.min(scales.indexOf(scale) + 1, scales.length)] : (scales[(scales.indexOf(scale) - 1 || 0)]);
    if (nextScale > 0 && nextScale <= 4) {
      sendWhiteboardMessage(SocketActions.SET_SCALE, { settings: { scale: nextScale } });
    }
  };

  const setPage = (next: boolean | number) => {
    let page = pageNumber || 1;
    next ? page++ : page--;
    const nextPage = 'number' === typeof next ?
      Math.min(next, numPages) :
      next ? Math.min(page, numPages) : (page || 1);
    sendWhiteboardMessage(SocketActions.SET_PAGE, { settings: { page: Math.max(1, nextPage) } });
  };

  const debouncedStroke = useDebounce(strokeColor, 1000);

  useEffect(() => {
    sendWhiteboardMessage(SocketActions.SET_STROKE, { settings: { stroke: debouncedStroke } });
    whiteboard.settings.stroke = debouncedStroke;
  }, [debouncedStroke]);

  return <AppBar position="static">
    <Toolbar>

      <Grid container size="grow" sx={{ py: { xs: 2, sm: 0 } }}>
        <Grid container size={{ xs: 12, xl: 'auto' }}>
          {children}

          <Divider sx={{ mx: 2, display: { xs: 'none', sm: 'flex' } }} orientation="vertical" variant="middle" flexItem />

          <Tooltip title="Pan">
            <IconButton
              {...targets(`whiteboard pan`, `set the mouse to pan mode in order to drag the whiteboard around`)}
              onClick={() => setPointerEvents('none')}
            >
              <PanToolIcon color={panning ? 'info' : 'primary'} />
            </IconButton>
          </Tooltip>

          <Tooltip title="Pen">
            <IconButton
              {...targets(`whiteboard pen`, `set the mouse to pen drawing mode`)}
              onClick={() => setDrawStyle(false)}
            >
              <EditIcon color={!panning && penning ? 'info' : 'primary'} />
            </IconButton>
          </Tooltip>

          <Tooltip title="Brush">
            <IconButton
              {...targets(`whiteboard brush`, `set the mouse to brush drawing mode`)}
              onClick={() => setDrawStyle(true)}
            >
              <BrushIcon color={!panning && !penning ? 'info' : 'primary'} />
            </IconButton>
          </Tooltip>

          <Tooltip title="Select Color">
            <IconButton
              {...targets(`whiteboard color select`, `select a color to draw with`)}
            >
              <Box
                sx={{
                  width: '24px',
                  height: '24px',
                  borderRadius: '24px',
                  overflow: 'hidden'
                }}
              >
                <Box
                  sx={{
                    border: 'none',
                    width: '200%',
                    height: '200%',
                    cursor: 'pointer',
                    transform: 'translate(-25%, -25%)'
                  }}
                  component="input"
                  type="color"
                  value={strokeColor}
                  onChange={e => setStrokeColor(e.target.value)}
                />
              </Box>
            </IconButton>
          </Tooltip>

          <Tooltip title="Clear Canvas">
            <IconButton
              {...targets(`whiteboard clear canvas`, `remove all drawings from whiteboard`)}
              color="error"
              onClick={() => {
                if (whiteboardRef) {
                  whiteboardRef.getContext('2d')?.clearRect(0, 0, whiteboardRef.width, whiteboardRef.height)
                }
              }}
            >
              <LayersClearIcon />
            </IconButton>
          </Tooltip>
        </Grid>


        {numPages > 0 && <Grid container size={{ xs: 12, sm: 'auto' }}>
          <Divider sx={{ mx: 2, display: { xs: 'none', xl: 'flex' } }} orientation="vertical" variant="middle" flexItem />
          <Tooltip title="Zoom Out" children={
            <IconButton
              {...targets(`whiteboard zoom out`, `zoom the whiteboard out`)}
              onClick={() => setScale(false)}
            >
              <RemoveIcon fontSize="small" />
            </IconButton>
          } />
          <TextField
            {...targets(`whiteboard zoom select`, ``, `change the whiteboard zoom setting`)}
            select
            variant="standard"
            value={scale}
            onChange={e => setScale(parseFloat(e.target.value))}
            slotProps={{
              input: {
                startAdornment: <ZoomInIcon sx={{ mr: 1 }} />
              }
            }}
          >
            {scales.map(v => <MenuItem key={v} value={v}>{Math.round(parseFloat(v.toFixed(2)) * 100)}%</MenuItem>)}
          </TextField>
          <Tooltip title="Zoom In" children={
            < IconButton
              {...targets(`whiteboard zoom in`, `zoom the whiteboard in`)}
              onClick={() => setScale(true)}
            >
              <AddIcon fontSize="small" />
            </IconButton>
          } />

        </Grid>}

        {numPages > 1 && <Grid container size="auto">
          <Divider sx={{ mx: 2, display: { xs: 'none', sm: 'flex' } }} orientation="vertical" variant="middle" flexItem />
          <Tooltip title="First Page" children={
            <IconButton
              {...targets(`whiteboard first page`, `go to the first page of shared document`)}
              onClick={() => setPage(1)}
            >
              <KeyboardDoubleArrowLeftIcon fontSize="small" />
            </IconButton>
          } />
          <Tooltip title="Previous Page" children={
            <IconButton
              {...targets(`whiteboard previous page`, `go to the previous page of shared document`)}
              onClick={() => setPage(false)}
            >
              <KeyboardArrowLeftIcon fontSize="small" />
            </IconButton>
          } />
          <TextField
            {...targets(`whiteboard page number`, ``, `change to a specific page of the currently shared document`)}
            variant="standard"
            value={pageNumber}
            onChange={e => setPage(parseInt(e.target.value))}
            sx={{ width: '64px' }}
            slotProps={{
              input: {
                startAdornment: <MenuBookIcon sx={{ mr: 1 }} />,
              }
            }}
          />
          <Typography
            variant="h6"
            component="div"
            sx={{ ml: 2 }}
          >
            of {numPages}
          </Typography>
          <Tooltip title="Next Page" children={
            <IconButton
              {...targets(`whiteboard next page`, `go to the next page of shared document`)}
              onClick={() => setPage(true)}
            >
              <KeyboardArrowRightIcon fontSize="small" />
            </IconButton>
          } />
          <Tooltip title="Last Page" children={
            <IconButton
              {...targets(`whiteboard last page`, `go to the last page of shared document`)}
              onClick={() => setPage(numPages)}
            >
              <KeyboardDoubleArrowRightIcon fontSize="small" />
            </IconButton>
          } />
        </Grid>}

        {numPages > 0 && <Tooltip title="Close File">
          <IconButton
            {...targets(`exchange close file`, `close the currently shared file`)}
            color="error"
            onClick={() => {
              sendWhiteboardMessage(SocketActions.SHARE_FILE, { sharedFile: null });
            }}
          >
            <InsertPageBreak />
          </IconButton>
        </Tooltip>}

      </Grid>
    </Toolbar>
  </AppBar>
}

export default WhiteboardOptionsMenu;
