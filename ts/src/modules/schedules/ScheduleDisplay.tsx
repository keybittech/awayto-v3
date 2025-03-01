import React, { CSSProperties, useCallback, useMemo, useState, useEffect, useRef } from 'react';
import { FixedSizeGrid } from 'react-window';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';

import { useSchedule, useTimeName, deepClone, getRelativeDuration, ISchedule, IScheduleBracket, IScheduleBracketSlot, useStyles, useUtil, plural, targets } from 'awayto/hooks';

type GridCell = {
  columnIndex: number, rowIndex: number, style: CSSProperties
}

export interface ScheduleDisplayProps extends IComponent {
  schedule: ISchedule;
  setSchedule?(value: ISchedule): void;
  isKiosk?: boolean;
};

const bracketColors = ['cadetblue', 'forestgreen', 'brown', 'chocolate', 'darkslateblue', 'goldenrod', 'indianred', 'teal'];

export default function ScheduleDisplay({ isKiosk, schedule, setSchedule }: ScheduleDisplayProps): React.JSX.Element {

  const scheduleDisplay = useMemo(() => deepClone(schedule) as Required<ISchedule>, [schedule]);

  const classes = useStyles();

  const { openConfirm } = useUtil();

  const parentRef = useRef<HTMLDivElement>(null);
  const [selected, setSelected] = useState({} as Record<string, IScheduleBracketSlot>);
  const [selectedBracket, setSelectedBracket] = useState<Required<IScheduleBracket>>();
  const [buttonDown, setButtonDown] = useState(false);
  const [parentBox, setParentBox] = useState([0, 0]); // old can probably remove

  const displayOnly = isKiosk || !setSchedule;

  const scheduleTimeUnitName = useTimeName(scheduleDisplay.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(scheduleDisplay.bracketTimeUnitId);
  const slotTimeUnitName = useTimeName(scheduleDisplay.slotTimeUnitId);

  const {
    columns,
    rows,
    durations,
    xAxisTypeName
  } = useSchedule({ scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration: scheduleDisplay.slotDuration });

  const cellHeight = 30;
  const currentWidth = Math.max(60, parentBox[0] / (columns + 1));

  const scheduleBracketsValues = useMemo(() => Object.values(scheduleDisplay.brackets || {}) as Required<IScheduleBracket>[], [scheduleDisplay.brackets]);

  const setValue = useCallback((startTime: string) => {
    if (selectedBracket) {
      const bracket = scheduleDisplay.brackets[selectedBracket.id];
      if (bracket) {
        if (!bracket.slots) bracket.slots = {};

        const target = `schedule_bracket_slot_selection_${startTime}`;
        const exists = selected[target];

        const slot = {
          id: (new Date()).getTime().toString(),
          startTime,
          scheduleBracketId: selectedBracket.id
        } as IScheduleBracketSlot;

        if (exists?.id) {
          if (exists.scheduleBracketId !== bracket.id) return;
          delete bracket.slots[exists.id];
          delete selected[target];
        } else if (slot.id && Object.keys(bracket.slots).length * scheduleDisplay.slotDuration < getRelativeDuration(selectedBracket.duration, scheduleDisplay.bracketTimeUnitName, scheduleDisplay.slotTimeUnitName)) {
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
  }, [schedule, selectedBracket, scheduleBracketsValues, selected]);

  const Cell = useCallback((gridCell: GridCell) => {
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
    const bracketColor = exists ? bracketColors[scheduleBracketsValues.findIndex(b => b.id === exists.scheduleBracketId)] : '#eee';

    return <Tooltip key={`grid_cell_tooltip_${gridCell.columnIndex}_${gridCell.rowIndex}`} title={completeContextFormat}>
      <Box
        style={gridCell.style}
        sx={{
          userSelect: 'none',
          cursor: 'pointer',
          backgroundColor: exists ? `color-mix(in srgb, ${bracketColor} 90%, transparent)` : 'white',
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
  }, [durations, selected, scheduleBracketsValues, buttonDown, selectedBracket, slotTimeUnitName, xAxisTypeName, displayOnly, currentWidth]);

  useEffect(() => {
    const resizeObserver = new ResizeObserver(([event]) => {
      setParentBox([event.contentRect.width, event.contentRect.height]);
    });

    if (parentRef && parentRef.current) {
      resizeObserver.observe(parentRef.current);
    }
  }, []);

  useEffect(() => {
    if (!Object.keys(selected).length && scheduleBracketsValues.some(b => b.slots && Object.keys(b.slots).length)) {
      const newSelected = {} as Record<string, IScheduleBracketSlot>;
      scheduleBracketsValues.forEach(b => {
        b.slots && Object.values(b.slots).forEach(s => {
          newSelected[`schedule_bracket_slot_selection_${s.startTime}`] = s;
        });
      });
      setSelected(newSelected);
    }
  }, [selected, scheduleBracketsValues]);

  const RenderedGrid = useCallback(() => {
    return (isNaN(rows) || isNaN(columns)) ? <></> : <>
      <FixedSizeGrid
        style={{ position: 'absolute', top: 0, left: 0, backgroundColor: '#666' }}
        rowCount={rows + 1}
        columnCount={columns + 1}
        rowHeight={cellHeight}
        columnWidth={currentWidth}
        height={parentBox[1]}
        width={parentBox[0]}
      >
        {Cell}
      </FixedSizeGrid>
    </>
  }, [rows, columns, parentBox[0], parentBox[1]]);

  return <>
    {!displayOnly && <>
      <Box p={2} component="fieldset" sx={classes.legendBox}>
        <legend>Step 1. Select a Bracket</legend>
        <Typography variant="body1">Brackets are blocks of time that can be applied to the schedule. You can add multiple brackets, in case certain services only occur at certain times. You can click the X to remove a bracket.</Typography>
        <Grid sx={{ display: 'flex', alignItems: 'flex-end', flexWrap: 'wrap' }}>

          <Grid display='flex' flexDirection='row'>
            {scheduleBracketsValues.map((bracket, i) => {
              if (!bracket.slots) bracket.slots = {};
              return <Box key={`bracket-chip${i + 1}new`} m={1}>
                <Chip
                  {...targets(`schedule display bracket selection ${i}`, `${getRelativeDuration(bracket.duration, bracketTimeUnitName, slotTimeUnitName) - (Object.keys(bracket.slots).length * scheduleDisplay.slotDuration)} ${slotTimeUnitName}s for ${Object.values(bracket.services).map(s => s.name).join(', ')}`, `interact with or delete a bracket from the schedule`)}
                  sx={{
                    '&:hover': {
                      backgroundColor: 'rgba(48, 64, 80, 0.4)',
                      cursor: 'pointer'
                    },
                    borderWidth: '1px',
                    borderStyle: 'solid',
                    borderColor: bracketColors[i],
                    backgroundColor: selectedBracket?.id === bracket.id ? 'rgba(48, 64, 80, 0.25)' : undefined,
                    boxShadow: selectedBracket?.id === bracket.id ? 2 : undefined
                  }}
                  onDelete={() => {
                    openConfirm({
                      isConfirming: true,
                      confirmEffect: 'Delete a schedule bracket.',
                      confirmAction: () => {
                        const newSelected = Object.keys(selected).reduce((m, d) => {
                          if (selected[d].scheduleBracketId == bracket?.id) return m;
                          return {
                            ...m,
                            [d]: selected[d]
                          }
                        }, {});
                        setSelected({ ...newSelected });
                        delete scheduleDisplay.brackets[bracket.id];
                        setSchedule && setSchedule({ ...schedule, brackets: { ...scheduleDisplay.brackets } });
                      }
                    });
                  }}
                  onClick={() => {
                    setSelectedBracket(bracket);
                  }}
                />
              </Box>
            })}
          </Grid>
        </Grid>
      </Box>
    </>}

    <Typography pb={1} variant="body2">This schedule represents 1 {scheduleTimeUnitName} of {bracketTimeUnitName}s where every {plural(scheduleDisplay.slotDuration, slotTimeUnitName, slotTimeUnitName + 's')} is schedulable.</Typography>
    {displayOnly ? <>
      <Box ref={parentRef} sx={{ position: 'relative', overflow: 'scroll', width: '100%', height: `${cellHeight * Math.min(8, rows + 1)}px` }}>
        <RenderedGrid />
      </Box>
    </> : <>
      <Box component="fieldset" p={2} sx={classes.legendBox}>
        <legend>Step 2. Design your Schedule</legend>
        <Typography variant="body1">Select a time by clicking on it. Press and hold to select multiple times. Selections will remove time from the bracket.</Typography>
        <Box onMouseLeave={() => setButtonDown(false)}>
          <RenderedGrid />
        </Box>
      </Box>
    </>}
  </>;
}
