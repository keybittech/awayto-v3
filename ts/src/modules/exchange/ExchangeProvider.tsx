import React, { useMemo, useState } from 'react';
import { useParams } from 'react-router';

import { ExchangeActions, useComponents, SocketMessage, siteApi } from 'awayto/hooks';

import ExchangeContext from './ExchangeContext';

export function ExchangeProvider({ children }: IComponent): React.JSX.Element {

  const { exchangeId } = useParams();
  if (!exchangeId) return <></>;

  const { WSTextProvider, WSCallProvider } = useComponents();

  const [topicMessages, setTopicMessages] = useState<SocketMessage[]>([]);

  const exchangeContext = {
    exchangeId,
    topicMessages,
    setTopicMessages,
    getBookingFiles: siteApi.useBookingServiceGetBookingFilesQuery({ id: exchangeId })
  } as ExchangeContextType;

  return useMemo(() => !WSTextProvider || !WSCallProvider ? <></> :
    <ExchangeContext.Provider value={exchangeContext}>
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
    [WSTextProvider, WSCallProvider, exchangeContext]
  );
}

export default ExchangeProvider;
