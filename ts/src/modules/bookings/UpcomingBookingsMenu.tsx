import React, { useContext } from 'react';
import { useNavigate } from 'react-router-dom';

import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

import JoinFullIcon from '@mui/icons-material/JoinFull';

import { bookingFormat, targets } from 'awayto/hooks';

import BookingContext, { BookingContextType } from './BookingContext';

interface UpcomingBookingsMenuProps extends IComponent {
  upcomingBookingsAnchorEl: null | HTMLElement;
  upcomingBookingsMenuId: string;
  isUpcomingBookingsOpen: boolean;
  handleMenuClose: () => void;
}

export function UpcomingBookingsMenu({ handleMenuClose, upcomingBookingsAnchorEl, upcomingBookingsMenuId, isUpcomingBookingsOpen }: UpcomingBookingsMenuProps): React.JSX.Element {

  const nav = useNavigate();
  const navigate = (loc: string) => {
    nav(loc);
  }

  const { bookingValues: upcomingBookings } = useContext(BookingContext) as BookingContextType;

  return <Menu
    anchorEl={upcomingBookingsAnchorEl}
    anchorOrigin={{
      vertical: 'bottom',
      horizontal: 'right',
    }}
    id={upcomingBookingsMenuId}
    transformOrigin={{
      vertical: 'top',
      horizontal: 'right',
    }}
    open={!!upcomingBookings.length && isUpcomingBookingsOpen}
    onClose={handleMenuClose}
    MenuListProps={{
      'aria-labelledby': 'topbar-exchange-toggle'
    }}
  >
    <ListItem>Join an Exchange</ListItem>
    {upcomingBookings.map((booking, i) => {

      if (booking.slotDate && booking.scheduleBracketSlot?.startTime && booking.scheduleBracketSlot?.id && booking.service?.name && booking.serviceTier?.name) {

        // const dt = bookingDT(booking.slotDate, booking.scheduleBracketSlot.startTime);

        return <MenuItem
          autoFocus={i == 0}
          key={`upcoming_appt_ub_${i}`}
          {...targets(
            `join exchange ${booking.slotDate} ${booking.scheduleBracketSlot.startTime}`,
            `go to exchange for ${bookingFormat(booking.slotDate, booking.scheduleBracketSlot.startTime)}`
          )}
          onClick={() => {
            navigate(`/exchange/${booking.id}`);
          }}
        >
          <ListItemIcon><JoinFullIcon /></ListItemIcon>
          <ListItemText
            primary={`${booking.scheduleName} ${booking.service.name} ${booking.serviceTier.name}`}
            secondary={`${bookingFormat(booking.slotDate, booking.scheduleBracketSlot.startTime)}`}
          />
        </MenuItem>
      } else {
        return <span key={`appt_placeholder${i}`} />;
      }
    })}
  </Menu>
}

export default UpcomingBookingsMenu;
