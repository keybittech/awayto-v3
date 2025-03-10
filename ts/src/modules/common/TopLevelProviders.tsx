import React, { Suspense } from 'react';
import WebSocketProvider from '../web_socket/WebSocketProvider';
import BookingProvider from '../bookings/BookingProvider';
import GroupProvider from '../groups/GroupProvider';
import GroupScheduleProvider from '../group_schedules/GroupScheduleProvider';

export function TopLevelProviders({ children }: IComponent): React.JSX.Element {
  return <Suspense>
    <WebSocketProvider>
      <BookingProvider>
        <GroupProvider>
          <GroupScheduleProvider>
            {children}
          </GroupScheduleProvider>
        </GroupProvider>
      </BookingProvider>
    </WebSocketProvider>
  </Suspense>;
}

export default TopLevelProviders;
