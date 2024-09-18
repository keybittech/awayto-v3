import React, { useState, useCallback, useContext, useEffect } from 'react';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';

import { useComponents, useContexts, IGroupUserScheduleStub, shortNSweet, siteApi } from 'awayto/hooks';

declare global {
  interface IComponent {
    editGroupUserScheduleStub?: Required<IGroupUserScheduleStub>;
  }
}

export function ManageScheduleStubModal({ editGroupUserScheduleStub, closeModal }: Required<IComponent>): React.JSX.Element {

  const { GroupScheduleContext, GroupScheduleSelectionContext } = useContexts();

  const {
    selectGroupSchedule: { item: groupSchedule },
    getGroupUserSchedules: { data: getGroupUserSchedulesRequest }
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const {
    quote,
    selectedDate,
    firstAvailable
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const { ScheduleDatePicker, ScheduleTimePicker } = useComponents();

  const [patchGroupUserScheduleStubReplacement] = siteApi.useGroupUserScheduleServicePatchGroupUserScheduleStubReplacementMutation();
  const [getGroupUserScheduleStubReplacement] = siteApi.useLazyGroupUserScheduleServiceGetGroupUserScheduleStubReplacementQuery();

  const [replacement, setReplacement] = useState(editGroupUserScheduleStub.replacement);

  const originalReplacement = editGroupUserScheduleStub?.replacement && { ...editGroupUserScheduleStub.replacement };

  useEffect(() => {
    if (Object.keys(quote).length) {
      const { userScheduleId, tierName } = editGroupUserScheduleStub;

      if (userScheduleId && quote.slotDate && quote.startTime && tierName) {
        getGroupUserScheduleStubReplacement({
          userScheduleId,
          slotDate: quote.slotDate,
          startTime: quote.startTime,
          tierName
        }).unwrap().then(stubsResponse => {
          if (stubsResponse.groupUserScheduleStubs[0].replacement) {
            setReplacement(stubsResponse.groupUserScheduleStubs[0].replacement);
          }
        }).catch(console.error);
      }
    }
  }, [quote]);

  const handleSubmit = useCallback(() => {
    patchGroupUserScheduleStubReplacement({
      patchGroupUserScheduleStubReplacementRequest: {
        userScheduleId: editGroupUserScheduleStub.userScheduleId!,
        quoteId: editGroupUserScheduleStub.quoteId!,
        slotDate: (selectedDate || firstAvailable.time).format("YYYY-MM-DD"),
        startTime: replacement?.startTime!,
        serviceTierId: replacement?.serviceTierId!,
        scheduleBracketSlotId: replacement?.scheduleBracketSlotId!

      }
    }).unwrap().then(() => {
      if (closeModal)
        closeModal();
    }).catch(console.error);
  }, [editGroupUserScheduleStub, replacement]);

  return <>
    <Card>
      <CardHeader title={`${shortNSweet(editGroupUserScheduleStub.slotDate, editGroupUserScheduleStub.startTime)}`} subheader={`${editGroupUserScheduleStub.serviceName} ${editGroupUserScheduleStub.tierName}`} />
      <CardContent>

        {groupSchedule && !getGroupUserSchedulesRequest?.groupUserSchedules?.length ? <Alert severity="info">

          The master schedule {groupSchedule.schedule?.name} has no available user schedules.
        </Alert> : <>
          {originalReplacement && <>
            <Box mb={2}>
              <Typography>Use an existing slot at the same date and time:</Typography>
            </Box>
            <Box mb={4}>
              <Button fullWidth variant="contained" color="secondary" onClick={handleSubmit}>Reassign to {originalReplacement.username}</Button>
            </Box>

            <Grid container direction="row" alignItems="center" spacing={2} mb={4}>
              <Grid item flexGrow={1}>
                <Divider />
              </Grid>
              <Grid item>
                Or
              </Grid>
              <Grid item flexGrow={1}>
                <Divider />
              </Grid>
            </Grid>
          </>}

          <Box mb={4}>
            <Box mb={2}>
              <Typography>Select a new date and time:</Typography>
            </Box>
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <ScheduleDatePicker
                  key={editGroupUserScheduleStub.userScheduleId}
                />
              </Grid>
              <Grid item xs={6}>
                <ScheduleTimePicker
                  key={selectedDate?.format("YYYY-MM-DD")}
                />
              </Grid>
            </Grid>
          </Box>

          {replacement?.username && <Box my={2}>
            <Button onClick={handleSubmit} fullWidth variant="contained" color="secondary">Reassign to {replacement.username}</Button>
          </Box>}
        </>}


      </CardContent>
      <CardActions>
        <Grid container justifyContent="space-between">
          <Button onClick={closeModal}>Close</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default ManageScheduleStubModal;
