import React, { useEffect } from 'react';

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

  const { setVaultKey } = useAuth();
  const { authenticated } = useAppSelector(state => state.auth);

  const { data: profileRequest, isLoading, isSuccess, isError } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery(undefined, {
    skip: !authenticated,
  });

  const isAppLoading = authenticated && isLoading;

  useEffect(() => {
    if (isSuccess || isError || !isAppLoading) {
      window.INT_SITE_LOAD = true;
      if (profileRequest?.userProfile.vaultKey) {
        setVaultKey({ vaultKey: profileRequest?.userProfile.vaultKey });
      }
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
