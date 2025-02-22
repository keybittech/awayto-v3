import { createContext } from 'react';

import { IGroup, IGroupForm, IGroupRole, IGroupSchedule, IGroupService } from 'awayto/hooks';

export interface GroupContextType {
  groups: IGroup[];
  group: IGroup;
  groupSchedules: IGroupSchedule[];
  groupServices: IGroupService[];
  groupForms: IGroupForm[];
  groupRoles: IGroupRole[];
  GroupSelect: (p: IComponent) => React.JSX.Element;
}

export const GroupContext = createContext<GroupContextType | null>(null);

export default GroupContext;
