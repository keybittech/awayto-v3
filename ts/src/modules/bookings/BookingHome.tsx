import React, { useContext } from 'react';

import BookingContext from './BookingContext';

export function BookingHome(): React.JSX.Element {

  const { bookingValues: upcomingBookings } = useContext(BookingContext) as BookingContextType;

  console.log({ upcomingBookings });

  return <></>;
}

export default BookingHome;
