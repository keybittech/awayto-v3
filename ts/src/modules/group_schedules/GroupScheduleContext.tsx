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

export interface GroupScheduleContextType {
  getGroupSchedules: UseSiteQuery<GroupScheduleServiceGetGroupSchedulesApiArg, GroupScheduleServiceGetGroupSchedulesApiResponse>;
  getGroupUserScheduleStubs: UseSiteQuery<GroupUserScheduleServiceGetGroupUserScheduleStubsApiArg, GroupUserScheduleServiceGetGroupUserScheduleStubsApiResponse>;
  getGroupUserSchedules: UseSiteQuery<GroupUserScheduleServiceGetGroupUserSchedulesApiArg, GroupUserScheduleServiceGetGroupUserSchedulesApiResponse>;
  selectGroupSchedule: UseSelectOneResponse<IGroupSchedule>;
  selectGroupScheduleService: UseSelectOneResponse<IService>;
  selectGroupScheduleServiceTier: UseSelectOneResponse<IServiceTier>;
}

export const GroupScheduleContext = createContext<GroupScheduleContextType | null>(null);

export default GroupScheduleContext;

