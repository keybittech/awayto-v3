import React, { useMemo } from 'react';

import { siteApi, useComponents, useSelectOne } from 'awayto/hooks';

import GroupScheduleContext from './GroupScheduleContext';

export function GroupScheduleProvider({ children }: IComponent): React.JSX.Element {
  const { GroupScheduleSelectionProvider } = useComponents();

  const getGroupSchedules = siteApi.useGroupScheduleServiceGetGroupSchedulesQuery();
  const getGroupUserScheduleStubs = siteApi.useGroupUserScheduleServiceGetGroupUserScheduleStubsQuery();

  const selectGroupSchedule = useSelectOne('Schedule', {
    data: getGroupSchedules.data?.groupSchedules
  });

  const getGroupUserSchedules = siteApi.useGroupUserScheduleServiceGetGroupUserSchedulesQuery(
    { groupScheduleId: selectGroupSchedule.item?.schedule?.id || '' },
    { skip: !selectGroupSchedule.item?.schedule?.id }
  );

  const selectGroupScheduleService = useSelectOne('Service', {
    data: getGroupUserSchedules.data?.groupUserSchedules?.flatMap(gus => Object.values(gus.brackets || {}).flatMap(b => Object.values(b.services || {})))
  });

  const selectGroupScheduleServiceTier = useSelectOne('Tier', {
    data: Object.values(selectGroupScheduleService.item?.tiers || {}).sort((a, b) => new Date(a.createdOn || '').getTime() - new Date(b.createdOn || '').getTime())
  });

  const groupScheduleContext = {
    getGroupSchedules,
    getGroupUserScheduleStubs,
    getGroupUserSchedules,
    selectGroupSchedule,
    selectGroupScheduleService,
    selectGroupScheduleServiceTier
  } as GroupScheduleContextType;

  return useMemo(() => !GroupScheduleSelectionProvider ? <></> :
    <GroupScheduleContext.Provider value={groupScheduleContext}>
      <GroupScheduleSelectionProvider>
        {children}
      </GroupScheduleSelectionProvider>
    </GroupScheduleContext.Provider>,
    [GroupScheduleSelectionProvider, groupScheduleContext]
  );
}

export default GroupScheduleProvider;
