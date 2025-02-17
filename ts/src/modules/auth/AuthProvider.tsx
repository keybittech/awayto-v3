import React, { startTransition, useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { keycloak, authSlice, useAppDispatch, siteApi, SiteRoles } from 'awayto/hooks';

import AuthContext from './AuthContext';
import App from '../common/App';

const {
  REACT_APP_KC_CLIENT,
} = process.env as { [prop: string]: string };

function AuthProvider(): React.JSX.Element {

  const dispatch = useAppDispatch();
  const location = useLocation();

  const [init, setInit] = useState(false);

  const [getUserProfileDetails] = siteApi.useLazyUserProfileServiceGetUserProfileDetailsQuery();

  useEffect(() => {
    startTransition(async () => {
      if (location.pathname == '/join' && location.search.includes('groupCode')) {
        void keycloak.init({}).then(async () => {
          const redirectUri = window.location.toString().split('?')[0].replace('/join', '');
          const kcRegisterUrl = await keycloak.createRegisterUrl({ redirectUri });
          window.location.href = kcRegisterUrl + '&' + location.search.slice(1, location.search.length);
        }).catch(console.error);
      } else if (location.pathname == '/register') {
        void keycloak.init({}).then(async () => {
          const redirectUri = window.location.toString().replace('/register', '')
          window.location.href = await keycloak.createRegisterUrl({ redirectUri });
        }).catch(console.error);
      } else {
        try {
          const authenticated = await keycloak.init({ onLoad: 'login-required' });
          dispatch(authSlice.actions.setAuthenticated({ authenticated }));
          setInit(true);
        } catch (e) {
          console.log(e);
        }
      }
    })
  }, []);

  return useMemo(() => !init ? <></> :
    <AuthContext.Provider value={{}}>
      <App />
    </AuthContext.Provider>, [init]);
}

export default AuthProvider;
