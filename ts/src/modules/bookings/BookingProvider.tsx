import { useMemo, useState } from "react";

import { siteApi, IBooking } from 'awayto/hooks';

import BookingContext from "./BookingContext";

export function BookingProvider({ children }: IComponent): React.JSX.Element {

  const [bookingValuesChanged, setBookingValuesChanged] = useState(false);
  const [selectedBooking, setSelectedBooking] = useState<IBooking[]>([]);

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const bookingValues = useMemo(() => Object.values(profileRequest?.userProfile?.bookings || {}), [profileRequest?.userProfile]);

  const bookingContext = {
    bookingValues,
    setBookingValuesChanged,
    bookingValuesChanged,
    selectedBooking,
    setSelectedBooking,
    handleSelectPendingBooking(booking) {
      const currentIndex = selectedBooking.indexOf(booking);
      const newChecked = [...selectedBooking];

      if (currentIndex === -1) {
        newChecked.push(booking);
      } else {
        newChecked.splice(currentIndex, 1);
      }

      setSelectedBooking(newChecked);
    },
    handleSelectPendingBookingAll() {
      const bookingValuesSet = selectedBooking.length === bookingValues.length ?
        selectedBooking.filter(v => !bookingValues.some(bv => bv.id == v.id)) :
        [...selectedBooking, ...bookingValues.filter(v => !selectedBooking.includes(v))];

      setSelectedBooking(bookingValuesSet);
    }
  } as BookingContextType | null;

  return useMemo(() => <BookingContext.Provider value={bookingContext}>
    {children}
  </BookingContext.Provider>, [bookingContext]);

}

export default BookingProvider;
