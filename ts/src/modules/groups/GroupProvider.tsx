import React, { useMemo } from 'react';

import { siteApi, useSelectOne, IGroup } from 'awayto/hooks';

import GroupContext, { GroupContextType } from './GroupContext';

export function GroupProvider({ children }: IComponent): React.JSX.Element {
  console.log('group provider load');

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const groups = useMemo(() => Object.values(profileRequest?.userProfile?.groups || {}) as Required<IGroup>[], [profileRequest?.userProfile.groups]);

  const { item: group, comp: GroupSelect } = useSelectOne<IGroup>('Groups', { data: groups });

  const { data: groupSchedulesRequest, isLoading: l1 } = siteApi.useGroupScheduleServiceGetGroupSchedulesQuery();
  const { data: groupServicesRequest, isLoading: l2 } = siteApi.useGroupServiceServiceGetGroupServicesQuery();
  const { data: groupFormsRequest, isLoading: l3 } = siteApi.useGroupFormServiceGetGroupFormsQuery();
  const { data: groupRolesRequest, isLoading: l4 } = siteApi.useGroupRoleServiceGetGroupRolesQuery();

  const loading = l1 || l2 || l3 || l4;

  const groupContext: GroupContextType = {
    groups,
    group: group || {},
    groupSchedules: groupSchedulesRequest?.groupSchedules || [],
    groupServices: groupServicesRequest?.groupServices || [],
    groupForms: groupFormsRequest?.groupForms || [],
    groupRoles: groupRolesRequest?.groupRoles || [],
    GroupSelect
  };

  return useMemo(() => loading ? <></> : <GroupContext.Provider value={groupContext}>
    {children}
  </GroupContext.Provider>, [loading, groupContext]);
}

export default GroupProvider;
