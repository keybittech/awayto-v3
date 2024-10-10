import React, { useMemo } from 'react';

import { siteApi, useSelectOne, isExternal, IGroup } from 'awayto/hooks';

import GroupContext from './GroupContext';

export function GroupProvider({ children }: IComponent): React.JSX.Element {

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery(undefined, { skip: isExternal(window.location.pathname) });

  const groups = useMemo(() => Object.values(profileRequest?.userProfile?.groups || {}) as Required<IGroup>[], [profileRequest?.userProfile.groups]);

  const { item: group, comp: GroupSelect } = useSelectOne<IGroup>('Groups', { data: groups });

  const { data: groupSchedulesRequest } = siteApi.useGroupScheduleServiceGetGroupSchedulesQuery();
  const { data: groupServicesRequest } = siteApi.useGroupServiceServiceGetGroupServicesQuery();
  const { data: groupFormsRequest } = siteApi.useGroupFormServiceGetGroupFormsQuery();
  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();

  const groupContext = {
    groups,
    group,
    groupSchedules: groupSchedulesRequest?.groupSchedules,
    groupServices: groupServicesRequest?.groupServices,
    groupForms: groupFormsRequest?.groupForms,
    groupRoles: groupRolesRequest?.groupRoles,
    GroupSelect
  } as GroupContextType | null;

  return useMemo(() => <GroupContext.Provider value={groupContext}>
    {children}
  </GroupContext.Provider>, [groupContext]);
}

export default GroupProvider;
