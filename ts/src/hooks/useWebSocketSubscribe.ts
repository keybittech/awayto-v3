// useWebSocket.js
import { useContext, useCallback, useEffect, useMemo, useState } from 'react';
import { generateLightBgColor } from './util';
import { SocketActions, SocketParticipant, SocketResponseHandler } from './web_socket';

import WebSocketContext, { WebSocketContextType } from '../modules/web_socket/WebSocketContext';

export function useWebSocketSend() {
  const context = useContext(WebSocketContext) as WebSocketContextType;
  return context.transmit;
}

export function useWebSocketSubscribe<T>(topic: string, callback: SocketResponseHandler<T>) {

  const {
    connectionId,
    connected,
    transmit,
    subscribe
  } = useContext(WebSocketContext) as WebSocketContextType;

  const [subscribed, setSubscribed] = useState(false);

  const [userList, setUserList] = useState<Record<string, SocketParticipant>>({});

  const callbackRef = useCallback(callback, [callback]);

  useEffect(() => {
    if (subscribed) {
    }
  }, [subscribed]);

  useEffect(() => {
    if (connected) {
      const unsubscribe = subscribe(topic, async (message) => {
        console.log("Current action: ", message.action, SocketActions[message.action])

        if (SocketActions.SUBSCRIBE == message.action) {
          setSubscribed(true);

          transmit(false, SocketActions.LOAD_SUBSCRIBERS, topic);
        } else if (SocketActions.LOAD_SUBSCRIBERS == message.action) {

          const socketParticipants = message.payload as Record<string, SocketParticipant>;

          setUserList(ul => {
            for (const participant of Object.values(socketParticipants)) {
              const sub = ul[participant.scid];
              if (sub) {
                sub.cids = Array.from(new Set([...sub.cids, ...participant.cids]))
                sub.online = participant.online
                ul[sub.scid] = sub;
              } else {
                participant.color = generateLightBgColor();
                ul[participant.scid] = participant;
              }
            }
            return { ...ul };
          })

          if (message.sender == connectionId) {
            transmit(false, SocketActions.LOAD_MESSAGES, topic, { page: 1, pageSize: 10 });
          }
        } else if (SocketActions.UNSUBSCRIBE_TOPIC === message.action) {

          const [scid] = (message.payload as string).split(':');

          setUserList(ul => {
            ul[scid].online = false;
            return { ...ul };
          })
        } else {
          await callbackRef(message);
        }
      });

      transmit(false, SocketActions.SUBSCRIBE, topic);

      return () => {
        transmit(false, SocketActions.UNSUBSCRIBE, topic);

        unsubscribe();
      };
    }
  }, [transmit, connected, topic]);

  return useMemo(() => ({
    userList,
    subscribed,
    connectionId,
    connected,
    storeMessage: (action: SocketActions, payload?: Partial<T>) => {
      if (connected) {
        transmit(true, action, topic, payload);
      }
    },
    sendMessage: (action: SocketActions, payload?: Partial<T>) => {
      if (connected) {
        transmit(false, action, topic, payload);
      }
    }
  }), [connectionId, connected, subscribed, userList]);
}
