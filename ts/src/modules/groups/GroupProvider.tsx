import React, { useMemo } from 'react';

import { siteApi, useContexts, useSelectOne, isExternal, IGroup } from 'awayto/hooks';

export function GroupProvider({ children }: IComponent): React.JSX.Element {
  const { GroupContext } = useContexts();

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

  return useMemo(() => !GroupContext ? <></> :
    <GroupContext.Provider value={groupContext}>
      {children}
    </GroupContext.Provider>,
    [GroupContext, groupContext]
  );
}

export default GroupProvider;
