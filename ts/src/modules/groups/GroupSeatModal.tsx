import { useState } from "react";

import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardHeader from "@mui/material/CardHeader";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { siteApi, targets } from "awayto/hooks";


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
          <Grid size={12}>
            <TextField
              {...targets(`group seat number`, `# to add`, `modify the number of group seats to be added`)}
              fullWidth
              type="number"
              value={seatsToAdd}
              onChange={e => setSeatsToAdd(Math.max(0, parseInt(e.target.value)))}
              helperText="Input the number to add. One seat allows one login per Scheduling user per month."
            />
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent={"space-between"}>
          <Button
            {...targets(`group seat close`, `close the group seat modal`)}
            onClick={closeModal}
          >Cancel</Button>
          <Button
            {...targets(`group seat submit`, `submit the form to add more group seats`)}
            disabled={seatsToAdd < 1}
            onClick={handleSubmit}
          >Add Seats</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default GroupSeatModal;
