import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router';

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

import { siteApi, useUtil, useGroupForm, IFormVersionSubmission, IFile, bookingFormat, targets, useStyles } from 'awayto/hooks';

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

  const classes = useStyles();
  const navigate = useNavigate();

  const { setSnack, openConfirm } = useUtil();
  const [postQuote] = siteApi.useQuoteServicePostQuoteMutation();
  const [files, setFiles] = useState<IFile[]>([]);
  const [dialog, setDialog] = useState('');
  const [didSubmit, setDidSubmit] = useState(false);

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
  } = useGroupForm(groupScheduleService?.formId, didSubmit);

  const {
    form: tierForm,
    comp: TierForm,
    valid: tierFormValid
  } = useGroupForm(groupScheduleServiceTier?.formId, didSubmit);

  useEffect(() => {
    getDateSlots();
  }, []);

  if (!groupSchedule?.id) {
    return <Alert
      severity="info"
      action={
        <Button
          {...targets(`request quote create group schedule`, `go to the group schedule page to create a group master schedule`)}
          variant="text"
          color="info"
          onClick={() => navigate('/group/manage/schedules')}
        >
          Create a master schedule
        </Button>
      }
    >
      There are no group schedules, or they are currently paused.
    </Alert>
  }

  const hasForms = Boolean(serviceForm?.id || tierForm?.id);

  return <>
    <Card variant="outlined">
      <CardHeader
        title="Service Request"
        subheader="Request services from a group. Some fields may be required depending on the service."
      />
      {!groupUserSchedulesRequest?.groupUserSchedules ? <>
        <Grid container p={2}>
          <GroupScheduleSelect />
        </Grid>
        <Alert
          severity="info"
          action={
            <Button
              {...targets(`request quote create personal schedule`, `go to the personal schedule page to create a personal schedule`)}
              variant="text"
              color="info"
              onClick={() => navigate('/schedule')}
            >
              Create your schedule
            </Button>
          }
        >
          There are no active user schedules, or they are currently paused.
        </Alert>
      </> : <>
        <CardContent>
          <Grid container spacing={2}>
            <Grid container size={{ xs: 12, md: 5 }}>
              <GroupScheduleSelect />
              <GroupScheduleServiceSelect />
              <GroupScheduleServiceTierSelect />
            </Grid>
            <Grid container size={{ xs: 12, md: 5 }} alignContent="flex-start">
              <ScheduleDatePicker key={`${groupSchedule?.schedule?.id}_date_picker`} />
              <ScheduleTimePicker key={`${groupSchedule?.schedule?.id}_time_picker`} />
            </Grid>
            <Grid size={{ xs: 12, md: 2 }}>
              {groupScheduleServiceTier && <Button
                sx={classes.variableText}
                fullWidth
                onClick={() => { setDialog('features') }}
                variant="underline"
              >
                View Features
              </Button>}
              <Button
                sx={classes.variableText}
                fullWidth
                onClick={() => { setDialog('files') }}
                variant="underline"
              >
                Add Files
              </Button>
            </Grid>
          </Grid>
          {hasForms && <Grid container mt={2} spacing={2} direction="column">
            <Typography variant="h6">Request Details</Typography>
            {serviceForm && <Grid size="grow">
              {ServiceForm}
            </Grid>}
            {tierForm && <Grid size="grow">
              {TierForm}
            </Grid>}
          </Grid>}
        </CardContent>
        <CardActionArea
          {...targets(`request quote submit request`, `submit your completed request for service`)}
          onClick={() => {
            setDidSubmit(true);
            if (!serviceFormValid || !tierFormValid || !groupScheduleServiceTier?.id || !quote.slotDate || !quote.startTime || !quote.scheduleBracketSlotId) {
              setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out and without error.' });
              return;
            }

            openConfirm({
              isConfirming: true,
              confirmEffect: 'Request service on ' + bookingFormat(quote.slotDate, quote.startTime) +
                ' for ' + groupScheduleService?.name + ': ' + groupScheduleServiceTier.name,
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
                }).catch((e: { data: string }) => {
                  if (e.data.includes('select a new time')) {
                    getDateSlots();
                  }
                });
              }
            });

          }}
        >
          <Box m={2} sx={{ display: 'flex' }}>
            <Typography color="secondary" variant="button">Submit Request</Typography>
          </Box>
        </CardActionArea>
      </>}
    </Card>

    <Dialog
      fullWidth
      maxWidth="md"
      open={!!dialog.length}
      onClose={() => { setDialog('') }}
      slotProps={{
        paper: {
          sx: { bgcolor: 'secondary.contrastText' }
        }
      }}
      slots={{
        transition: Transition
      }}
    >
      {dialog == 'features' ?
        <ServiceTierAddons service={groupScheduleService} /> :
        dialog == 'files' ? <FileManager files={files} setFiles={setFiles} /> :
          <></>
      }
    </Dialog>
  </>
}

export default RequestQuote;
