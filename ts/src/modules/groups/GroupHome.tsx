import React from 'react';

import Box from '@mui/material/Box';

import ManageGroups from './ManageGroups';

export function GroupHome(props: IComponent): React.JSX.Element {
  return <>
    <Box mb={4}>
      <ManageGroups {...props} />
    </Box>
  </>
}

export default GroupHome;
