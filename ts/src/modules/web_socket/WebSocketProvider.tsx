import React, { useEffect, useMemo, useRef, useState } from 'react';

import { useUtil, SocketResponseHandler, siteApi, SocketActions, SocketResponse } from 'awayto/hooks';

import WebSocketContext, { WebSocketContextType } from './WebSocketContext';

const {
  VITE_REACT_APP_APP_HOST_NAME,
} = import.meta.env;

const defaultPadding = 5;

function paddedLen(len: number) {
  let strLen = len.toString();
  while (strLen.length < defaultPadding) strLen = "0" + strLen;
  return strLen;
}

function generateMessage(action: SocketActions, store: boolean, historical: boolean, timestamp: string, topic: string, cid: string, payload: string) {
  return [
    paddedLen(action.toString().length), action,
    paddedLen(1), String(store)[0],
    paddedLen(1), String(historical)[0],
    paddedLen(timestamp.length), timestamp,
    paddedLen(topic.length), topic,
    paddedLen(cid.length), cid,
    paddedLen(payload.length), payload,
  ].join('');
}

function parseField(cursor: number, data: string): [number, string, null | Error] {
  const lenEnd = cursor + defaultPadding;

  if (data.length < lenEnd) {
    return [0, "", new Error("length index out of range")];
  }

  const lenStr = data.substring(cursor, lenEnd);
  const valLen = parseInt(lenStr, 10);

  const valEnd = lenEnd + valLen;

  if (data.length < valEnd) {
    return [0, "", new Error("value index out of range")];
  }

  const val = data.substring(lenEnd, valEnd);

  return [valEnd, val, null];
}

function parseMessage(eventData: string) {
  const messageParams = Array(7);

  let cursor = 0;
  for (let i = 0; i < messageParams.length; i++) {
    const [newCursor, curr, err] = parseField(cursor, eventData);
    if (err) {
      console.error(err);
      return
    }
    messageParams[i] = curr;
    cursor = newCursor;
  }

  const socketResponse: SocketResponse<string> = {
    action: parseInt(messageParams[0]),
    store: messageParams[1] == 't',
    historical: messageParams[2] == 't',
    timestamp: messageParams[3] || (new Date()).toISOString(),
    topic: messageParams[4],
    sender: messageParams[5],
    payload: messageParams[6],
  };

  return socketResponse;
}

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
      const [_, connId] = ticket.split(":");

      const ws = new WebSocket(`wss://${VITE_REACT_APP_APP_HOST_NAME}/sock?ticket=${ticket}`)

      ws.onopen = () => {
        console.log('socket open', connId);
        if (reconnectSnackShown.current) {
          setSnack({ snackOn: 'Reconnected!', snackType: 'success' });
          reconnectSnackShown.current = false;
        }
        setConnectionId(connId);
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

        const socketResponse = parseMessage(event.data);

        if (!socketResponse) {
          return
        }

        if (socketResponse.payload == "PING") {
          ws.send(
            generateMessage(SocketActions.PING_PONG, false, false, "", "", "", "PONG")
          );
        } else if (SocketActions.ROLE_CALL == socketResponse.action) {
          await getUserProfileDetails();
        } else if (socketResponse.topic) {
          const listeners = messageListeners.current.get(socketResponse.topic);

          if (listeners) {
            for (const listener of listeners) {
              socketResponse.payload = JSON.parse(socketResponse.payload || '{}')
              await listener(socketResponse);
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
        socket.send(
          generateMessage(action, store, false, "", topic, connectionId, JSON.stringify(payload) || '')
        );
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
