import React, { useMemo } from 'react';

import { siteApi, useFileContents } from 'awayto/hooks';

import FileContext, { FileContextType } from './FileContext';
import FileManager from './FileManager';

export function FileProvider({ children }: IComponent): React.JSX.Element {

  const getFiles = siteApi.useFileServiceGetFilesQuery();

  const { fileContents, getFileContents } = useFileContents();

  const fileContext: FileContextType = {
    getFiles,
    fileContents,
    getFileContents,
    fileManager: FileManager
  };

  return useMemo(() => <FileContext.Provider value={fileContext}>
    {children}
  </FileContext.Provider>, [fileContext]);
}
