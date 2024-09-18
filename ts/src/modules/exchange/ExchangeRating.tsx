import React, { useEffect, useState } from 'react';

import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';

import { siteApi } from 'awayto/hooks';

const ratings = ['🙁', '🙂'];

declare global {
  interface IProps {
    exchangeId?: string;
    rating?: string;
  }
}

export function ExchangeRating({ exchangeId, rating }: Required<IProps>): React.JSX.Element {
  const [changing, setChanging] = useState(false);
  const [patchBookingRating, { data: ratingDataResponse }] = siteApi.useBookingServicePatchBookingRatingMutation();

  const [currentRating, setCurrentRating] = useState(0);

  useEffect(() => {
    if (ratingDataResponse?.success) {
      setCurrentRating(parseInt(rating) || 0);
      setChanging(false);
    }
  }, [ratingDataResponse?.success, rating]);

  return <>
    <Typography variant="button">1-Click Rating</Typography>
    <Typography variant="body1">
      {!currentRating || changing && ratings.map((r, i) => <Tooltip key={`rating_${i}`} title={`Rate the Appointment`} children={
        <IconButton onClick={() => {
          patchBookingRating({ patchBookingRatingRequest: { id: exchangeId, rating: (i + 1).toString() } }).catch(console.error);
        }}>{r}</IconButton>
      } />)}
    </Typography>
    {!changing && currentRating > 0 && <>
      {`Your rating: ${ratings[currentRating - 1]} ${currentRating}`}
      <Typography sx={{ textDecoration: 'underline', cursor: 'pointer' }} onClick={() => setChanging(true)}>Change</Typography>
    </>}
  </>;
}

export default ExchangeRating;
