import React, { useMemo } from 'react';

import { siteApi, useSelectOne } from 'awayto/hooks';

import GroupScheduleContext, { GroupScheduleContextType } from './GroupScheduleContext';
import GroupScheduleSelectionProvider from './GroupScheduleSelectionProvider';

export function GroupScheduleProvider({ children }: IComponent): React.JSX.Element {
  console.log('group schedule provider load');

  const getGroupSchedules = siteApi.useGroupScheduleServiceGetGroupSchedulesQuery();
  const getGroupUserScheduleStubs = siteApi.useGroupUserScheduleServiceGetGroupUserScheduleStubsQuery();

  const selectGroupSchedule = useSelectOne('Schedule', {
    data: getGroupSchedules.data?.groupSchedules
  });

  const getGroupUserSchedules = siteApi.useGroupUserScheduleServiceGetGroupUserSchedulesQuery(
    { groupScheduleId: selectGroupSchedule.item?.schedule?.id || '' },
    { skip: !selectGroupSchedule.item?.schedule?.id }
  );

  const loading = getGroupSchedules.isLoading || getGroupUserScheduleStubs.isLoading || getGroupUserSchedules.isLoading;

  const selectGroupScheduleService = useSelectOne('Service', {
    data: getGroupUserSchedules.data?.groupUserSchedules?.flatMap(gus => Object.values(gus.brackets || {}).flatMap(b => Object.values(b.services || {})))
  });

  const selectGroupScheduleServiceTier = useSelectOne('Tier', {
    data: Object.values(selectGroupScheduleService.item?.tiers || {}).sort((a, b) => new Date(a.createdOn || '').getTime() - new Date(b.createdOn || '').getTime())
  });

  const groupScheduleContext: GroupScheduleContextType = {
    getGroupSchedules,
    getGroupUserScheduleStubs,
    getGroupUserSchedules,
    selectGroupSchedule,
    selectGroupScheduleService,
    selectGroupScheduleServiceTier
  };

  return useMemo(() => loading ? <></> : <GroupScheduleContext.Provider value={groupScheduleContext}>
    <GroupScheduleSelectionProvider>
      {children}
    </GroupScheduleSelectionProvider>
  </GroupScheduleContext.Provider>, [loading, groupScheduleContext]);
}

export default GroupScheduleProvider;
