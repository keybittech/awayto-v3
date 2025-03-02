import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router';

import Alert from '@mui/material/Alert';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';

import { siteApi, useUtil, useGroupForm, IFormVersionSubmission, IFile } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from '../group_schedules/GroupScheduleSelectionContext';
import ScheduleDatePicker from '../group_schedules/ScheduleDatePicker';
import ScheduleTimePicker from '../group_schedules/ScheduleTimePicker';
import ServiceTierAddons from '../services/ServiceTierAddons';
import FileManager from '../files/FileManager';

export function RequestQuote(_: IComponent): React.JSX.Element {

  const navigate = useNavigate();
  const { setSnack } = useUtil();
  const [postQuote] = siteApi.useQuoteServicePostQuoteMutation();
  const [files, setFiles] = useState<IFile[]>([]);

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

  const { quote, getDateSlots } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

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

  return <>

    <Card variant="outlined" sx={{ bgcolor: 'primary.dark' }}>
      <CardHeader
        title="Create Request"
        subheader="Request services from a group. Some fields may be required depending on the service."
      />
      <CardContent>
        <Grid container spacing={2}>
          <Grid size={4}>
            <GroupScheduleSelect />
          </Grid>
          <Grid size={4}>
            <GroupScheduleServiceSelect />
          </Grid>
          <Grid size={4}>
            <GroupScheduleServiceTierSelect />
          </Grid>
        </Grid>

      </CardContent>
    </Card>

    {!groupUserSchedulesRequest?.groupUserSchedules ? <Alert sx={{ marginTop: '16px' }} severity="info">
      There are no active schedules or operations are currently halted.
    </Alert> : <>
      <Grid container spacing={2} mt={1}>
        <Grid container spacing={1} size={{ xs: 12, md: 4 }} direction="column">
          <Grid>
            <ScheduleDatePicker key={groupSchedule?.schedule?.id} />
          </Grid>
          <Grid>
            <ScheduleTimePicker key={groupSchedule?.schedule?.id} />
          </Grid>
        </Grid>
        <Grid container size={{ xs: 12, md: 8, xl: 8 }} sx={{ mt: { xs: 0, sm: 1.5 } }} spacing={2} direction="column">
          <Grid>
            <ServiceTierAddons service={groupScheduleService} />
          </Grid>
          <Grid>
            <FileManager files={files} setFiles={setFiles} />
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

    {groupUserSchedulesRequest?.groupUserSchedules && <Card>
      <CardActionArea onClick={() => {
        if (!serviceFormValid || !tierFormValid || !groupScheduleServiceTier?.id || !quote.slotDate || !quote.scheduleBracketSlotId) {
          setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out.' });
          return;
        }

        postQuote({
          postQuoteRequest: {
            slotDate: quote.slotDate,
            scheduleBracketSlotId: quote.scheduleBracketSlotId,
            serviceTierId: groupScheduleServiceTier.id,
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
          setSnack({ snackOn: 'Your request has been made successfully!' });
          navigate('/');
        }).catch(console.error);
      }}>
        <Box m={2} sx={{ display: 'flex' }}>
          <Typography color="secondary" variant="button">Submit Request</Typography>
        </Box>
      </CardActionArea>
    </Card>}

  </>
}

export default RequestQuote;
