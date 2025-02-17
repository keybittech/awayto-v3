import React, { Suspense } from 'react';

import { useComponents } from 'awayto/hooks';

export function TopLevelProviders({ children }: IComponent): React.JSX.Element {
  const { BookingProvider, GroupProvider, GroupScheduleProvider, WebSocketProvider } = useComponents();
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
  </Suspense>
}

export default TopLevelProviders;
