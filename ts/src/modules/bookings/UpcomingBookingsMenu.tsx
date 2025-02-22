import React, { useContext } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';

import JoinFullIcon from '@mui/icons-material/JoinFull';
import DoneIcon from '@mui/icons-material/Done';

import { useUtil, bookingDT, dayjs, shortNSweet } from 'awayto/hooks';

import BookingContext, { BookingContextType } from './BookingContext';

interface UpcomingBookingsMenuProps extends IComponent {
  upcomingBookingsAnchorEl: null | HTMLElement;
  upcomingBookingsMenuId: string;
  isUpcomingBookingsOpen: boolean;
  handleMenuClose: () => void;
}

export function UpcomingBookingsMenu({ handleMenuClose, upcomingBookingsAnchorEl, upcomingBookingsMenuId, isUpcomingBookingsOpen }: UpcomingBookingsMenuProps): React.JSX.Element {
  const { exchangeId } = useParams();
  const nav = useNavigate();
  const navigate = (loc: string) => {
    nav(loc);
  }
  const { openConfirm } = useUtil();

  const minsAgo15 = dayjs.duration(-15, 'years');
  // const startOfDay = dayjs().startOf('day');

  const { bookingValues: upcomingBookings } = useContext(BookingContext) as BookingContextType;

  return exchangeId ?
    <Tooltip title="Go to Exchange Summary">
      <Button
        color="success"
        aria-label={`go to exchange summary`}
        onClick={() => {
          openConfirm({
            isConfirming: true,
            confirmEffect: `Continue to the Exchange summary.`,
            confirmSideEffect: {
              approvalAction: 'All Done',
              approvalEffect: 'Continue to the Exchange summary.',
              rejectionAction: 'Keep Chatting',
              rejectionEffect: 'Return to the chat.',
            },
            confirmAction: approval => {
              if (approval) {
                navigate(`/exchange/${exchangeId}/summary`);
              }
            }
          });
        }}
        variant="outlined"
        startIcon={<DoneIcon />}
      >
        Go to Summary
      </Button>
    </Tooltip> :
    <Menu
      anchorEl={upcomingBookingsAnchorEl}
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'right',
      }}
      id={upcomingBookingsMenuId}
      keepMounted
      transformOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      open={!!isUpcomingBookingsOpen}
      onClose={handleMenuClose}
    >
      <List>
        {upcomingBookings.length ? <Box sx={{ width: 300 }}>
          {upcomingBookings.map((booking, i) => {

            if (booking.slotDate && booking.scheduleBracketSlot?.startTime && booking.scheduleBracketSlot?.id && booking.service?.name && booking.serviceTier?.name) {

              const dt = bookingDT(booking.slotDate, booking.scheduleBracketSlot.startTime);

              return <ListItem
                key={`upcoming_appt_ub_${i}`}
                secondaryAction={dayjs().isAfter(dt.add(minsAgo15)) && <>
                  <Tooltip title="Join Exchange">
                    <Button
                      aria-label={`go to exchange for ${shortNSweet(booking.slotDate, booking.scheduleBracketSlot?.startTime)}`}
                      onClick={() => {
                        navigate(`/exchange/${booking.id}`);
                      }}
                      variant="text"
                      startIcon={<JoinFullIcon />}
                    >
                      Join
                    </Button>
                  </Tooltip>
                </>}
              >
                <ListItemText
                  primary={`${shortNSweet(booking.slotDate, booking.scheduleBracketSlot.startTime)}`}
                  secondary={`${booking.service.name} ${booking.serviceTier.name}`}
                />
              </ListItem>
            } else {
              return <span key={`appt_placeholder${i}`} />;
            }
          })}
        </Box> : <Box sx={{ width: 250 }}>
          <ListItem>
            <ListItemText>No upcoming appointments.</ListItemText>
          </ListItem>
        </Box>}
      </List>
    </Menu>
}

export default UpcomingBookingsMenu;
