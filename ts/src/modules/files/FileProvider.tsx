import React, { useMemo } from 'react';

import { siteApi, useComponents, useFileContents } from 'awayto/hooks';

import FileContext from './FileContext';

export function FileProvider({ children }: IComponent): React.JSX.Element {

  const { FileManager } = useComponents();

  const getFiles = siteApi.useFileServiceGetFilesQuery();

  const { fileContents, getFileContents } = useFileContents();

  const fileContext = {
    getFiles,
    fileContents,
    getFileContents,
    fileManager: FileManager
  } as FileContextType;

  return useMemo(() => <FileContext.Provider value={fileContext}>
    {children}
  </FileContext.Provider>, [fileContext]);
}
