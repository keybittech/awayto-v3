import React, { useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { keycloak, refreshToken, useAppSelector, useContexts, useAuth, useComponents } from 'awayto/hooks';

function AuthProvider(): React.JSX.Element {
  const location = useLocation();

  const { AuthContext } = useContexts();
  const { App } = useComponents();

  const [init, setInit] = useState(false);

  const [_token] = useState(localStorage.getItem('kc_token') as string);
  const [_refreshToken] = useState(localStorage.getItem('kc_refreshToken') as string);

  const { authenticated } = useAppSelector(state => state.auth);
  const { setAuthenticated } = useAuth();

  useEffect(() => {
    if (location.pathname == '/register') {
      void keycloak.init({}).then(() => {
        const redirectUri = window.location.toString().replace('/register', '')
        window.location.href = keycloak.createRegisterUrl({ redirectUri });
      }).catch(console.error);
    } else {
      void keycloak.init({
        onLoad: 'login-required',
        checkLoginIframe: false,
        token: _token,
        refreshToken: _refreshToken
      }).then(async (currentAuth) => {
        if (currentAuth) {
          setInterval(refreshToken, 55 * 1000);

          await refreshToken();
          // await fetch('/api/auth/checkin', {
          //   headers: {
          //     'Content-Type': 'application/json',
          //     'Authorization': `Bearer ${keycloak.token as string}`
          //   }
          // });
        }

        setAuthenticated({ authenticated: currentAuth });
        setInit(true);
      }).catch((err) => {
        console.log({ err: err as string });
      });
    }
  }, []);

  const authContext = {
    authenticated,
    keycloak,
    refreshToken
  } as AuthContextType;

  return useMemo(() => !AuthContext || !init ? <></> :
    <AuthContext.Provider value={authContext}>
      <App />
    </AuthContext.Provider>,
    [AuthContext, authContext, init]
  );
}

export default AuthProvider;
