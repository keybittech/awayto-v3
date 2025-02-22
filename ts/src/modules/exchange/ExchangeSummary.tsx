import React, { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';

import { siteApi, useUtil, useAccordion, useGroupForm } from 'awayto/hooks';

import ExchangeRating from './ExchangeRating';
import AccordionWrap from '../common/AccordionWrap';

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
  } = useGroupForm(booking?.service?.surveyId);

  const {
    form: tierSurvey,
    comp: TierSurvey,
    valid: tierSurveyValid
  } = useGroupForm(booking?.serviceTier?.surveyId);

  const noSurveys = !booking?.service?.surveyId && !booking?.serviceTier?.surveyId;

  const ServiceSurveyAccordion = useAccordion((booking?.quote?.serviceName || '') + ' Questionnaire', didSubmit && !serviceSurveyValid, true);
  const TierSurveyAccordion = useAccordion((booking?.quote?.serviceTierName || '') + ' Questionnaire', didSubmit && !tierSurveyValid, true);

  return <>
    <Card sx={{ mb: 2 }}>
      <CardHeader title="Summary Review" subheader="Your feedback is important and helps make services better." />
      <CardContent>
        <ExchangeRating rating={booking.rating || '0'} exchangeId={summaryId} />
      </CardContent>
    </Card>

    {noSurveys ? <Alert severity="info">
      This service requires no further feedback. Thank you!
    </Alert> : <>
      {serviceSurvey && <AccordionWrap {...ServiceSurveyAccordion}>
        <ServiceSurvey />
      </AccordionWrap>}

      {tierSurvey && <AccordionWrap {...TierSurveyAccordion}>
        <TierSurvey />
      </AccordionWrap>}


      <Card>
        <CardActionArea onClick={() => {
          if (!serviceSurveyValid || !tierSurveyValid) {
            setSnack({ snackType: 'error', snackOn: 'Please ensure all required fields are filled out.' });
            setDidSubmit(true);
            return;
          }

          setDidSubmit(false);
        }}>
          <Box m={2} sx={{ display: 'flex' }}>
            <Typography color="secondary" variant="button">Submit Request</Typography>
          </Box>
        </CardActionArea>
      </Card>
    </>}

  </>;
}

export default ExchangeSummary;
