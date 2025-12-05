import { createSlice } from '@reduxjs/toolkit';

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

export type IAuth = {
  authenticated?: boolean;
  sessionId?: string;
  vaultKey?: string;
}

export async function logout() {
  try {
    sessionStorage.clear();
    localStorage.clear();
    window.location.href = '/auth/logout';
  } catch (error) {
    console.error('Logout failed:', error);
  }
}

export const authConfig = {
  name: 'auth',
  initialState: {
    authenticated: undefined,
    sessionId: '',
    vaultKey: ''
  } as IAuth,
  reducers: {
    setAuthenticated: (state: IAuth, action: { payload: IAuth }) => {
      state.authenticated = action.payload.authenticated;
    },
    setVault: (state: IAuth, action: { payload: IAuth }) => {
      const { vaultKey, sessionId } = action.payload;
      if (vaultKey) state.vaultKey = vaultKey;
      if (sessionId) state.sessionId = sessionId;
    },
  },
};

export const authSlice = createSlice(authConfig);

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

export const hasRoleBits = function(userRoleBits: number, targetRoles: SiteRoles[]): boolean {
  if (!targetRoles || targetRoles.length === 0) return false;
  const targetBits = targetRoles.reduce((acc, role) => acc | role, 0);
  return (userRoleBits & targetBits) > 0;
}
