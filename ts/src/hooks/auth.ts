import { createSlice } from '@reduxjs/toolkit';
import Keycloak from 'keycloak-js';

const {
  REACT_APP_KC_REALM,
  REACT_APP_KC_CLIENT,
  REACT_APP_KC_PATH
} = process.env as { [prop: string]: string };

export const keycloak = new Keycloak({
  url: REACT_APP_KC_PATH,
  realm: REACT_APP_KC_REALM,
  clientId: REACT_APP_KC_CLIENT
});

export const refreshToken = async () => {
  await keycloak.updateToken(-1);
  localStorage.setItem('kc_token', keycloak.token as string);
  localStorage.setItem('kc_refreshToken', keycloak.refreshToken as string);
  return true;
}

/**
 * @category Authorization
 */
export type KcSiteOpts = {
  regroup: (groupId?: string) => Promise<void>;
}

/**
 * @category Authorization
 */
export type DecodedJWTToken = {
  resource_access: {
    [prop: string]: { roles: string[] }
  },
  groups: string[]
}

/**
 * @category Authorization
 */
export enum SiteRoles {
  APP_ROLE_CALL = 'APP_ROLE_CALL',
  APP_GROUP_ADMIN = 'APP_GROUP_ADMIN',
  APP_GROUP_ROLES = 'APP_GROUP_ROLES',
  APP_GROUP_USERS = 'APP_GROUP_USERS',
  // APP_GROUP_MATRIX = 'APP_GROUP_MATRIX',
  APP_GROUP_SERVICES = 'APP_GROUP_SERVICES',
  APP_GROUP_BOOKINGS = 'APP_GROUP_BOOKINGS',
  APP_GROUP_FEATURES = 'APP_GROUP_FEATURES',
  APP_GROUP_SCHEDULES = 'APP_GROUP_SCHEDULES'
}

/**
 * @category Authorization
 */
export type StrategyUser = {
  sub: string;
}

/**
 * @category Authorization
 */

export const hasRole = function(availableUserGroupRoles?: string[], targetRoles?: string[]): boolean {
  if (!targetRoles || !availableUserGroupRoles) return false;
  return availableUserGroupRoles.some(gr => targetRoles.includes(gr))
}

/**
 * @category Authorization
 */
export const getTokenHeaders = function(): { headers: Record<string, string> } {
  const token = localStorage.getItem('kc_token');
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token as string}`
    }
  }
}

export type IAuth = {
  authenticated: boolean;
}

export const authConfig = {
  name: 'auth',
  initialState: {
    authenticated: false
  } as IAuth,
  reducers: {
    setAuthenticated: (state: IAuth, action: { payload: IAuth }) => {
      state.authenticated = action.payload.authenticated;
    },
  },
};

export const newAuthSlice = createSlice(authConfig);
