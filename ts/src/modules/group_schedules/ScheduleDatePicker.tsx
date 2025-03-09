import React, { useContext, useCallback } from 'react';

import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { dayjs, targets, useTimeName } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';
import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';

export function ScheduleDatePicker(_: IComponent): React.JSX.Element {

  const {
    selectGroupSchedule: {
      item: groupSchedule,
    },
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const {
    setStartOfMonth,
    selectedDate,
    setSelectedDate,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const scheduleTimeUnitName = useTimeName(groupSchedule?.schedule?.scheduleTimeUnitId);

  const slotDurations = dateSlots?.map(x => x.startTime || '');
  const scheduleStartTime = dayjs(groupSchedule?.schedule?.startTime).startOf('week');

  const monthComparator = useCallback((date: dayjs.Dayjs) => {
    if (!slotDurations) return true; // true means to disable the date when comparing
    const positionInCycle = date.tz(groupSchedule?.schedule?.timezone).diff(scheduleStartTime, 'day') % 28;
    return !slotDurations.includes(positionInCycle ? `P${positionInCycle}D` : 'PT0S');
  }, [groupSchedule, scheduleStartTime, slotDurations]);

  const weekComparator = useCallback((date: dayjs.Dayjs) => {
    if (!dateSlots) return true;
    const dateFormat = date.format("YYYY-MM-DD");
    return !dateSlots.find(ds => ds.startDate == dateFormat);
  }, [dateSlots]);

  const comparator = 'week' == scheduleTimeUnitName ? weekComparator : monthComparator;

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
    shouldDisableDate={comparator}
    slotProps={{
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
