import { useState } from "react";

import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardHeader from "@mui/material/CardHeader";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { siteApi } from "awayto/hooks";


export function GroupSeatModal({ closeModal }: IComponent): React.JSX.Element {

  const [seatsToAdd, setSeatsToAdd] = useState(0);

  const [postGroupSeats] = siteApi.useGroupSeatServicePostGroupSeatMutation();

  const handleSubmit = function() {
    console.log({ seatsToAdd })

    void postGroupSeats({ postGroupSeatRequest: { seats: seatsToAdd } });
  }

  return <>
    <Card>
      <CardHeader title={`Group Seats`}></CardHeader>
      <CardContent>
        <Grid container spacing={4}>
          <Grid item xs={12}>
            <TextField
              fullWidth
              type="number"
              id="seatsToAdd"
              label="# to add"
              value={seatsToAdd}
              name="seatsToAdd"
              onChange={e => setSeatsToAdd(Math.max(0, parseInt(e.target.value)))}
              helperText="Input the number to purchase. One seat allows one login per Scheduling user per month."
            />
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent={"space-between"}>
          <Button onClick={closeModal}>Cancel</Button>
          <Button disabled={seatsToAdd < 1} onClick={handleSubmit}>Add Seats</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default GroupSeatModal;
