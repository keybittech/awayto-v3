import React, { useContext, useState } from 'react';

import Typography from '@mui/material/Typography';
import { DataGrid } from '@mui/x-data-grid/DataGrid';

import { bookingFormat, useGrid } from 'awayto/hooks';

import BookingContext, { BookingContextType } from './BookingContext';

export function BookingHome(_: IComponent): React.JSX.Element {

  const { bookingValues: upcomingBookings } = useContext(BookingContext) as BookingContextType;
  const [selected, setSelected] = useState<string[]>([]);

  const bookingsGridProps = useGrid({
    rows: upcomingBookings,
    columns: [
      { flex: 1, headerName: 'Date', field: 'slotDate' },
      {
        flex: 1, headerName: 'Time', field: 'bookingTime',
        renderCell: ({ row }) => row.slotDate && row.scheduleBracketSlot?.startTime ? bookingFormat(row.slotDate, row.scheduleBracketSlot?.startTime) : ''
      },
    ],
    selected,
    onSelected: p => setSelected(p as string[]),
    toolbar: () => <>
      <Typography variant="button">Upcoming Appointments:</Typography>
      {/* !!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box> */}
    </>
  });
  return <DataGrid {...bookingsGridProps} />
}

export default BookingHome;
