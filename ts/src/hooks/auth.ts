import { createSlice } from '@reduxjs/toolkit';

import Keycloak from 'keycloak-js';

import AccessibilityNewIcon from '@mui/icons-material/AccessibilityNew';
import LockResetIcon from '@mui/icons-material/LockReset';
import GroupsIcon from '@mui/icons-material/Groups';
import AssignmentTurnedInIcon from '@mui/icons-material/AssignmentTurnedIn';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import PermContactCalendarIcon from '@mui/icons-material/PermContactCalendar';
import BusinessIcon from '@mui/icons-material/Business';
import Diversity3Icon from '@mui/icons-material/Diversity3';
import LockIcon from '@mui/icons-material/Lock';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';

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
  UNRESTRICTED = 0,
  APP_ROLE_CALL = 1,
  APP_GROUP_ADMIN = 2,
  APP_GROUP_BOOKINGS = 4,
  APP_GROUP_SCHEDULES = 8,
  APP_GROUP_SERVICES = 16,
  APP_GROUP_SCHEDULE_KEYS = 32,
  APP_GROUP_ROLES = 64,
  APP_GROUP_USERS = 128,
  APP_GROUP_PERMISSIONS = 256,
}

export const SiteRoleDetails = {
  [SiteRoles.UNRESTRICTED]: {
    name: 'Unrestricted',
    description: '',
    resource: '',
    icon: AccessibilityNewIcon,
  },
  [SiteRoles.APP_ROLE_CALL]: {
    name: 'Role Call',
    description: 'Refetch roles',
    resource: '',
    icon: LockResetIcon,
  },
  [SiteRoles.APP_GROUP_ADMIN]: {
    name: 'Manage Group',
    description: 'Manage group',
    resource: '/group/manage',
    icon: AdminPanelSettingsIcon,
  },
  [SiteRoles.APP_GROUP_BOOKINGS]: {
    name: 'Requests',
    description: 'Request service',
    resource: '/request',
    icon: AssignmentTurnedInIcon,
  },
  [SiteRoles.APP_GROUP_SCHEDULES]: {
    name: 'Personal Schedule',
    description: 'Edit personal schedule',
    resource: '/schedule',
    icon: CalendarMonthIcon,
  },
  [SiteRoles.APP_GROUP_SERVICES]: {
    name: 'Group Services',
    description: 'Edit group services',
    resource: '/group/manage/services',
    icon: BusinessIcon,
  },
  [SiteRoles.APP_GROUP_SCHEDULE_KEYS]: {
    name: 'Group Schedules',
    description: 'Edit group schedules',
    resource: '/group/manage/schedules',
    icon: PermContactCalendarIcon,
  },
  [SiteRoles.APP_GROUP_ROLES]: {
    name: 'Group Roles',
    description: 'Edit group roles',
    resource: '/group/manage/roles',
    icon: GroupsIcon,
  },
  [SiteRoles.APP_GROUP_USERS]: {
    name: 'Group Users',
    description: 'Edit group users',
    resource: '/group/manage/users',
    icon: Diversity3Icon,
  },
  [SiteRoles.APP_GROUP_PERMISSIONS]: {
    name: 'Group Permissions',
    description: 'Edit group permissions',
    resource: '/group/manage/permissions',
    icon: LockIcon,
  },
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

export const hasRoleBits = function(userRoleBits: number, targetRoles: SiteRoles[]): boolean {
  if (!targetRoles || targetRoles.length === 0) return false;
  const targetBits = targetRoles.reduce((acc, role) => acc | role, 0);
  return (userRoleBits & targetBits) > 0;
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
  authenticated?: boolean;
}

export const authConfig = {
  name: 'auth',
  initialState: {
    authenticated: undefined
  } as IAuth,
  reducers: {
    setAuthenticated: (state: IAuth, action: { payload: IAuth }) => {
      state.authenticated = action.payload.authenticated;
    },
  },
};

export const authSlice = createSlice(authConfig);
