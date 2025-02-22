import React, { useEffect, useMemo, useRef, useState } from 'react';

import { useUtil, SocketResponse, SocketResponseHandler, siteApi, SocketActions } from 'awayto/hooks';

import WebSocketContext, { WebSocketContextType } from './WebSocketContext';

const {
  VITE_REACT_APP_APP_HOST_NAME,
} = import.meta.env;

function WebSocketProvider({ children }: IComponent): React.JSX.Element {

  const [getTicket] = siteApi.useLazySockServiceGetSocketTicketQuery();
  const [getUserProfileDetails] = siteApi.useLazyUserProfileServiceGetUserProfileDetailsQuery();

  const { setSnack } = useUtil();

  const [socket, setSocket] = useState<WebSocket | undefined>();
  const [connectionId, setConnectionId] = useState('');
  const reconnectSnackShown = useRef(false);
  const initialConnectionMade = useRef(false);

  const messageListeners = useRef(new Map<string, Set<SocketResponseHandler<unknown>>>());

  const connect = () => {

    void getTicket().unwrap().then(async res => {
      const { ticket } = res;
      const [_, connectionId] = ticket.split(":");

      const ws = new WebSocket(`wss://${VITE_REACT_APP_APP_HOST_NAME}/sock?ticket=${ticket}`)

      ws.onopen = () => {
        console.log('socket open', connectionId);
        if (reconnectSnackShown.current) {
          setSnack({ snackOn: 'Reconnected!', snackType: 'success' });
          reconnectSnackShown.current = false;
        }
        setConnectionId(connectionId);
        setSocket(ws);
        initialConnectionMade.current = true;
      };

      ws.onclose = () => {
        console.log('socket closed. reconnecting...');
        if (!reconnectSnackShown.current) {
          setSnack({ snackOn: 'Bad connection. Attempting to reconnect.', snackType: 'info' });
          reconnectSnackShown.current = true;
        }
        setTimeout(() => {
          connect();
        }, 5000);
      };

      ws.onerror = (error) => {
        console.error("socket error:", error);
        ws.close();
      };

      ws.onmessage = async (event) => {
        const {
          timestamp,
          sender,
          action,
          topic,
          payload,
          historical
        } = JSON.parse(event.data) as SocketResponse<string>;


        if (payload == "PING") {
          ws.send(JSON.stringify({ sender: connectionId, payload: "PONG" }));
        } else if (SocketActions.ROLE_CALL == action) {
          await getUserProfileDetails();
        } else {
          const listeners = messageListeners.current.get(topic);

          if (listeners) {
            for (const listener of listeners) {
              await listener({ timestamp, sender, action, topic, payload: payload || {}, historical });
            }
          }
        }
      };
    }).catch(() => {
      setSnack({ snackOn: 'Connection lost. Please refresh the page.', snackType: 'warning' });
      reconnectSnackShown.current = true;
    });
  }

  useEffect(() => {
    connect();
  }, []);

  const webSocketContext: WebSocketContextType = {
    connectionId,
    connected: socket?.readyState === WebSocket.OPEN,
    transmit(store, action, topic, payload) {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          timestamp: (new Date()).toUTCString(),
          sender: connectionId,
          store,
          action,
          actionName: SocketActions[action],
          topic,
          payload
        }));
      }
    },
    subscribe(topic, callback) {
      const listeners = messageListeners.current.get(topic) || new Set();
      listeners.add(callback);
      messageListeners.current.set(topic, listeners);

      return () => {
        listeners.delete(callback);
        if (listeners.size === 0) {
          messageListeners.current.delete(topic);
        }
      };
    },
  };

  return useMemo(() => !initialConnectionMade.current ? <></> :
    <WebSocketContext.Provider value={webSocketContext}>
      {children}
    </WebSocketContext.Provider>,
    [initialConnectionMade.current, webSocketContext]
  );
}

export default WebSocketProvider;
