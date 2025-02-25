import React, { useMemo } from 'react';

import { siteApi, useFileContents } from 'awayto/hooks';

import FileContext, { FileContextType } from './FileContext';

export function FileProvider({ children }: IComponent): React.JSX.Element {

  const getFiles = siteApi.useFileServiceGetFilesQuery();

  const { fileContents, getFileContents } = useFileContents();

  const fileContext: FileContextType = {
    getFiles,
    fileContents,
    getFileContents,
  };

  return useMemo(() => <FileContext.Provider value={fileContext}>
    {children}
  </FileContext.Provider>, [fileContext]);
}
