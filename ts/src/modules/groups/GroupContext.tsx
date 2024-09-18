import { createContext } from 'react';

import { IDefaultedComponent, IGroup, IGroupForm, IGroupRole, IGroupSchedule, IGroupService } from 'awayto/hooks';

declare global {
  type GroupContextType = {
    groups: IGroup[];
    group: IGroup;
    groupSchedules: IGroupSchedule[];
    groupServices: IGroupService[];
    groupForms: IGroupForm[];
    groupRoles: IGroupRole[];
    GroupSelect: IDefaultedComponent;
  }
}

export const GroupContext = createContext<GroupContextType | null>(null);

export default GroupContext;
