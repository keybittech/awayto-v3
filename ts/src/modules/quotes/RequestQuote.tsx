import React, { Suspense, useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router';

import Alert from '@mui/material/Alert';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import { CircularProgress } from '@mui/material';

import { useComponents, siteApi, useUtil, useGroupForm, useAccordion, useFiles, IFormVersionSubmission } from 'awayto/hooks';

import GroupContext from '../groups/GroupContext';
import GroupScheduleContext from '../group_schedules/GroupScheduleContext';
import GroupScheduleSelectionContext from '../group_schedules/GroupScheduleSelectionContext';

export function RequestQuote(props: IComponent): React.JSX.Element {

  const navigate = useNavigate();
  const { setSnack } = useUtil();
  const { ScheduleDatePicker, ScheduleTimePicker, ServiceTierAddons, AccordionWrap } = useComponents();
  const [postQuote] = siteApi.useQuoteServicePostQuoteMutation();
  const [didSubmit, setDidSubmit] = useState(false);
  const [expanded, setExpanded] = useState<string | false>(false);

  const handleChange = (panel: string) => (_: React.SyntheticEvent, isExpanded: boolean) => {
    setExpanded(isExpanded ? panel : false);
  };

  const {
    GroupSelect
  } = useContext(GroupContext) as GroupContextType;

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

  const {
    files,
    comp: FileManagerComp
  } = useFiles();

  useEffect(() => {
    getDateSlots();
  }, []);

  // const ServiceTierAddonsAccordion = useAccordion('Features', false, expanded === 'service_features', handleChange('service_features'));
  // const SelectTimeAccordion = useAccordion('Select Time', false, expanded === 'select_time', handleChange('select_time'));
  // const GroupScheduleServiceAccordion = useAccordion((groupScheduleService?.name || '') + ' Questionnaire', didSubmit && !serviceFormValid, expanded === 'service_questionnaire', handleChange('service_questionnaire'));
  // const GroupScheduleServiceTierAccordion = useAccordion((groupScheduleServiceTier?.name || '') + ' Questionnaire', didSubmit && !tierFormValid, expanded === 'tier_questionnaire', handleChange('tier_questionnaire'));
  // const FileManagerAccordion = useAccordion('Files', false, expanded === 'file_manager', handleChange('file_manager'));

  return <>
    <Grid container spacing={2} direction="column" alignContent="center">

      <Grid size={{ xs: 12, md: 10, xl: 8 }}>
        <Card>
          <CardHeader
            title="Create Request"
            subheader="Request services from a group. Some fields may be required depending on the service."
          // action={<GroupSelect />}
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
        </Alert> : <Suspense fallback={<CircularProgress />}>

          <Grid container spacing={2} mt={1}>
            <Grid container spacing={1} size={{ xs: 12, md: 4 }} direction="column">
              <Grid>
                <ScheduleDatePicker key={groupSchedule?.schedule.id} />
              </Grid>
              <Grid>
                <ScheduleTimePicker key={groupSchedule?.schedule.id} />
              </Grid>
            </Grid>
            <Grid container size={{ xs: 12, md: 8, xl: 8 }} sx={{ mt: { xs: 0, sm: 1.5 } }} spacing={2} direction="column">
              <Grid>
                <ServiceTierAddons service={groupScheduleService} />
              </Grid>
              <Grid>
                <FileManagerComp {...props} />
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
        </Suspense>}
      </Grid>

      {groupUserSchedulesRequest?.groupUserSchedules && <Grid size={{ xs: 12, md: 10, xl: 8 }}>
        <Card>
          <CardActionArea onClick={() => {
            if (!serviceFormValid || !tierFormValid || !groupScheduleServiceTier || !quote.slotDate || !quote.scheduleBracketSlotId) {
              setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out.' });
              setDidSubmit(true);
              return;
            }

            setDidSubmit(false);

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
        </Card>
      </Grid>}

    </Grid>
  </>

}

export default RequestQuote;
