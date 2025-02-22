import { FunctionComponent, createContext } from 'react';

import { dayjs, IGroupScheduleDateSlots, IQuote } from 'awayto/hooks';

export interface GroupScheduleSelectionContextType {
  quote: IQuote;
  setQuote(quote: IQuote): void;
  selectedDate: dayjs.Dayjs;
  setSelectedDate(date: dayjs.Dayjs): void;
  selectedTime: dayjs.Dayjs;
  setSelectedTime(time: dayjs.Dayjs): void;
  startOfMonth: dayjs.Dayjs;
  setStartOfMonth(start: dayjs.Dayjs): void;
  dateSlots?: IGroupScheduleDateSlots[];
  getDateSlots: () => void;
  firstAvailable: { time: dayjs.Dayjs, scheduleBracketSlotId: string };
  bracketSlotDateDayDiff: number;
  GroupScheduleDateSelection?: FunctionComponent<IComponent>;
  GroupScheduleTimeSelection?: FunctionComponent<IComponent>;
}

export const GroupScheduleSelectionContext = createContext<GroupScheduleSelectionContextType | null>(null);

export default GroupScheduleSelectionContext;
