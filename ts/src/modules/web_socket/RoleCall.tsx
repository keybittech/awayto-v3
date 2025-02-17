import React from 'react';

import { SocketActions, siteApi, useWebSocketSubscribe } from 'awayto/hooks';

const RoleCall = ({ children }: IComponent): React.JSX.Element => {

  const { refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  useWebSocketSubscribe<{ roleCallTime: number }>("group:role_call", async ({ action }) => {
    if (SocketActions.ROLE_CALL == action) {
      console.log({ rolecalling: true })
      await getUserProfileDetails();
    }
  });

  return <>
    {children}
  </>;
}

export default RoleCall;
