import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { dayjs, IQuote, TimeUnit, siteApi } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';
import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function GroupScheduleSelectionProvider({ children }: IComponent): React.JSX.Element {

  const { selectGroupSchedule: { item: groupSchedule } } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [startOfMonth, setStartOfMonth] = useState(dayjs().startOf(TimeUnit.MONTH));
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs>();
  const [selectedTime, setSelectedTime] = useState<string>();

  const [quote, setQuote] = useState({} as IQuote);

  const [getScheduleDateSlots, { data: dateSlotsRequest, isFetching }] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleByDateQuery();

  const selectedDateFormat = selectedDate?.format('YYYY-MM-DD');

  const dateSlots = dateSlotsRequest?.groupScheduleDateSlots || [];

  if (dateSlots.length && !selectedDate) {
    setSelectedDate(dayjs(dateSlots[0].startDate));
    setSelectedTime(dateSlots[0].startTime);
  }

  const selectedSlots = useMemo(() => {
    const ds = dateSlots.map(x => {
      if (x.startTime && selectedDateFormat == x.startDate && dayjs(x.weekStart).add(dayjs.duration(x.startTime)).isAfter(dayjs())) {
        return x
      }
    }).filter(x => !!x);
    if (ds.length) {
      setSelectedTime(ds[0].startTime);
    }
    return ds;
  }, [dateSlots, selectedDateFormat]);

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
