import React, { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActionArea from '@mui/material/CardActionArea';

import { siteApi, useUtil, useGroupForms, targets } from 'awayto/hooks';

import ExchangeRating from './ExchangeRating';
import FormDisplay from '../forms/FormDisplay';

export function ExchangeSummary(_: IComponent): React.JSX.Element {
  const { summaryId } = useParams();
  if (!summaryId) return <></>;

  const { setSnack } = useUtil();

  const [didSubmit, setDidSubmit] = useState(false);
  const { data: bookingRequest } = siteApi.useBookingServiceGetBookingByIdQuery({ id: summaryId || '' });

  const booking = useMemo(() => bookingRequest?.booking || {}, [bookingRequest?.booking]);

  const {
    forms: serviceSurveys,
    setForm: setServiceForm,
    valid: serviceSurveyValid
  } = useGroupForms(booking?.service?.surveyIds);

  const {
    forms: tierSurveys,
    setForm: setTierForm,
    valid: tierSurveyValid
  } = useGroupForms(booking?.serviceTier?.surveyIds);

  const hasForms = Boolean(booking?.service?.surveyIds?.length || booking?.serviceTier?.surveyIds?.length);

  return <>
    <Card variant="outlined">
      <CardHeader
        title="Summary Review"
        subheader="Your feedback is important and helps make services better."
      />

      <Box ml={2}>
        <ExchangeRating rating={booking.rating} exchangeId={summaryId} />
      </Box>

      <CardContent>
        {hasForms ? <Grid container spacing={2} direction="column">
          {!!serviceSurveys?.length && booking.service && <>
            <Typography variant="h5">{booking.service.name} Survey</Typography>
            {serviceSurveys?.map((sf, i) => (
              <Box key={`service_form_intake_${i}`}>
                {i !== 0 && <Divider sx={{ my: 2 }} />}
                <Typography variant="subtitle1">{sf.name}</Typography>
                <Grid key={sf.id} pt={2} size="grow">
                  <FormDisplay form={sf} setForm={val => setServiceForm(i, val)} didSubmit={didSubmit} />
                </Grid>
              </Box>
            ))}
          </>}
          {!!tierSurveys?.length && booking.serviceTier && <>
            <Typography variant="h5">{booking.serviceTier.name} Survey</Typography>
            {tierSurveys?.map((tf, i) => (
              <Box key={`tier_form_intake_${i}`}>
                <Divider sx={{ my: 2 }} />
                <Typography variant="subtitle1">{tf.name}</Typography>
                <Grid key={tf.id} pt={2} size="grow">
                  <FormDisplay form={tf} setForm={val => setTierForm(i, val)} didSubmit={didSubmit} />
                </Grid>
              </Box>
            ))}
          </>}
        </Grid> : <Alert severity="info">
          This service requires no further feedback. Thank you!
        </Alert>}
      </CardContent>

      {hasForms && <CardActionArea
        {...targets(`exchange summary submit`, `submit post-session review forms`)}
        onClick={() => {
          if (!serviceSurveyValid || !tierSurveyValid) {
            setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out.' });
            setDidSubmit(true);
            return;
          }

          setDidSubmit(false);
        }}
      >
        <Box m={2} sx={{ display: 'flex' }}>
          <Typography color="secondary" variant="button">Complete Review</Typography>
        </Box>
      </CardActionArea>}

    </Card>
  </>;
}

export default ExchangeSummary;
