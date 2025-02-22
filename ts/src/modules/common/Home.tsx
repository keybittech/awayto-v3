import React, { useContext, useMemo } from 'react';

import Box from '@mui/material/Box';

import { siteApi, SiteRoleDetails, SiteRoles } from 'awayto/hooks';
import { Card, CardHeader, Chip, Tooltip } from '@mui/material';
import { useNavigate } from 'react-router';

import GroupContext, { GroupContextType } from '../groups/GroupContext';
import BookingHome from '../bookings/BookingHome';
import PendingQuotesProvider from '../quotes/PendingQuotesProvider';
import QuoteHome from '../quotes/QuoteHome';

export function Home(props: IComponent): React.JSX.Element {

  const {
    GroupSelect
  } = useContext(GroupContext) as GroupContextType;

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const navigate = useNavigate();

  const roleActions = useMemo(() => {
    const augr = profileRequest?.userProfile.availableUserGroupRoles;
    if (!augr) return <></>;
    return Object.values(SiteRoles).filter(r => augr.includes(r)).map((r, i) => {
      const rd = SiteRoleDetails[r];
      return <Tooltip key={`role_listing_${i + 1}`} title={rd.name} >
        <Chip sx={{ margin: .5 }} color="info" label={rd.description} onClick={() => navigate(rd.resource)} />
      </Tooltip>;
    });
  }, [profileRequest?.userProfile.availableUserGroupRoles, navigate]);

  return <>
    <Card sx={{ padding: '12px' }} variant="outlined">
      <CardHeader
        title={`${profileRequest?.userProfile.firstName} ${profileRequest?.userProfile.lastName}`}
        subheader={`${profileRequest?.userProfile.roleName}`}
        action={<GroupSelect />}
      />
      {roleActions}
    </Card>
    <Box mb={2}>
      <BookingHome {...props} />
    </Box>
    {/* <Box mb={2}>
        <GroupHome {...props} />
      </Box> */}
    <PendingQuotesProvider>
      <QuoteHome {...props} />
    </PendingQuotesProvider>
  </>;
}

export default Home;
