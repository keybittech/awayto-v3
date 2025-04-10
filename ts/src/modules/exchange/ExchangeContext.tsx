import { createContext } from 'react';
import {
  BookingServiceGetBookingFilesApiArg,
  BookingServiceGetBookingFilesApiResponse,
  UseSiteQuery,
  SocketMessage
} from 'awayto/hooks';

export interface ExchangeContextType {
  exchangeId: string;
  topicMessages: SocketMessage[];
  setTopicMessages(selector: (prop: Partial<SocketMessage>[]) => SocketMessage[]): void;
  getBookingFiles: UseSiteQuery<BookingServiceGetBookingFilesApiArg, BookingServiceGetBookingFilesApiResponse>;
}

export const ExchangeContext = createContext<ExchangeContextType | null>(null);

export default ExchangeContext;
