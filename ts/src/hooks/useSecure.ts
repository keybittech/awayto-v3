import { useCallback } from 'react';
import { hasRoleBits, SiteRoles } from './auth';
import { siteApi } from './api';

export function useSecure(): (targetRoles: SiteRoles[]) => boolean {
  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const hasRoleCb = useCallback((targetRoles: SiteRoles[]) => {
    if (profileRequest?.userProfile.roleBits) {
      return hasRoleBits(profileRequest?.userProfile.roleBits, targetRoles);
    }
    return false;
  }, [profileRequest?.userProfile]);

  return hasRoleCb;
}
