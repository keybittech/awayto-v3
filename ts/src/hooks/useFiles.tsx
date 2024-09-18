import React, { useState, useCallback } from 'react';

import { useComponents } from './useComponents';
import { IFile } from 'awayto/hooks';

export function useFiles(): {
  files: IFile[];
  comp: (props: IComponent) => JSX.Element;
} {
  const { FileManager } = useComponents();
  const [files, setFiles] = useState<IFile[]>([]);

  const comp = useCallback((props: IComponent) => {
    return <FileManager {...props} files={files} setFiles={setFiles} />;
  }, [files]);

  return { files, comp };
}

export default useFiles;
