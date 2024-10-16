import React, { useContext } from 'react';

import TextField from '@mui/material/TextField';
import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { dayjs } from 'awayto/hooks';

import GroupScheduleSelectionContext from './GroupScheduleSelectionContext';

export function ScheduleDatePicker(): React.JSX.Element {

  const {
    setStartOfMonth,
    selectedDate,
    setSelectedDate,
    firstAvailable,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  return <DesktopDatePicker
    value={selectedDate}
    onChange={(date: dayjs.Dayjs | null) => setSelectedDate(date ? date.isBefore(firstAvailable.time) ? firstAvailable.time : date : null)}
    label="Date"
    inputFormat="MM/DD/YYYY"
    minDate={firstAvailable.time}
    onOpen={() => setSelectedDate(selectedDate.isAfter(firstAvailable.time) ? selectedDate : firstAvailable.time)}
    onMonthChange={date => date && setStartOfMonth(date.startOf('month'))}
    onYearChange={date => date && setStartOfMonth(date.startOf('month'))}
    renderInput={(params) => <TextField fullWidth {...params} />}
    disableHighlightToday={true}
    shouldDisableDate={date => {
      if (date && dateSlots?.length) {
        return !dateSlots.filter(s => s.startDate === date.format("YYYY-MM-DD")).length;
      }
      return true;
    }}
  />
}

export default ScheduleDatePicker;
