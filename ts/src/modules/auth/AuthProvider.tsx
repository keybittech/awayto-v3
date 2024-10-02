import React, { useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { keycloak, refreshToken, useAppSelector, useContexts, useAuth, useComponents } from 'awayto/hooks';

function AuthProvider(): React.JSX.Element {
  const location = useLocation();

  const { AuthContext } = useContexts();
  const { App } = useComponents();

  const [init, setInit] = useState(false);
  const { authenticated } = useAppSelector(state => state.auth);
  const { setAuthenticated } = useAuth();

  useEffect(() => {
    async function go() {
      if (location.pathname == '/register') {
        void keycloak.init({}).then(() => {
          const redirectUri = window.location.toString().replace('/register', '')
          window.location.href = keycloak.createRegisterUrl({ redirectUri });
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
