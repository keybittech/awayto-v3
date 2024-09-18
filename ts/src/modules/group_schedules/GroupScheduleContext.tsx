import { createContext } from 'react';
import {
  UseSelectOneResponse,
  IGroupSchedule,
  IService,
  IServiceTier,
  GroupScheduleServiceGetGroupSchedulesApiArg,
  GroupScheduleServiceGetGroupSchedulesApiResponse,
  UseSiteQuery,
  GroupUserScheduleServiceGetGroupUserScheduleStubsApiArg,
  GroupUserScheduleServiceGetGroupUserScheduleStubsApiResponse,
  GroupUserScheduleServiceGetGroupUserSchedulesApiArg,
  GroupUserScheduleServiceGetGroupUserSchedulesApiResponse
} from 'awayto/hooks';

declare global {
  type GroupScheduleContextType = {
    getGroupSchedules: UseSiteQuery<GroupScheduleServiceGetGroupSchedulesApiArg, GroupScheduleServiceGetGroupSchedulesApiResponse>;
    getGroupUserScheduleStubs: UseSiteQuery<GroupUserScheduleServiceGetGroupUserScheduleStubsApiArg, GroupUserScheduleServiceGetGroupUserScheduleStubsApiResponse>;
    getGroupUserSchedules: UseSiteQuery<GroupUserScheduleServiceGetGroupUserSchedulesApiArg, GroupUserScheduleServiceGetGroupUserSchedulesApiResponse>;
    selectGroupSchedule: UseSelectOneResponse<Required<IGroupSchedule>>;
    selectGroupScheduleService: UseSelectOneResponse<Required<IService>>;
    selectGroupScheduleServiceTier: UseSelectOneResponse<Required<IServiceTier>>;
  }
}

export const GroupScheduleContext = createContext<GroupScheduleContextType | null>(null);

export default GroupScheduleContext;

