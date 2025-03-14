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

  const alertStyle = {
    width: '100%',
    height: '100%',
    alignItems: 'center',
  };

  if (!selectedDate || !selectedTime) {
    return <Alert sx={alertStyle} variant="outlined" color="info" severity="info">Select a date.</Alert>;
  }

  if (!selectedSlots?.length) {
    return <Alert sx={alertStyle} variant="outlined" color="warning" severity="warning">No available times.</Alert>;
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
