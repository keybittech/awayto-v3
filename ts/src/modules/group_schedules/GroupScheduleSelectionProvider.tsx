import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { dayjs, IQuote, TimeUnit, siteApi, useTimeName } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';
import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function GroupScheduleSelectionProvider({ children }: IComponent): React.JSX.Element {

  const { selectGroupSchedule: { item: groupSchedule } } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [startOfMonth, setStartOfMonth] = useState(dayjs().startOf(TimeUnit.MONTH));
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>();
  const [selectedTime, setSelectedTime] = useState<string>();

  const [quote, setQuote] = useState({} as IQuote);

  const [getScheduleDateSlots, { data: dateSlotsRequest, isFetching }] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleByDateQuery();

  const scheduleTimeUnitName = useTimeName(groupSchedule?.schedule?.scheduleTimeUnitId);
  const scheduleStartTime = dayjs(groupSchedule?.schedule?.startTime).startOf('week');
  const selectedDateFormat = selectedDate?.format('YYYY-MM-DD');

  const dateSlots = dateSlotsRequest?.groupScheduleDateSlots || [];

  if (dateSlots.length && !selectedDate) {
    setSelectedDate(dayjs(dateSlots[0].startDate));
    setSelectedTime(dateSlots[0].startTime);
  }

  const selectedSlots = useMemo(() => {
    let ds = [];
    let firstAvailable = '';
    if (selectedDate && 'month' == scheduleTimeUnitName) { // There should only be a single available time for monthly schedules, the full day
      const positionInCycle = selectedDate.tz(groupSchedule?.schedule?.timezone).diff(scheduleStartTime, 'day') % 28;
      firstAvailable = positionInCycle ? `P${positionInCycle}D` : 'PT0S';
      ds = dateSlots.filter(ds => ds.startTime == firstAvailable);
    } else { // Weekly schedules will have an array of time slots on whatever date is selected
      ds = dateSlots.map(x =>
        x.startTime && selectedDateFormat == x.startDate && dayjs(x.weekStart).add(dayjs.duration(x.startTime)).isAfter(dayjs()) && x
      ).filter(x => !!x);
      if (ds.length && ds[0].startTime) {
        firstAvailable = ds[0].startTime;
      }
    }
    if (firstAvailable.length) {
      setSelectedTime(firstAvailable);
    }
    return ds;
  }, [dateSlots, selectedDate, selectedDateFormat]);

  const getDateSlots = useCallback(() => {
    if (groupSchedule?.schedule?.id?.length && startOfMonth && !isFetching) {
      getScheduleDateSlots({
        groupScheduleId: groupSchedule?.schedule?.id,
        date: startOfMonth.format("YYYY-MM-DD"),
      });
    }
  }, [groupSchedule?.schedule, startOfMonth, isFetching]);

  useEffect(() => {
    const nowTime = dayjs().startOf('month');
    if (groupSchedule?.schedule?.id && (nowTime.isSame(startOfMonth) || nowTime.isBefore(startOfMonth))) {
      getDateSlots();
    }
  }, [startOfMonth, groupSchedule?.schedule]);

  useEffect(() => {
    if (selectedTime) {
      const slot = selectedSlots.find(s => s.startTime == selectedTime);
      setQuote(!slot ? {} : {
        scheduleBracketSlotId: slot.scheduleBracketSlotId,
        startTime: slot.startTime,
        slotDate: selectedDateFormat,
      });
    }
  }, [selectedDateFormat, selectedTime, selectedSlots]);

  const groupScheduleSelectionContext: GroupScheduleSelectionContextType = {
    quote,
    setQuote,
    selectedDate,
    setSelectedDate,
    selectedTime,
    setSelectedTime,
    startOfMonth,
    setStartOfMonth,
    dateSlots,
    getDateSlots,
    selectedSlots,
  };

  return useMemo(() => <GroupScheduleSelectionContext.Provider value={groupScheduleSelectionContext}>
    {children}
  </GroupScheduleSelectionContext.Provider>, [groupScheduleSelectionContext]);
}

export default GroupScheduleSelectionProvider;
