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

export const refreshToken = async (dur?: number) => {
  await keycloak.updateToken(dur || 50);
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
  UNRESTRICTED = 'UNRESTRICTED',
  APP_ROLE_CALL = 'APP_ROLE_CALL',
  APP_GROUP_ADMIN = 'APP_GROUP_ADMIN',
  APP_GROUP_BOOKINGS = 'APP_GROUP_BOOKINGS',
  APP_GROUP_SCHEDULES = 'APP_GROUP_SCHEDULES',
  APP_GROUP_SERVICES = 'APP_GROUP_SERVICES',
  APP_GROUP_SCHEDULE_KEYS = 'APP_GROUP_SCHEDULE_KEYS',
  APP_GROUP_ROLES = 'APP_GROUP_ROLES',
  APP_GROUP_USERS = 'APP_GROUP_USERS',
  APP_GROUP_PERMISSIONS = 'APP_GROUP_PERMISSIONS',
  // APP_GROUP_FEATURES = 'APP_GROUP_FEATURES',
}

export const SiteRoleDetails = {
  [SiteRoles.UNRESTRICTED]: {
    name: 'Unrestricted',
    description: '',
    resource: ''
  },
  [SiteRoles.APP_ROLE_CALL]: {
    name: 'Role Call',
    description: 'Refetch roles',
    resource: ''
  },
  [SiteRoles.APP_GROUP_ADMIN]: {
    name: 'Admin',
    description: 'Manage group',
    resource: '/group/manage'
  },
  [SiteRoles.APP_GROUP_BOOKINGS]: {
    name: 'Requests',
    description: 'Request a service',
    resource: '/request',
  },
  [SiteRoles.APP_GROUP_SCHEDULES]: {
    name: 'Personal Schedule',
    description: 'Edit personal schedule',
    resource: '/schedule',
  },
  [SiteRoles.APP_GROUP_SERVICES]: {
    name: 'Group Services',
    description: 'Edit group services',
    resource: '/group/manage/services',
  },
  [SiteRoles.APP_GROUP_SCHEDULE_KEYS]: {
    name: 'Group Schedules',
    description: 'Edit group schedules',
    resource: '/group/manage/schedules',
  },
  [SiteRoles.APP_GROUP_ROLES]: {
    name: 'Group Roles',
    description: 'Edit group roles',
    resource: '/group/manage/roles'
  },
  [SiteRoles.APP_GROUP_USERS]: {
    name: 'Group Users',
    description: 'Edit group users',
    resource: '/group/manage/users',
  },
  [SiteRoles.APP_GROUP_PERMISSIONS]: {
    name: 'Group Permissions',
    description: 'Edit group permissions',
    resource: '/group/manage/permissions',
  },
  // [SiteRoles.APP_GROUP_FEATURES]: {
  //   name: 'Features',
  //   description: 'Edit group service features',
  //   resource: '/group/manage/services',
  // },
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
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${keycloak.token}`
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

export const authSlice = createSlice(authConfig);
