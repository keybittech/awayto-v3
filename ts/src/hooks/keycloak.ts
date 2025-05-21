import Keycloak from 'keycloak-js';

const {
  VITE_REACT_APP_KC_REALM,
  VITE_REACT_APP_KC_CLIENT,
  VITE_REACT_APP_KC_PATH,
  VITE_REACT_APP_APP_HOST_URL,
} = import.meta.env;

export const keycloak = new Keycloak({
  url: VITE_REACT_APP_KC_PATH,
  realm: VITE_REACT_APP_KC_REALM,
  clientId: VITE_REACT_APP_KC_CLIENT
});

export const refreshToken = async (dur?: number) => {
  await keycloak.updateToken(dur);
  return true;
}

export const setAuthHeaders = (headers?: Headers) => {
  if (!headers) {
    headers = new Headers();
  }
  headers.set('X-Tz', Intl.DateTimeFormat().resolvedOptions().timeZone);
  headers.set('Authorization', 'Bearer ' + keycloak.token);
  return headers;
}

export const login = async () => {
  const authenticated = await keycloak.init({ onLoad: 'login-required' });
  if (authenticated) {
    await fetch('/login', { headers: setAuthHeaders() });
  }
  return authenticated;
}

export const logout = async () => {
  localStorage.clear();
  await fetch('/logout', { headers: setAuthHeaders() });
  await keycloak.logout({ redirectUri: VITE_REACT_APP_APP_HOST_URL });
}

export const getTokenHeaders = function(): { headers: Record<string, string> } {
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${keycloak.token}`
    }
  }
}
