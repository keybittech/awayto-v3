import React, { createContext } from 'react';

export interface WSTextContextType {
  wsTextConnectionId: string;
  wsTextConnected: boolean;
  chatLog: React.JSX.Element;
  messagesEnd: React.JSX.Element;
  submitMessageForm: React.JSX.Element;
}

export const WSTextContext = createContext<WSTextContextType | null>(null);

export default WSTextContext;
