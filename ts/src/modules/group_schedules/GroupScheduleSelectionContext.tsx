import { FunctionComponent, createContext } from 'react';

import { dayjs, IGroupScheduleDateSlots, IQuote } from 'awayto/hooks';
import { Duration } from 'dayjs/plugin/duration';

type DateSlotWithDuration = IGroupScheduleDateSlots & {
  duration?: Duration;
}

export interface GroupScheduleSelectionContextType {
  quote: IQuote;
  setQuote(quote: IQuote): void;
  selectedDate?: dayjs.Dayjs;
  setSelectedDate(date: dayjs.Dayjs | undefined): void;
  selectedTime?: dayjs.Dayjs;
  setSelectedTime(time: dayjs.Dayjs | undefined): void;
  startOfMonth: dayjs.Dayjs;
  setStartOfMonth(start: dayjs.Dayjs): void;
  dateSlots?: DateSlotWithDuration[];
  getDateSlots: () => void;
  GroupScheduleDateSelection?: FunctionComponent<IComponent>;
  GroupScheduleTimeSelection?: FunctionComponent<IComponent>;
}

export const GroupScheduleSelectionContext = createContext<GroupScheduleSelectionContextType | null>(null);

export default GroupScheduleSelectionContext;
