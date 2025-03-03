import React, { useMemo } from 'react';
import { useNavigate } from 'react-router';

import Grid from '@mui/material/Grid';
import Tooltip from '@mui/material/Tooltip';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';

import { siteApi, SiteRoleDetails, SiteRoles, targets, useSecure } from 'awayto/hooks';

import BookingHome from '../bookings/BookingHome';
import PendingQuotesProvider from '../quotes/PendingQuotesProvider';
import QuoteHome from '../quotes/QuoteHome';
import RequestQuote from '../quotes/RequestQuote';

export function Home(props: IComponent): React.JSX.Element {

  const navigate = useNavigate();
  const hasRole = useSecure();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const roleActions = useMemo(() => {
    const augr = profileRequest?.userProfile.availableUserGroupRoles;
    if (!augr) return <></>;
    return Object.values(SiteRoles).filter(r => augr.includes(r)).map((r, i) => {
      const rd = SiteRoleDetails[r];
      return <Tooltip key={`role_listing_${i + 1}`} title={rd.name} >
        <Chip
          {...targets(`available role actions ${rd.description}`, rd.description, `perform the ${rd.description} action`)}
          sx={{ margin: .5 }}
          color="info"
          onClick={() => navigate(rd.resource)}
        />
      </Tooltip>;
    });
  }, [profileRequest?.userProfile.availableUserGroupRoles, navigate]);

  return <Grid container size={{ xs: 12, md: 10, lg: 9, xl: 8 }}>
    <Grid container spacing={2}>
      <Grid size={{ xs: 12, md: 3 }}>
        <Card sx={{ p: 2 }} variant="outlined">
          <CardHeader
            title={`${profileRequest?.userProfile.firstName} ${profileRequest?.userProfile.lastName}`}
            subheader={`${profileRequest?.userProfile.roleName}`}
          // action={<GroupSelect />}
          />
          {roleActions}
        </Card>
      </Grid>



      <Grid container size={{ xs: 12, md: 9 }} direction="column" spacing={2}>

        {hasRole([SiteRoles.APP_GROUP_BOOKINGS]) && <Grid size="grow">
          <RequestQuote {...props} />
        </Grid>}


        {hasRole([SiteRoles.APP_GROUP_SCHEDULES]) && <Grid size="grow">
          <PendingQuotesProvider>
            <QuoteHome {...props} />
          </PendingQuotesProvider>
        </Grid>}

      </Grid>
    </Grid>
    <Grid size="grow">
      <BookingHome {...props} />
    </Grid>
  </Grid>;
}

export default Home;
