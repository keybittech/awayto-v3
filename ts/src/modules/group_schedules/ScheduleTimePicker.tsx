import React, { useContext, useMemo } from 'react';

import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Alert from '@mui/material/Alert';

import { targets, bookingDTHours, useTimeName } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';
import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';

export function ScheduleTimePicker(_: IComponent): React.JSX.Element {

  const {
    selectGroupSchedule: {
      item: groupSchedule,
    },
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const {
    selectedDate,
    selectedTime,
    setSelectedTime,
    selectedSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const scheduleTimeUnitName = useTimeName(groupSchedule?.schedule?.scheduleTimeUnitId);

  const selections = useMemo(() => selectedSlots?.map((ds, i) => {
    return <MenuItem key={`date-slot-selection-key-${i}`} value={ds.startTime}>
      {'week' == scheduleTimeUnitName && ds.startDate && ds.startTime ? bookingDTHours(ds.startDate, ds.startTime) : 'Full day'}
    </MenuItem>
  }).filter(x => !!x), [selectedSlots, scheduleTimeUnitName]);

  if (!selectedDate || !selectedTime) {
    return <Alert sx={{ width: '100%' }} variant="outlined" color="info" severity="info">Select a date. Times will appear here.</Alert>;
  }

  if (!selectedSlots?.length) {
    return <Alert sx={{ width: '100%' }} variant="outlined" color="warning" severity="warning">No available times remain.</Alert>;
  }

  return <TextField
    {...targets('select time picker select', 'Time', 'select a time')}
    select
    fullWidth
    required
    value={selectedTime}
    onChange={e => setSelectedTime(e.target.value)}
  >
    {selections}
  </TextField>;
}

export default ScheduleTimePicker;
