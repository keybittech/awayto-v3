import { IBooking, IFile } from './api';
import { SocketMessage } from './web_socket';

export enum ExchangeActions {
  EXCHANGE_TEXT = 0,
  EXCHANGE_CALL = 1,
  EXCHANGE_WHITEBOARD = 2,
}

export interface DraggableBoxData {
  id: string;
  x: number;
  y: number;
  color: string;
  text: string;
}

/**
 * @category Exchange
 * @purpose tracks essential props between participants during whiteboard interactions
 */
export interface IWhiteboard {
  selectedText?: string;
  sharedFile?: IFile | null;
  lines: {
    startPoint: {
      x: number;
      y: number;
    };
    endPoint: {
      x: number;
      y: number;
    };
  }[];
  settings: Partial<{
    stroke: string;
    page: number;
    scale: number;
    highlight: boolean;
    position: number[];
  }>;
  boxes: DraggableBoxData[];
}

/**
 * @category Exchange
 * @purpose maps websocket responses which contain common WebRTC protocol objects
 */
export type ExchangeSessionAttributes = {
  sdp: RTCSessionDescriptionInit | null;
  ice: RTCIceCandidateInit;
  formats: string[];
  message: string;
  style: SocketMessage['style'];
  target: string;
};

/**
 * @category Exchange
 * @purpose contains Exchange participant WebRTC stream and connection objects
 */
export type Sender = {
  pc?: RTCPeerConnection;
  mediaStream?: MediaStream;
  peerResponse?: boolean;
}

/**
 * @category Exchange
 * @purpose tracks all existing participants in an ongoing Exchange
 */
export type SenderStreams = {
  [prop: string]: Sender
}

/**
 * @category Exchange
 * @purpose parent container for the Exchange UI where users chat, share documents, and participate in voice and video calls
 */
export type IExchange = {
  booking: IBooking;
};
