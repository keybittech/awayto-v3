import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { dayjs, IQuote, TimeUnit, siteApi } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';
import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function GroupScheduleSelectionProvider({ children }: IComponent): React.JSX.Element {

  const { selectGroupSchedule: { item: groupSchedule } } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [startOfMonth, setStartOfMonth] = useState(dayjs().startOf(TimeUnit.MONTH));
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>();
  const [selectedTime, setSelectedTime] = useState<dayjs.Dayjs>();

  const [quote, setQuote] = useState({} as IQuote);

  const [getScheduleDateSlots, { data: dateSlotsRequest, isFetching }] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleByDateQuery();

  const dateSlots = useMemo(() => {
    const ds = dateSlotsRequest?.groupScheduleDateSlots || [];
    return ds.map(x => {
      if (!x.startTime) return undefined;
      const duration = dayjs.duration(x.startTime);
      return {
        ...x,
        duration,
        hour: duration.hours(),
        minute: duration.minutes(),
      }
    }).filter(x => !!x); // remove undef
  }, [dateSlotsRequest]);

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
    if (!dateSlots.length || !selectedTime || !selectedDate) {
      setQuote({});
      return;
    }
    const selectedDateFormat = selectedDate.format('YYYY-MM-DD');
    const selectedTimeHours = selectedTime.hour();
    const selectedTimeMinutes = selectedTime.minute();
    const slot = dateSlots.find(s => s.startDate === selectedDateFormat && selectedTimeHours === s.hour && selectedTimeMinutes === s.minute);
    if (!slot) return;
    setQuote({
      scheduleBracketSlotId: slot.scheduleBracketSlotId,
      startTime: slot.startTime,
      slotDate: selectedDateFormat,
    });
  }, [selectedDate, selectedTime, dateSlots]);

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
  };

  return useMemo(() => <GroupScheduleSelectionContext.Provider value={groupScheduleSelectionContext}>
    {children}
  </GroupScheduleSelectionContext.Provider>, [groupScheduleSelectionContext]);
}

export default GroupScheduleSelectionProvider;
