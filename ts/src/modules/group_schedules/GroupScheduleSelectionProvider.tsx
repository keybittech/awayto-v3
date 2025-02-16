import React, { useContext, useEffect, useMemo, useState } from 'react';

import { dayjs, IGroupScheduleDateSlots, IQuote, TimeUnit, quotedDT, userTimezone, siteApi } from 'awayto/hooks';

import GroupScheduleContext from './GroupScheduleContext';
import GroupScheduleSelectionContext from './GroupScheduleSelectionContext';

export function GroupScheduleSelectionProvider({ children }: IProps): React.JSX.Element {

  const { selectGroupSchedule: { item: groupSchedule } } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [firstAvailable, setFirstAvailable] = useState({ time: dayjs().startOf('day'), scheduleBracketSlotId: '' });
  const [startOfMonth, setStartOfMonth] = useState(dayjs().startOf(TimeUnit.MONTH));
  const [selectedDate, setSelectedDate] = useState<dayjs.Dayjs | null>(firstAvailable.time!);
  const [selectedTime, setSelectedTime] = useState<dayjs.Dayjs | null>(firstAvailable.time!);

  const [quote, setQuote] = useState({} as IQuote);

  const tz = useMemo(() => encodeURIComponent(btoa(userTimezone)), [userTimezone]);

  const { data: dateSlots } = siteApi.useGroupScheduleServiceGetGroupScheduleByDateQuery({
    groupScheduleId: groupSchedule?.schedule?.id || '',
    date: startOfMonth.format("YYYY-MM-DD"),
    timezone: tz,
  }, { skip: !groupSchedule?.schedule?.id });

  if (dateSlots?.groupScheduleDateSlots?.length && !firstAvailable.scheduleBracketSlotId) {
    const [slot] = dateSlots.groupScheduleDateSlots as Required<IGroupScheduleDateSlots>[];
    const time = quotedDT(slot.weekStart, slot.startTime);
    const firstAvail = { ...slot, time };
    setFirstAvailable(firstAvail);
    setSelectedDate(time);
    setSelectedTime(time);
  }

  const bracketSlotDateDayDiff = useMemo(() => {
    if (selectedDate) {
      const startOfDay = selectedDate.startOf('day');
      return startOfDay.diff(startOfDay.day(0), TimeUnit.DAY);
    }
    return 0;
  }, [selectedDate]);

  useEffect(() => {
    const date = selectedDate?.format('YYYY-MM-DD');
    const timeHour = selectedTime?.hour() || 0;
    const timeMins = selectedTime?.minute() || 0;
    const duration = dayjs.duration(0)
      .add(bracketSlotDateDayDiff, TimeUnit.DAY)
      .add(timeHour, TimeUnit.HOUR)
      .add(timeMins, TimeUnit.MINUTE);
    const [slot] = dateSlots?.groupScheduleDateSlots?.filter(s => {
      const startTimeDuration = dayjs.duration(s.startTime!);
      return s.startDate === date && duration.hours() === startTimeDuration.hours() && duration.minutes() === startTimeDuration.minutes();
    }) || [] as Required<IGroupScheduleDateSlots>[];
    if (slot) {
      setQuote({ slotDate: date, scheduleBracketSlotId: slot.scheduleBracketSlotId, startTime: slot.startTime } as IQuote);
    }
  }, [selectedDate, selectedTime]);

  const groupScheduleSelectionContext = {
    quote,
    setQuote,
    selectedDate,
    setSelectedDate,
    selectedTime,
    setSelectedTime,
    startOfMonth,
    setStartOfMonth,
    dateSlots: dateSlots?.groupScheduleDateSlots,
    firstAvailable,
    bracketSlotDateDayDiff,
  } as GroupScheduleSelectionContextType;

  return useMemo(() => <GroupScheduleSelectionContext.Provider value={groupScheduleSelectionContext}>
    {children}
  </GroupScheduleSelectionContext.Provider>, [groupScheduleSelectionContext]);
}

export default GroupScheduleSelectionProvider;
