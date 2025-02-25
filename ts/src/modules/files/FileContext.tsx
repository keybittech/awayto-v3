import { createContext } from 'react';
import {
  IFile,
  UseFileContents,
  FileServiceGetFilesApiArg,
  FileServiceGetFilesApiResponse,
  UseSiteQuery
} from 'awayto/hooks';

export interface FileContextType {
  getFiles: UseSiteQuery<FileServiceGetFilesApiArg, FileServiceGetFilesApiResponse>;
  fileContents?: IFile;
  getFileContents: ReturnType<UseFileContents>['getFileContents'];
}

export const FileContext = createContext<FileContextType | null>(null);

export default FileContext;
