export enum SocketActions {
  PING_CHANNEL = 0,
  START_STREAM = 1,
  STOP_STREAM = 2,
  STREAM_INQUIRY = 3,
  UNSUBSCRIBE_TOPIC = 4,
  SUBSCRIBE_TOPIC = 5,
  LOAD_MESSAGES = 6,
  HAS_MORE_MESSAGES = 7,
  SUBSCRIBE = 8,
  UNSUBSCRIBE = 9,
  LOAD_SUBSCRIBERS = 10,
  SUBSCRIBERS_PRESENT = 11,
  TEXT = 12,
  RTC = 13,
  SET_POSITION = 14,
  SET_PAGE = 15,
  SET_SCALE = 16,
  SET_STROKE = 17,
  DRAW_LINES = 18,
  SHARE_FILE = 19,
  CHANGE_SETTING = 20,
  SUBSCRIBE_INIT = 21,
  SET_SELECTED_TEXT = 22,
}

/**
 * @category Web Socket
 * @purpose the form of a socket response
 */
export type SocketResponse<T> = {
  store?: boolean;
  sender: string;
  action: SocketActions;
  topic: string;
  timestamp: string;
  payload: Partial<T>;
  historical?: boolean;
};

/**
 * @category Web Socket
 * @purpose handles topic listener results
 */
export type SocketResponseHandler<T> = (response: SocketResponse<T>) => void | Promise<void>;

/**
 * @category Web Socket
 * @purpose participant object based off anon socket connections
 */
export type SocketParticipant = {
  scid: string; // sock_connection id
  cids: string[]; // connection_id
  name: string;
  role: string;
  color: string;
  exists: boolean;
  online: boolean;
}

/**
 * @category Web Socket
 * @purpose provides structure to chat messages during interactions
 */
export type SocketMessage = SocketParticipant & {
  style: 'utterance' | 'written';
  action?: () => React.JSX.Element;
  sender: string;
  message: string;
  timestamp: string;
};
