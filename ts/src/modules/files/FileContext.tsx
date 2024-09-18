import { createContext } from 'react';
import {
  IFile,
  UseFileContents,
  FileServiceGetFilesApiArg,
  FileServiceGetFilesApiResponse,
  IDefaultedComponent,
  UseSiteQuery
} from 'awayto/hooks';

declare global {
  type FileContextType = {
    getFiles: UseSiteQuery<FileServiceGetFilesApiArg, FileServiceGetFilesApiResponse>;
    fileContents: IFile;
    fileManager: IDefaultedComponent;
    getFileContents: ReturnType<UseFileContents>['getFileContents'];
  }
}

export const FileContext = createContext<FileContextType | null>(null);

export default FileContext;
