import React, { useState, useEffect, useMemo, useRef, useCallback } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';

import { SocketMessage, SocketActions, plural, useWebSocketSubscribe } from 'awayto/hooks';

import WSTextContext, { WSTextContextType } from './WSTextContext';
import GroupedMessages from './GroupedMessages';
import SubmitMessageForm from './SubmitMessageForm';

interface WSTextProviderProps extends IComponent {
  topicId: string;
  topicMessages: SocketMessage[];
  setTopicMessages: React.Dispatch<React.SetStateAction<SocketMessage[]>>;
}

export function WSTextProvider({ children, topicId, topicMessages, setTopicMessages }: WSTextProviderProps): React.JSX.Element {
  if (!topicId) return <></>;

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const lastMessage = useRef<string>(undefined);

  const [hasMore, setHasMore] = useState(false);
  const [page, setPage] = useState(1);

  const {
    userList,
    connectionId,
    connected,
    sendMessage,
    storeMessage
  } = useWebSocketSubscribe<{ page: number, pageSize: number, message: string, style: SocketMessage['style'] }>(topicId, ({ timestamp, sender, action, payload, historical }) => {

    if (action == SocketActions.HAS_MORE_MESSAGES) {
      setHasMore(true)
      return
    }

    const { message, style } = payload;

    if (message && style && setTopicMessages) {
      for (const user of Object.values(userList)) {
        if (user?.cids.includes(sender) || historical) {
          setTopicMessages(m => {
            const newMessage = {
              ...user,
              sender,
              style,
              message,
              timestamp
            }

            if (historical) {
              return [newMessage, ...m];
            } else {
              return [...m, newMessage];
            }
          })
        };
      }
    }
  });

  const getMore = useCallback(() => {
    const nextPage = page + 1;
    sendMessage(SocketActions.LOAD_MESSAGES, { page: nextPage, pageSize: 10 });
    setPage(nextPage);
    setHasMore(false);
  }, [page]);

  useEffect(() => {
    if (topicMessages && topicMessages.length) {
      const lt = topicMessages[topicMessages.length - 1].timestamp;
      if (lastMessage.current != lt) {
        lastMessage.current = lt
        messagesEndRef.current?.scrollIntoView({ behavior: 'auto', block: 'end', inline: 'nearest' })
      }
    }
  }, [messagesEndRef.current, topicMessages]);

  const wsTextContext: WSTextContextType = {
    wsTextConnectionId: connectionId,
    wsTextConnected: connected,
    messagesEnd: useMemo(() => <Box ref={messagesEndRef} />, []),
    chatLog: useMemo(() => <>
      {!topicMessages?.length && <Card sx={{ marginBottom: '8px' }}>
        <CardHeader subheader="Welcome to the chat! Messages will appear below..." />
      </Card>}
      {hasMore && <Box onClick={getMore} sx={{ flex: 1, textAlign: 'center', cursor: 'pointer', }}>
        Load previous...
      </Box>}
      <GroupedMessages topicMessages={topicMessages} />
    </>, [topicMessages, hasMore, getMore]),
    submitMessageForm: useMemo(() => {
      const userListNames = Object.keys(userList).length ? Object.values(userList).filter(u => u?.online).map(u => u?.name || '') : []
      return <>
        <SubmitMessageForm
          sendTextMessage={(message: string) => {
            storeMessage(SocketActions.TEXT, { style: 'written', message });
          }}
        />
        <Typography variant="caption">
          {userListNames.length && `${plural(userListNames.length, 'participant', 'participants')}: ${userListNames.join(', ')}`}
        </Typography>
      </>
    }, [userList])
  };

  return useMemo(() => <WSTextContext.Provider value={wsTextContext}>
    {children}
  </WSTextContext.Provider>, [wsTextContext]);
}

export default WSTextProvider;
