import { useState } from 'react';

import Button from '@mui/material/Button';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';

import { plural, targets, useStyles, siteApi } from 'awayto/hooks';

export function GroupSeatModal({ closeModal }: IComponent): React.JSX.Element {

  const classes = useStyles();

  const [postGroupSeats] = siteApi.useGroupSeatServicePostGroupSeatMutation();

  return <>
    <DialogTitle>Group Seats</DialogTitle>
    <DialogContent>
      <Typography>One seat allows one login per Scheduling user per month.</Typography>
      <Grid container spacing={4} sx={{ justifyContent: 'space-evenly' }}>
        {[1, 5, 10, 100].map((gs, i) => {
          const label = plural(gs, 'seat', 'seats');
          return <Button
            {...targets(`group seat modal ${gs} seats`, `add ${label} to group`)}
            key={`group_seat_add_${i}`}
            variant="underline"
            sx={{
              ...classes.variableText,
              my: .5,
            }}
            onClick={_ => postGroupSeats({ postGroupSeatRequest: { seats: gs } })}
          >
            {label}
          </Button>;
        })}
      </Grid>
    </DialogContent >
    <DialogActions>
      <Grid container justifyContent={"space-between"}>
        <Button
          {...targets(`group seat close`, `close the group seat modal`)}
          onClick={closeModal}
        >Cancel</Button>
        {/* <Button */}
        {/*   {...targets(`group seat submit`, `submit the form to add more group seats`)} */}
        {/*   disabled={seatsToAdd < 1} */}
        {/*   onClick={handleSubmit} */}
        {/* >Add Seats</Button> */}
      </Grid>
    </DialogActions>
  </>
}

export default GroupSeatModal;
