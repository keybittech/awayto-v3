import React, { useEffect, useState } from 'react';

import Grid from '@mui/material/Grid';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';

import ThumbUpIcon from '@mui/icons-material/ThumbUp';
import ThumbDownIcon from '@mui/icons-material/ThumbDown';

import { siteApi, targets } from 'awayto/hooks';

const ratings = [<ThumbDownIcon />, <ThumbUpIcon />];

interface ExchangeRatingProps extends IComponent {
  exchangeId: string;
  rating?: number;
}

export function ExchangeRating({ exchangeId, rating }: ExchangeRatingProps): React.JSX.Element {
  const [patchBookingRating] = siteApi.useBookingServicePatchBookingRatingMutation();

  const [currentRating, setCurrentRating] = useState(rating);

  useEffect(() => {
    setCurrentRating(rating);
  }, [rating]);

  return <Grid container spacing={1}>
    {ratings.map((r, i) => <Tooltip key={`rating_${i + 1}`} title={`Rate the Appointment`} children={
      <IconButton
        {...targets(`exchange summary rating ${i + 1}`, `rate the service low to high, currently ${i + 1}`)}
        size="large"
        sx={{
          bgcolor: currentRating == i + 1 ? 'secondary.dark' : 'unset'
        }}
        onClick={() => {
          setCurrentRating(i + 1);
          patchBookingRating({
            patchBookingRatingRequest: {
              id: exchangeId,
              rating: i + 1
            }
          });
        }}
      >
        {r}
      </IconButton>
    } />)}
  </Grid>;
}

export default ExchangeRating;
