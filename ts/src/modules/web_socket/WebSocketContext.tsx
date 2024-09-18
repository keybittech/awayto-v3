import { createContext } from 'react';
import { SocketActions, SocketResponseHandler } from 'awayto/hooks';

declare global {
  type WebSocketContextType = {
    connectionId: string;
    connected: boolean;
    transmit: (store: boolean, action: SocketActions, topic: string, payload?: Partial<unknown>) => void;
    subscribe: <T>(topic: string, callback: SocketResponseHandler<T>) => () => void;
  }
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);

export default WebSocketContext;
