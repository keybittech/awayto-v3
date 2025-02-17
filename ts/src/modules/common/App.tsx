import React from 'react';

import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';

import { siteApi, theme, useAppSelector } from 'awayto/hooks';

import Layout from './Layout';

import Onboard from './Onboard';
import ConfirmAction from './ConfirmAction';
import SnackAlert from './SnackAlert';

export default function App(props: IComponent): React.JSX.Element {

  const { authenticated } = useAppSelector(state => state.auth);

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  return <>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <ThemeProvider theme={theme}>
        <CssBaseline />

        <SnackAlert />
        <ConfirmAction />
        {authenticated && profileRequest?.userProfile?.active ? <Layout {...props} /> : <Onboard {...props} />}
      </ThemeProvider>
    </LocalizationProvider>
  </>
}
