import React, { useContext } from 'react';

import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { dayjs, targets } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function ScheduleDatePicker(_: IComponent): React.JSX.Element {

  const {
    setStartOfMonth,
    selectedDate,
    setSelectedDate,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  return <DesktopDatePicker
    {...targets(`date picker selection`, `Date`, `select a date`)}
    value={selectedDate || null}
    format="MM/DD/YYYY"
    onMonthChange={date => date && setStartOfMonth(date.startOf('month'))}
    onChange={(date: dayjs.Dayjs | null) => setSelectedDate(date ? date : undefined)}
    onYearChange={date => date && setStartOfMonth(date.startOf('month'))}
    disableHighlightToday={true}
    slotProps={{
      textField: { fullWidth: true, required: true }
    }}
    shouldDisableDate={date => {
      if (date && dateSlots?.length) {
        return !dateSlots.filter(s => s.startDate === date.format("YYYY-MM-DD")).length;
      }
      return true;
    }}
  />
}

export default ScheduleDatePicker;
