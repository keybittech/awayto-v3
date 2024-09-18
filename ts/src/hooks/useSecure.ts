import { useCallback } from 'react';
import { hasRole, SiteRoles } from './auth';
import { siteApi } from './api';

export function useSecure(): (targetRoles: SiteRoles[]) => boolean {
  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const hasRoleCb = useCallback((targetRoles: SiteRoles[]) => {
    if (profileRequest?.userProfile) {
      return hasRole(profileRequest?.userProfile.availableUserGroupRoles, targetRoles);
    }
    return false;
  }, [profileRequest?.userProfile]);

  return hasRoleCb;
}
