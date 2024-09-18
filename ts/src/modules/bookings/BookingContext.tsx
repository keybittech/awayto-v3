import { IBooking } from 'awayto/hooks';
import { createContext } from 'react';

declare global {
  type BookingContextType = {
    bookingValues: IBooking[];
    setBookingValuesChanged: (prop: boolean) => void;
    bookingValuesChanged: boolean;
    selectedBooking: IBooking[];
    setSelectedBooking: (quotes: IBooking[]) => void;
    handleSelectPendingBooking: (prop: IBooking) => void;
    handleSelectPendingBookingAll: () => void;
  }
}

export const BookingContext = createContext<BookingContextType | null>(null);

export default BookingContext;
