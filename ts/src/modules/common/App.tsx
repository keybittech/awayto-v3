import React, { useContext, useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import ThemeProvider from '@mui/material/styles/ThemeProvider';
import CssBaseline from '@mui/material/CssBaseline';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';

import { siteApi, useAppSelector, useComponents, theme, SiteRoles, refreshToken } from 'awayto/hooks';

import Layout from './Layout';
import reportWebVitals from '../../reportWebVitals';
import AuthContext from '../auth/AuthContext';
import { PaletteMode, useColorScheme } from '@mui/material';

const {
  REACT_APP_KC_CLIENT
} = process.env as { [prop: string]: string };

export default function App(props: IComponent): React.JSX.Element {
  // const location = useLocation();
  const navigate = useNavigate();

  const { keycloak } = useContext(AuthContext) as AuthContextType;

  const { ConfirmAction, SnackAlert, Onboard } = useComponents();

  const [ready, setReady] = useState(false);
  const [onboarding, setOnboarding] = useState(false);

  const { data: profileRes, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const { mode } = useColorScheme();

  const reloadProfile = async (): Promise<void> => {
    await refreshToken(61).then(async () => {
      await getUserProfileDetails().unwrap();
      navigate('/');
    }).catch(console.error);
  }

  useEffect(() => {
    // refresh the user profile every minute
    const interval: NodeJS.Timeout = setInterval(() => {
      const resources = keycloak.tokenParsed?.resource_access;
      if (resources && resources[REACT_APP_KC_CLIENT]?.roles.includes(SiteRoles.APP_ROLE_CALL)) {
        void getUserProfileDetails();
      }
    }, 58 * 1000);

    reportWebVitals(console.info);

    return () => {
      clearInterval(interval);
    }
  }, []);

  useEffect(() => {
    if (!profileRes) return;
    const profile = profileRes.userProfile;

    if (!profile.active) {
      setOnboarding(true);
    } else if (profile.active) {
      setOnboarding(false);
      setReady(true);
    }
  }, [profileRes]);

  return <>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <ThemeProvider theme={theme} defaultMode={mode}>
        <CssBaseline />

        <SnackAlert />
        <ConfirmAction />
        {onboarding && <Onboard {...props} reloadProfile={reloadProfile} />}
        {ready && <Layout {...props} />}
      </ThemeProvider>
    </LocalizationProvider>
  </>
}
