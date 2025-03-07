import React, { useMemo } from 'react';
import { useNavigate } from 'react-router';

import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Button from '@mui/material/Button';

import { siteApi, SiteRoleDetails, SiteRoles, targets, useSecure } from 'awayto/hooks';

import BookingHome from '../bookings/BookingHome';
import RequestQuote from '../quotes/RequestQuote';

export function Home(props: IComponent): React.JSX.Element {

  const navigate = useNavigate();
  const hasRole = useSecure();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => Object.values(profileRequest?.userProfile.groups || {}).find(g => g.active), [profileRequest?.userProfile]);

  const roleActions = useMemo(() => {
    const augr = profileRequest?.userProfile.availableUserGroupRoles;
    if (!augr) return <></>;
    return Object.values(SiteRoles).filter(r => augr.includes(r)).map((r, i) => {
      const rd = SiteRoleDetails[r];
      return <Button
        {...targets(`available role actions ${rd.description}`, `perform the ${rd.description} action`)}
        fullWidth
        key={`role_listing_${i + 1}`}
        sx={{
          my: .5,
          background: 'linear-gradient(to top, rgba(255, 255, 255, .05) 0%, transparent 33%)',
          textAlign: 'left',
          justifyContent: 'left',
          borderRadius: 0,
          textDecoration: 'underlined',
          borderBottom: '1px solid #aaa',
        }}
        variant="text"
        onClick={() => navigate(rd.resource)}
      >
        {rd.description}
      </Button>;
    });
  }, [profileRequest?.userProfile.availableUserGroupRoles, navigate]);

  return <Grid container size="grow">
    <Grid container spacing={2}>
      <Grid size={{ xs: 12, md: 3, lg: 2 }}>
        <Card sx={{ p: 2 }} variant="outlined">
          <CardHeader
            title={`${profileRequest?.userProfile.firstName} ${profileRequest?.userProfile.lastName}`}
            subheader={`${group?.name} ${profileRequest?.userProfile.roleName}`}
          />
        </Card>
        {roleActions}
      </Grid>



      <Grid container size={{ xs: 12, md: 9, lg: 10 }} direction="column" spacing={2}>

        {hasRole([SiteRoles.APP_GROUP_BOOKINGS]) && <Grid size="grow">
          <RequestQuote {...props} />
        </Grid>}

      </Grid>
    </Grid>
    <Grid size="grow">
      <BookingHome {...props} />
    </Grid>
  </Grid>;
}

export default Home;
