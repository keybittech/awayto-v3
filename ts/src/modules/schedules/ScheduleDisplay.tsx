import React, { CSSProperties, useMemo, useState, useEffect, useRef, useCallback, ComponentType } from 'react';
import { FixedSizeGrid, GridChildComponentProps } from 'react-window';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';

import { useSchedule, useTimeName, deepClone, getRelativeDuration, ISchedule, IScheduleBracket, IScheduleBracketSlot, useStyles, useUtil, plural, targets } from 'awayto/hooks';

type GridCell = {
  columnIndex: number, rowIndex: number, style: CSSProperties
}

export interface ScheduleDisplayProps extends IComponent {
  schedule: ISchedule;
  setSchedule?(value: ISchedule): void;
  isKiosk?: boolean;
};

export default function ScheduleDisplay({ isKiosk, schedule, setSchedule }: ScheduleDisplayProps): React.JSX.Element {

  const scheduleDisplay = useMemo(() => deepClone(schedule), [schedule]);

  const classes = useStyles();

  const { openConfirm } = useUtil();

  const parentRef = useRef<HTMLDivElement>(null);
  const [selected, setSelected] = useState({} as Record<string, IScheduleBracketSlot>);
  const [selectedBracket, setSelectedBracket] = useState<IScheduleBracket>();
  const [buttonDown, setButtonDown] = useState(false);
  const [parentBox, setParentBox] = useState([0, 0]);

  const displayOnly = isKiosk || !setSchedule;

  const scheduleTimeUnitName = useTimeName(scheduleDisplay.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(scheduleDisplay.bracketTimeUnitId);
  const slotTimeUnitName = useTimeName(scheduleDisplay.slotTimeUnitId);

  const {
    columns,
    rows,
    durations,
  } = useSchedule({ scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration: scheduleDisplay.slotDuration });

  const cellHeight = 30;
  const currentWidth = !columns ? 30 : Math.max(60, parentBox[0] / (columns + 1));

  const scheduleBracketsValues = useMemo(() => Object.values(scheduleDisplay.brackets || {}) as IScheduleBracket[], [scheduleDisplay.brackets]);

  const setValue = (startTime: string) => {
    if (!scheduleDisplay.brackets || !scheduleDisplay.slotDuration || !selectedBracket?.id || !selectedBracket.duration) {
      return
    }

    const bracket = scheduleDisplay.brackets[selectedBracket.id];
    if (bracket) {
      if (!bracket.slots) bracket.slots = {};

      const target = `schedule_bracket_slot_selection_${startTime}`;
      const exists = selected[target];

      const slot = {
        id: (new Date()).getTime().toString(),
        startTime,
        scheduleBracketId: selectedBracket.id,
        color: selectedBracket.color,
      } as IScheduleBracketSlot;

      if (exists?.id) {
        if (exists.scheduleBracketId !== bracket.id) return;
        delete bracket.slots[exists.id];
        delete selected[target];
      } else if (slot.id && Object.keys(bracket.slots).length * scheduleDisplay.slotDuration < getRelativeDuration(selectedBracket.duration, bracketTimeUnitName, slotTimeUnitName)) {
        bracket.slots[slot.id] = slot;
        selected[target] = slot;
      } else {
        // alert('you went over your allottment');
        setButtonDown(false);
        return;
      }

      setSchedule && setSchedule({ ...schedule, brackets: { ...scheduleDisplay.brackets } });
      setSelected({ ...selected });
    }
  }

  const Cell = useCallback((gridCell: GridCell) => {
    if (!durations) return <></>;

    if (gridCell.columnIndex == 0 && gridCell.rowIndex == 0) {
      return <></>;
    }

    if (gridCell.columnIndex == 0 || gridCell.rowIndex == 0) {
      return <Box
        style={gridCell.style}
        sx={{
          userSelect: 'none',
          cursor: 'pointer',
          textAlign: 'center',
          color: 'white',
          whiteSpace: 'nowrap',
          fontWeight: 700
        }}
      >
        {durations[gridCell.columnIndex][gridCell.rowIndex].contextFormat}
      </Box>
    }

    const { startTime, contextFormat, completeContextFormat } = durations[gridCell.columnIndex][gridCell.rowIndex];

    const target = `schedule_bracket_slot_selection_${startTime}`;
    const exists = selected[target];

    return <Tooltip key={`grid_cell_tooltip_${gridCell.columnIndex}_${gridCell.rowIndex}`} title={completeContextFormat}>
      <Box
        style={gridCell.style}
        sx={{
          userSelect: 'none',
          cursor: 'pointer',
          backgroundColor: exists ? `color-mix(in srgb, ${exists.color} 90%, transparent)` : 'white',
          textAlign: 'center',
          position: 'relative',
          '&:hover': {
            backgroundColor: '#bbb',
            color: '#222',
            opacity: '1',
            boxShadow: '2',
            // borderColor: '#bbb'
          },
          // border: exists ? `1px solid ${bracketColor}` : undefined,
          color: !exists ? '#666' : '#000',
          boxShadow: exists ? '2' : undefined,
          whiteSpace: 'nowrap',
          borderTop: '1px solid #aaa',
          borderRight: '1px solid #aaa',
        }}
        onMouseLeave={() => !displayOnly && buttonDown && setValue(startTime)}
        onMouseDown={() => !displayOnly && setButtonDown(true)}
        onMouseUp={() => {
          if (!displayOnly) {
            setButtonDown(false);
            setValue(startTime);
          }
        }}
      >
        <>{'day' !== slotTimeUnitName && currentWidth >= 120 ? completeContextFormat : contextFormat}</>
      </Box>
    </Tooltip>
  }, [durations, selected, scheduleBracketsValues, displayOnly, buttonDown, slotTimeUnitName, currentWidth]);

  const RenderedGrid = useCallback(({ children }: { children: ComponentType<GridChildComponentProps<any>> }) => !rows || !columns ? <></> : <FixedSizeGrid
    style={{ position: 'absolute', top: 0, left: 0, backgroundColor: '#666' }}
    rowCount={rows + 1}
    columnCount={columns + 1}
    rowHeight={cellHeight}
    columnWidth={currentWidth}
    height={parentBox[1]}
    width={parentBox[0]}
  >
    {children}
  </FixedSizeGrid>, [rows, columns, cellHeight, currentWidth, parentBox]);

  const RepresentedTime = () => <Typography pb={1} variant="body2">This schedule represents 1 {scheduleTimeUnitName} of {bracketTimeUnitName}s where every {plural(scheduleDisplay.slotDuration, slotTimeUnitName, slotTimeUnitName + 's')} is schedulable.</Typography>;

  useEffect(() => {
    const resizeObserver = new ResizeObserver(([event]) => {
      setParentBox([event.contentRect.width, event.contentRect.height]);
    });

    if (parentRef && parentRef.current) {
      resizeObserver.observe(parentRef.current);
    }
  }, []);

  useEffect(() => {
    if (scheduleBracketsValues.some(b => b.slots && Object.keys(b.slots).length)) {
      const newSelected = {} as Record<string, IScheduleBracketSlot>;
      scheduleBracketsValues.forEach(b => {
        b.slots && Object.values(b.slots).forEach(s => {
          newSelected[`schedule_bracket_slot_selection_${s.startTime}`] = {
            ...s,
            color: b.color,
          };
        });
      });
      setSelected(newSelected);
    }
  }, [scheduleBracketsValues]);

  const parentRefStyle = { position: 'relative', overflow: 'scroll', width: '100%', height: `${cellHeight * Math.min(8, (rows || 0) + 1)}px` };

  return <>
    {!displayOnly && <Box mb={2}>
      <Typography variant="h2">{schedule.name} Brackets</Typography>
      <Box p={2} component="fieldset" sx={classes.legendBox}>
        <legend>Step 1. Select a Bracket</legend>
        <Grid container direction="column" spacing={2}>
          <Grid>
            <Typography variant="body1">To start, select a bracket by clicking on it. Brackets are blocks of time wherein services are offered.</Typography>
          </Grid>
          <Grid container size="grow">

            {scheduleBracketsValues.length && <ToggleButtonGroup
              value={selectedBracket?.id}
              exclusive
              onChange={(_, bracketId: string) => {
                if (!bracketId) return setSelectedBracket(undefined);
                const sb = scheduleBracketsValues.find(b => b.id == bracketId);
                if (sb) {
                  setSelectedBracket({ ...sb });
                }
              }}
            >
              {scheduleBracketsValues.map((b, i) => {
                const bracketDuration = getRelativeDuration(b.duration, bracketTimeUnitName, slotTimeUnitName);
                const usedSlots = Object.keys(b.slots || {}).length * (scheduleDisplay.slotDuration || 0);
                const remainingDuration = bracketDuration - usedSlots;
                return <ToggleButton
                  key={`bracket_service_toggle_${i}`}
                  {...targets(
                    `schedule display bracket selection ${i + 1}`,
                    `select bracket ${i + 1} to edit the schedule with`
                  )}
                  sx={{
                    bgcolor: `${b.color}`,
                    alignItems: 'start',
                    '&:hover': { bgcolor: `${b.color}bb` },
                    '&.Mui-selected': {
                      bgcolor: `${b.color}aa`,
                      '&:hover': { bgcolor: `${b.color}66` }
                    }
                  }}
                  value={b.id || ''}
                >
                  <Box sx={{ color: '#333' }} textAlign="left">
                    <Typography>Bracket {i + 1}</Typography>
                    <Typography variant="caption">{remainingDuration} {slotTimeUnitName}s Remaining</Typography><br />
                    <Typography
                      color="info"
                      sx={{
                        p: .5,
                        float: 'right',
                        fontSize: '8px',
                        display: selectedBracket?.id == b.id ? 'inline' : 'none',
                        bgcolor: '#333'
                      }}
                    >Selected</Typography>
                  </Box>
                </ToggleButton>
              })}
            </ToggleButtonGroup>}
          </Grid>

          {selectedBracket && <Grid container size="grow">
            <Grid size="grow">
              <Typography><strong>Selected Bracket Services:</strong> {Object.values(selectedBracket.services || {}).map(s => s.name).join(', ')}</Typography>
            </Grid>
            <Button
              variant="text"
              color="error"
              onClick={() => {
                openConfirm({
                  isConfirming: true,
                  confirmEffect: 'Delete a schedule bracket.',
                  confirmAction: () => {
                    if (!selectedBracket.id || !scheduleDisplay.brackets) return;
                    const newSelected = Object.keys(selected).reduce((m, d) => {
                      if (selected[d].scheduleBracketId == selectedBracket.id) return m;
                      return {
                        ...m,
                        [d]: selected[d]
                      }
                    }, {});
                    setSelected({ ...newSelected });
                    delete scheduleDisplay.brackets[selectedBracket.id];
                    setSchedule && setSchedule({ ...schedule, brackets: { ...scheduleDisplay.brackets } });
                    setSelectedBracket(undefined);
                  }
                });
              }}
            >Delete Bracket</Button>
          </Grid>}
        </Grid>
      </Box>
    </Box>}


    {displayOnly ? <>
      <RepresentedTime />
      <Box
        ref={parentRef}
        sx={parentRefStyle}
      >
        <RenderedGrid>
          {Cell}
        </RenderedGrid>
      </Box>
    </> : <>
      <Box component="fieldset" p={2} sx={classes.legendBox}>
        <legend>Step 2. Design your Schedule</legend>
        <Box mb={2}>
          <Typography variant="body1">Select times to apply to the selected bracket. Press and hold to select multiple times.</Typography>
        </Box>
        <Box
          ref={parentRef}
          sx={parentRefStyle}
          onMouseLeave={() => setButtonDown(false)}
        >
          <RenderedGrid>
            {Cell}
          </RenderedGrid>
        </Box>
        <RepresentedTime />
      </Box>
    </>}
  </>;
}
