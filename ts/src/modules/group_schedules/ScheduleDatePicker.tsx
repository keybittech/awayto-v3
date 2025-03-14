import React, { useContext } from 'react';

import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { dateFormat, dayjs, targets } from 'awayto/hooks';

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
    formatDensity="spacious"
    disablePast={true}
    onMonthChange={date => date && setStartOfMonth(date.startOf('month'))}
    onChange={(date: dayjs.Dayjs | null) => setSelectedDate(date ? date : undefined)}
    onYearChange={date => date && setStartOfMonth(date.startOf('month'))}
    disableHighlightToday={true}
    shouldDisableDate={(date: dayjs.Dayjs) => {
      if (!dateSlots) return true;
      const df = dateFormat(date);
      return !dateSlots.find(ds => ds.startDate == df);
    }}
    reduceAnimations
    slotProps={{
      nextIconButton: {
        autoFocus: true // when no slots are available, focus the next button to preserve accessibility
      },
      openPickerButton: {
        color: 'secondary'
      },
      clearIcon: { sx: { color: 'red' } },
      field: {
        clearable: true,
        onClear: () => setSelectedDate(undefined)
      },
      textField: { fullWidth: true, required: true }
    }}
  />
}

export default ScheduleDatePicker;
