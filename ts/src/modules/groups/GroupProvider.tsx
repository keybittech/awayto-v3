import React, { useMemo } from 'react';

import { siteApi, useSelectOne, IGroup } from 'awayto/hooks';

import GroupContext, { GroupContextType } from './GroupContext';

export function GroupProvider({ children }: IComponent): React.JSX.Element {

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const groups = useMemo(() => Object.values(profileRequest?.userProfile?.groups || {}) as Required<IGroup>[], [profileRequest?.userProfile.groups]);

  const { item: group, comp: GroupSelect } = useSelectOne<IGroup>('Groups', { data: groups });

  const { data: groupSchedulesRequest } = siteApi.useGroupScheduleServiceGetGroupSchedulesQuery();
  const { data: groupServicesRequest } = siteApi.useGroupServiceServiceGetGroupServicesQuery();
  const { data: groupFormsRequest } = siteApi.useGroupFormServiceGetGroupFormsQuery();
  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();

  const groupContext: GroupContextType = {
    groups,
    group: group || {},
    groupSchedules: groupSchedulesRequest?.groupSchedules || [],
    groupServices: groupServicesRequest?.groupServices || [],
    groupForms: groupFormsRequest?.groupForms || [],
    groupRoles: groupRolesRequest?.groupRoles || [],
    GroupSelect
  };

  return useMemo(() => <GroupContext.Provider value={groupContext}>
    {children}
  </GroupContext.Provider>, [groupContext]);
}

export default GroupProvider;
