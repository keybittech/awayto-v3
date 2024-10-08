import React, { useContext, useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import ThemeProvider from '@mui/material/styles/ThemeProvider';
import CssBaseline from '@mui/material/CssBaseline';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';

import { siteApi, useAppSelector, useComponents, lightTheme, darkTheme, useContexts, SiteRoles } from 'awayto/hooks';

import Layout from './Layout';
import reportWebVitals from '../../reportWebVitals';

const {
  REACT_APP_KC_CLIENT
} = process.env as { [prop: string]: string };

export default function App(props: IComponent): React.JSX.Element {
  // const location = useLocation();
  const navigate = useNavigate();

  const { AuthContext } = useContexts();
  const { keycloak, refreshToken } = useContext(AuthContext) as AuthContextType;

  const { ConfirmAction, SnackAlert, Onboard } = useComponents();

  const { variant } = useAppSelector(state => state.theme);

  const [ready, setReady] = useState(false);
  const [onboarding, setOnboarding] = useState(false);

  const { data: profileRes, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  // const [attachUser] = siteApi.useGroupServiceAttachUserMutation();
  // const [activateProfile] = siteApi.useUserProfileServiceActivateProfileMutation();

  const reloadProfile = async (): Promise<void> => {
    await refreshToken().then(() => {
      void getUserProfileDetails();
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

    // if (location.pathname === "/registration/code/success") {
    //   const code = location.search.split('?code=')[1].split('&')[0];
    //   attachUser({ attachUserRequest: { code } }).unwrap().then(async () => {
    //     await activateProfile().unwrap().catch(console.error);
    //     await reloadProfile().catch(console.error);
    //   }).catch(console.error);
    // } else 

    if (!profile.active) {
      setOnboarding(true);
    } else if (profile.active) {
      setOnboarding(false);
      setReady(true);
    }
  }, [profileRes]);

  return <>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <ThemeProvider theme={'light' === variant ? lightTheme : darkTheme}>
        <CssBaseline />

        <SnackAlert />
        <ConfirmAction />
        {onboarding && <Onboard {...props} reloadProfile={reloadProfile} />}
        {ready && <Layout {...props} />}

        {/* !!isLoading && <Backdrop sx={{ zIndex: 9999, color: '#fff' }} open={!!isLoading}>
          <Grid container direction="column" alignItems="center">
            <CircularProgress color="inherit" />
            {loadingMessage && <Box m={4}>
              <Typography variant="caption">{loadingMessage}</Typography>
            </Box>}
          </Grid>
        </Backdrop> */}
      </ThemeProvider>
    </LocalizationProvider>
  </>
}
