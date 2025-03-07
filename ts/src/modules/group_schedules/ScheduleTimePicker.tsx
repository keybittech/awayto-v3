import React, { useContext, useMemo } from 'react';

import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Alert from '@mui/material/Alert';

import { targets, bookingDTHours } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function ScheduleTimePicker(_: IComponent): React.JSX.Element {

  const {
    selectedDate,
    selectedTime,
    setSelectedTime,
    selectedSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const selections = useMemo(() => selectedSlots?.map((ds, i) => {
    return ds.startTime && ds.startDate && <MenuItem key={`date-slot-selection-key-${i}`} value={ds.startTime}>
      {bookingDTHours(ds.startDate, ds.startTime)}
    </MenuItem>
  }).filter(x => !!x), [selectedSlots]);

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
