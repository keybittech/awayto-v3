import React, { useMemo } from 'react';
import { useNavigate } from 'react-router';

import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Button from '@mui/material/Button';

import { siteApi, SiteRoleDetails, SiteRoles, targets, useSecure, useStyles } from 'awayto/hooks';

import BookingHome from '../bookings/BookingHome';
import RequestQuote from '../quotes/RequestQuote';

export function Home(props: IComponent): React.JSX.Element {

  const classes = useStyles();
  const navigate = useNavigate();
  const hasRole = useSecure();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => Object.values(profileRequest?.userProfile.groups || {}).find(g => g.active), [profileRequest?.userProfile]);

  const roleActions = useMemo(() => {
    const augr = profileRequest?.userProfile.availableUserGroupRoles?.filter(x => ![SiteRoles.APP_GROUP_BOOKINGS].includes(x as SiteRoles));
    if (!augr) return <></>;
    return Object.values(SiteRoles).filter(r => augr.includes(r)).map((r, i) => {
      const rd = SiteRoleDetails[r];
      return <Button
        {...targets(`available role actions ${rd.description}`, `perform the ${rd.description} action`)}
        fullWidth
        key={`role_listing_${i + 1}`}
        variant="underline"
        sx={{
          ...classes.variableText,
          my: .5,
        }}
        onClick={() => navigate(rd.resource)}
      >
        {rd.name}
      </Button>;
    });
  }, [profileRequest?.userProfile.availableUserGroupRoles]);

  return <Grid container size="grow">
    <Grid container spacing={2}>
      <Grid size={{ md: 3, xl: 2 }}>
        <Card variant="outlined">
          <CardHeader
            title={`${profileRequest?.userProfile.firstName} ${profileRequest?.userProfile.lastName}`}
            subheader={`${group?.name} ${profileRequest?.userProfile.roleName}`}
          />
        </Card>
        <Grid sx={{ display: { xs: 'none', md: 'block' } }}>
          {roleActions}
        </Grid>
      </Grid>

      <Grid container size={{ xs: 12, md: 9, lg: 8, xl: 6 }} direction="column" spacing={2}>
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
