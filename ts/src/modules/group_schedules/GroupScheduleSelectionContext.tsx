import { FunctionComponent, createContext } from 'react';

import { dayjs, IGroupScheduleDateSlots, IQuote } from 'awayto/hooks';

export interface GroupScheduleSelectionContextType {
  quote: IQuote;
  setQuote(quote: IQuote): void;
  selectedDate?: dayjs.Dayjs;
  setSelectedDate(date: dayjs.Dayjs | undefined): void;
  selectedTime?: string;
  setSelectedTime(startTime?: string): void;
  startOfMonth: dayjs.Dayjs;
  setStartOfMonth(start: dayjs.Dayjs): void;
  dateSlots?: IGroupScheduleDateSlots[];
  selectedSlots?: IGroupScheduleDateSlots[];
  getDateSlots: () => void;
  GroupScheduleDateSelection?: FunctionComponent<IComponent>;
  GroupScheduleTimeSelection?: FunctionComponent<IComponent>;
}

export const GroupScheduleSelectionContext = createContext<GroupScheduleSelectionContextType | null>(null);

export default GroupScheduleSelectionContext;
