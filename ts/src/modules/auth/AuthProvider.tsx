import React, { useEffect, useMemo, useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { keycloak, refreshToken, useAppSelector, useAuth, useComponents } from 'awayto/hooks';

import AuthContext from './AuthContext';

function AuthProvider(): React.JSX.Element {
  const location = useLocation();

  const { App } = useComponents();

  const [init, setInit] = useState(false);
  const { authenticated } = useAppSelector(state => state.auth);
  const { setAuthenticated } = useAuth();

  useEffect(() => {
    async function go() {
      if (location.pathname == '/join' && location.search.includes('groupCode')) {
        void keycloak.init({}).then(() => {
          const redirectUri = window.location.toString().split('?')[0].replace('/join', '');
          window.location.href = keycloak.createRegisterUrl({ redirectUri }) + '&' + location.search.substr(1);
        }).catch(console.error);
      } else if (location.pathname == '/register') {
        void keycloak.init({}).then(async () => {
          const redirectUri = window.location.toString().replace('/register', '')
          window.location.href = await keycloak.createRegisterUrl({ redirectUri });
        }).catch(console.error);
      } else {
        try {
          const authenticated = await keycloak.init({
            onLoad: 'login-required',
          });

          setInterval(refreshToken, 55 * 1000);
          setAuthenticated({ authenticated });
          setInit(true);
        } catch (e) {
          console.log(e);
        }
      }
    }
    void go();
  }, []);

  const authContext = {
    authenticated,
    keycloak
  } as AuthContextType;

  return useMemo(() => !init ? <></> :
    <AuthContext.Provider value={authContext}>
      <App />
    </AuthContext.Provider>, [authContext, init]);
}

export default AuthProvider;
