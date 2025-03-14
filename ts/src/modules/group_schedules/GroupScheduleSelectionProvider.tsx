import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { dayjs, IQuote, TimeUnit, siteApi, dateFormat } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';
import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function GroupScheduleSelectionProvider({ children }: IComponent): React.JSX.Element {
  console.log('group schedule selection provider load');

  const { selectGroupSchedule: { item: groupSchedule } } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [startOfMonth, setStartOfMonth] = useState(dayjs().startOf(TimeUnit.MONTH));
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>();
  const [selectedTime, setSelectedTime] = useState<string>();

  const [quote, setQuote] = useState({} as IQuote);

  const [getScheduleDateSlots, { data: dateSlotsRequest, isFetching }] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleByDateQuery();

  const dateSlots = dateSlotsRequest?.groupScheduleDateSlots || [];

  const selectedSlots = useMemo(() => {
    if (!selectedDate) return [];
    const sdf = dateFormat(selectedDate);
    const ds = dateSlots.map(x =>
      x.startTime && sdf == x.startDate && dayjs(x.weekStart).add(dayjs.duration(x.startTime)).isAfter(dayjs()) && x
    ).filter(x => !!x);
    if (ds.length && ds[0].startTime) {
      setSelectedTime(ds[0].startTime);
    }
    return ds;
  }, [dateSlots, selectedDate]);

  const getDateSlots = useCallback(() => {
    if (groupSchedule?.schedule?.id?.length && startOfMonth && !isFetching) {
      getScheduleDateSlots({
        groupScheduleId: groupSchedule?.schedule?.id,
        date: dateFormat(startOfMonth),
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
    if (selectedDate && selectedTime && selectedSlots.length) {
      const slot = selectedSlots.find(s => s.startTime == selectedTime);
      setQuote(!slot ? {} : {
        scheduleBracketSlotId: slot.scheduleBracketSlotId,
        startTime: slot.startTime,
        slotDate: dateFormat(selectedDate),
      });
    } else if (quote.scheduleBracketSlotId) {
      setQuote({});
    }
  }, [selectedDate, selectedTime, selectedSlots]);

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
