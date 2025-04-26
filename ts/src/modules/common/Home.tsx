import React, { useMemo } from 'react';
import { useNavigate } from 'react-router';

import Chip from '@mui/material/Chip';
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
  const secure = useSecure();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => Object.values(profileRequest?.userProfile.groups || {}).find(g => g.active), [profileRequest?.userProfile]);

  const roleActions = useMemo(() => {
    const userRoleBits = profileRequest?.userProfile.roleBits;
    if (!userRoleBits) {
      return [<></>];
    }

    const actions = [];
    for (let r in SiteRoles) {
      const roleNum = parseInt(r)
      if (roleNum > 0 && roleNum != SiteRoles.APP_GROUP_BOOKINGS && (userRoleBits & roleNum) > 0) {
        const rd = SiteRoleDetails[roleNum as SiteRoles];
        actions.push(<Button
          {...targets(`available role actions ${rd.description}`, `perform the ${rd.description} action`)}
          fullWidth
          key={`role_listing_${roleNum}`}
          variant="underline"
          sx={{
            ...classes.variableText,
            my: .5,
          }}
          onClick={() => navigate(rd.resource)}
        >
          {rd.name}
        </Button>);
      }
    }

    return actions;
  }, [profileRequest?.userProfile.roleBits]);

  return <Grid container size="grow" spacing={1}>
    <Grid size={{ sm: 12, md: 9 }}>
      <Card variant="outlined">
        <CardHeader
          title={
            <>
              <Chip variant="outlined" label={profileRequest?.userProfile.roleName} /> &nbsp;
              {profileRequest?.userProfile.firstName} {profileRequest?.userProfile.lastName}
            </>
          }
          subheader={group?.name}
        />
      </Card>
    </Grid>
    <Grid container spacing={2}>
      <Grid size={{ md: 3, xl: 2 }} sx={{ display: { xs: 'none', md: 'block' } }}>
        {roleActions}
      </Grid>

      <Grid container size={{ sm: 12, md: 9 }} direction="column" spacing={2}>
        <Grid size="grow">
          {secure([SiteRoles.APP_GROUP_BOOKINGS]) && <RequestQuote {...props} />}
          {secure([SiteRoles.APP_GROUP_SCHEDULES]) && <BookingHome {...props} />}
        </Grid>
      </Grid>
    </Grid>
  </Grid>;
}

export default Home;
