import { IBooking } from 'awayto/hooks';
import { createContext } from 'react';

export interface BookingContextType {
  bookingValues: IBooking[];
  setBookingValuesChanged: (prop: boolean) => void;
  bookingValuesChanged: boolean;
  selectedBooking: IBooking[];
  setSelectedBooking: (quotes: IBooking[]) => void;
  handleSelectPendingBooking: (prop: IBooking) => void;
  handleSelectPendingBookingAll: () => void;
}

export const BookingContext = createContext<BookingContextType | null>(null);

export default BookingContext;
