import React, { useContext } from 'react';

import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { dayjs } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function ScheduleDatePicker(_: IComponent): React.JSX.Element {

  const {
    setStartOfMonth,
    selectedDate,
    setSelectedDate,
    firstAvailable,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  return <DesktopDatePicker
    value={selectedDate}
    sx={{
      mt: '12px'
    }}
    slotProps={{
      textField: { fullWidth: true }
    }}
    onChange={(date: dayjs.Dayjs | null) => setSelectedDate(date ? date.isBefore(firstAvailable.time) ? firstAvailable.time : date : firstAvailable.time)}
    label="Date"
    format="MM/DD/YYYY"
    minDate={firstAvailable.time}
    onOpen={() => setSelectedDate(selectedDate.isAfter(firstAvailable.time) ? selectedDate : firstAvailable.time)}
    onMonthChange={date => date && setStartOfMonth(date.startOf('month'))}
    onYearChange={date => date && setStartOfMonth(date.startOf('month'))}
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
