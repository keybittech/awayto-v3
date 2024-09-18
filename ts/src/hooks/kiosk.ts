import { IGroup, ISchedule } from './api';
export type IKiosk = IGroup & {
  schedules: Record<string, ISchedule>;
  updatedOn: string;
}
