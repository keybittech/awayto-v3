import React, { useMemo, useState } from 'react';
import { useParams } from 'react-router';

import { ExchangeActions, SocketMessage, siteApi } from 'awayto/hooks';

import ExchangeContext, { ExchangeContextType } from './ExchangeContext';
import WSTextProvider from '../web_socket/WSTextProvider';
import WSCallProvider from '../web_socket/WSCallProvider';

export function ExchangeProvider({ children }: IComponent): React.JSX.Element {

  const { exchangeId } = useParams();
  if (!exchangeId) return <></>;

  const [topicMessages, setTopicMessages] = useState<SocketMessage[]>([]);

  const exchangeContext: ExchangeContextType = {
    exchangeId,
    topicMessages,
    setTopicMessages,
    getBookingFiles: siteApi.useBookingServiceGetBookingFilesQuery({ id: exchangeId })
  };

  return useMemo(() => <ExchangeContext.Provider value={exchangeContext}>
    <WSTextProvider
      topicId={`exchange/${ExchangeActions.EXCHANGE_TEXT}:${exchangeContext.exchangeId}`}
      topicMessages={topicMessages}
      setTopicMessages={setTopicMessages}
    >
      <WSCallProvider
        topicId={`exchange/${ExchangeActions.EXCHANGE_CALL}:${exchangeContext.exchangeId}`}
        topicMessages={topicMessages}
        setTopicMessages={setTopicMessages}
      >
        {children}
      </WSCallProvider>
    </WSTextProvider>
  </ExchangeContext.Provider>,
    [exchangeContext]
  );
}

export default ExchangeProvider;
