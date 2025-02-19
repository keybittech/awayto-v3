import { FunctionComponent, createContext } from 'react';

import { dayjs, IGroupScheduleDateSlots, IQuote } from 'awayto/hooks';

declare global {
  type GroupScheduleSelectionContextType = {
    quote: IQuote;
    setQuote(quote: IQuote): void;
    selectedDate: dayjs.Dayjs;
    setSelectedDate(date: dayjs.Dayjs | null): void;
    selectedTime: dayjs.Dayjs;
    setSelectedTime(time: dayjs.Dayjs | null): void;
    startOfMonth: dayjs.Dayjs;
    setStartOfMonth(start: dayjs.Dayjs): void;
    dateSlots: Required<IGroupScheduleDateSlots>[];
    getDateSlots: () => void;
    firstAvailable: { time: dayjs.Dayjs, scheduleBracketSlotId: string };
    bracketSlotDateDayDiff: number;
    GroupScheduleDateSelection?: FunctionComponent<IComponent>;
    GroupScheduleTimeSelection?: FunctionComponent<IComponent>;
  }
}

export const GroupScheduleSelectionContext = createContext<GroupScheduleSelectionContextType | null>(null);

export default GroupScheduleSelectionContext;
