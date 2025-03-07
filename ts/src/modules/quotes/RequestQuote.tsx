import React, { useContext, useEffect, useState } from 'react';

import Alert from '@mui/material/Alert';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import Slide from '@mui/material/Slide';
import { TransitionProps } from '@mui/material/transitions';

import { siteApi, useUtil, useGroupForm, IFormVersionSubmission, IFile, bookingFormat } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from '../group_schedules/GroupScheduleSelectionContext';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import ScheduleDatePicker from '../group_schedules/ScheduleDatePicker';
import ScheduleTimePicker from '../group_schedules/ScheduleTimePicker';
import ServiceTierAddons from '../services/ServiceTierAddons';
import FileManager from '../files/FileManager';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>,
) {
  return <Slide direction="down" ref={ref} {...props} />;
});

export function RequestQuote(_: IComponent): React.JSX.Element {

  const { setSnack, openConfirm } = useUtil();
  const [postQuote] = siteApi.useQuoteServicePostQuoteMutation();
  const [files, setFiles] = useState<IFile[]>([]);
  const [dialog, setDialog] = useState(false);

  const {
    getGroupUserSchedules: {
      data: groupUserSchedulesRequest
    },
    selectGroupSchedule: {
      item: groupSchedule,
      comp: GroupScheduleSelect
    },
    selectGroupScheduleService: {
      item: groupScheduleService,
      comp: GroupScheduleServiceSelect
    },
    selectGroupScheduleServiceTier: {
      item: groupScheduleServiceTier,
      comp: GroupScheduleServiceTierSelect
    }
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const {
    quote,
    getDateSlots,
    setSelectedDate,
    setSelectedTime,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const {
    form: serviceForm,
    comp: ServiceForm,
    valid: serviceFormValid
  } = useGroupForm(groupScheduleService?.formId);

  const {
    form: tierForm,
    comp: TierForm,
    valid: tierFormValid
  } = useGroupForm(groupScheduleServiceTier?.formId);

  useEffect(() => {
    getDateSlots();
  }, []);

  if (!groupSchedule?.id) {
    return <Alert severity="info">
      There are no group schedules, or they are currently paused.
    </Alert>
  }

  return <>
    <Card variant="outlined">
      <CardHeader
        title="Service Request"
        subheader="Request services from a group. Some fields may be required depending on the service."
      />
      <CardContent>
        <Grid container spacing={2}>
          <Grid container size={{ xs: 12, md: 4 }}>
            <GroupScheduleSelect />
            <GroupScheduleServiceSelect />
            <GroupScheduleServiceTierSelect />
          </Grid>
          <Grid container size={{ xs: 12, md: 4 }} alignContent="flex-start">
            <ScheduleDatePicker key={`${groupSchedule?.schedule?.id}_date_picker`} />
            <ScheduleTimePicker key={`${groupSchedule?.schedule?.id}_time_picker`} />
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            {groupScheduleServiceTier && <Button
              fullWidth
              onClick={() => { setDialog(true) }}
            >
              View Features
            </Button>}
          </Grid>
        </Grid>
      </CardContent>
      <CardActionArea onClick={() => {
        if (!serviceFormValid || !tierFormValid || !groupScheduleServiceTier?.id || !quote.slotDate || !quote.startTime || !quote.scheduleBracketSlotId) {
          setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out and without error.' });
          return;
        }

        openConfirm({
          isConfirming: true,
          confirmEffect: 'Request service on ' + bookingFormat(quote.slotDate, quote.startTime) +
            ' for ' + quote.serviceTierName + ' ' + quote.serviceName,
          confirmAction: () => {
            postQuote({
              postQuoteRequest: {
                slotDate: quote.slotDate!,
                scheduleBracketSlotId: quote.scheduleBracketSlotId!,
                serviceTierId: groupScheduleServiceTier.id!,
                serviceFormVersionSubmission: (serviceForm ? {
                  formVersionId: serviceForm.version.id,
                  submission: serviceForm.version.submission
                } : {}) as IFormVersionSubmission,
                tierFormVersionSubmission: (tierForm ? {
                  formVersionId: tierForm.version.id,
                  submission: tierForm.version.submission
                } : {}) as IFormVersionSubmission,
                files
              }
            }).unwrap().then(() => {
              setSnack({ snackType: 'success', snackOn: 'You\'re all set!' });
              setSelectedDate(undefined);
              setSelectedTime(undefined);
            }).catch(console.error);
          }
        });

      }}>
        <Box m={2} sx={{ display: 'flex' }}>
          <Typography color="secondary" variant="button">Submit Request</Typography>
        </Box>
      </CardActionArea>
    </Card>

    {!groupUserSchedulesRequest?.groupUserSchedules ? <Alert sx={{ marginTop: '16px' }} severity="info">
      There are no active user schedules, or they are currently paused.
    </Alert> : <>
      <Grid container size="grow" spacing={2}>
        <Grid container size={{ xs: 12, md: 4 }} direction="column">
        </Grid>
        <Grid container size={{ xs: 12, md: 8 }} sx={{ my: 1.5 }} spacing={2} direction="column">
          <Grid size="grow">
          </Grid>
        </Grid>
      </Grid>

      <Grid container spacing={2} direction="column">
        {serviceForm && <Grid size="grow">
          <ServiceForm />
        </Grid>}
        {tierForm && <Grid size="grow">
          <TierForm />
        </Grid>}
      </Grid>
    </>}

    <Dialog
      fullWidth
      keepMounted
      maxWidth="md"
      open={dialog}
      onClose={() => { setDialog(false) }}
      slots={{
        transition: Transition
      }}
    >
      <FileManager files={files} setFiles={setFiles} />
      <ServiceTierAddons service={groupScheduleService} />
    </Dialog>
  </>
}

export default RequestQuote;
