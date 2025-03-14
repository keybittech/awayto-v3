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

import { siteApi, useUtil, useGroupForm, targets } from 'awayto/hooks';

import ExchangeRating from './ExchangeRating';

export function ExchangeSummary(_: IComponent): React.JSX.Element {
  const { summaryId } = useParams();
  if (!summaryId) return <></>;

  const { setSnack } = useUtil();

  const [didSubmit, setDidSubmit] = useState(false);
  const { data: bookingRequest } = siteApi.useBookingServiceGetBookingByIdQuery({ id: summaryId || '' }, { skip: !summaryId });

  const booking = useMemo(() => bookingRequest?.booking || {}, [bookingRequest?.booking]);

  const {
    form: serviceSurvey,
    comp: ServiceSurvey,
    valid: serviceSurveyValid
  } = useGroupForm(booking?.service?.surveyId, didSubmit);

  const {
    form: tierSurvey,
    comp: TierSurvey,
    valid: tierSurveyValid
  } = useGroupForm(booking?.serviceTier?.surveyId, didSubmit);

  const hasForms = booking?.service?.surveyId || booking?.serviceTier?.surveyId;

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
        <Divider sx={{ my: 2 }} />
        {hasForms ? <Grid container spacing={4} direction="column">
          <Typography variant="h6">Review Details</Typography>
          {serviceSurvey && <Grid size="grow">
            {ServiceSurvey}
          </Grid>}
          {tierSurvey && <Grid size="grow">
            {TierSurvey}
          </Grid>}
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
