import React, { useEffect } from 'react';
import { skipToken } from '@reduxjs/toolkit/query/react';

import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';

import { siteApi, theme, useAppSelector, useAuth } from 'awayto/hooks';

import Layout from './Layout';

import Onboard from './Onboard';
import ConfirmAction from './ConfirmAction';
import SnackAlert from './SnackAlert';

export default function App(props: IComponent): React.JSX.Element {

  const { setVaultKey, setUserId } = useAuth();
  const { authenticated, vaultKey } = useAppSelector(state => state.auth);

  const { data: keyRequest, isSuccess: keyIsSuccess } = siteApi.useVaultServiceGetVaultKeyQuery(authenticated ? undefined : skipToken);

  const { data: profileRequest, isLoading, isSuccess, isError } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery(vaultKey ? undefined : skipToken);

  const isAppLoading = (authenticated && isLoading) || !vaultKey;

  useEffect(() => {
    if (keyIsSuccess) {
      setVaultKey({ vaultKey: keyRequest?.key });
    }
  }, [keyIsSuccess, keyRequest]);

  useEffect(() => {
    if (profileRequest?.userProfile.email) {
      setUserId({ userId: profileRequest.userProfile.email });
    }
    if (isSuccess || isError || !isAppLoading) {
      window.INT_SITE_LOAD = true;
    }
  }, [isSuccess, isError, isAppLoading, profileRequest]);

  if (isAppLoading) {
    return <></>;
  }

  const showLayout = authenticated && profileRequest?.userProfile?.active;

  return <>
    <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale='en'>
      <ThemeProvider theme={theme}>
        <CssBaseline />

        <SnackAlert />
        <ConfirmAction />
        {showLayout ? <Layout {...props} /> : <Onboard {...props} />}
      </ThemeProvider>
    </LocalizationProvider>
  </>
}
