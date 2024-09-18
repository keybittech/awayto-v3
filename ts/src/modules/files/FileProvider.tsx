import React, { useMemo } from 'react';

import { siteApi, useComponents, useContexts, useFileContents } from 'awayto/hooks';

export function FileProvider({ children }: IComponent): React.JSX.Element {

  const { FileContext } = useContexts();

  const { FileManager } = useComponents();

  const getFiles = siteApi.useFileServiceGetFilesQuery();

  const { fileContents, getFileContents } = useFileContents();

  const fileContext = {
    getFiles,
    fileContents,
    getFileContents,
    fileManager: FileManager
  } as FileContextType;

  return useMemo(() => !FileContext ? <></> :
    <FileContext.Provider value={fileContext}>
      {children}
    </FileContext.Provider>,
    [FileContext, fileContext]
  );
}
