import { createContext } from 'react';
import {
  IFile,
  UseFileContents,
  FileServiceGetFilesApiArg,
  FileServiceGetFilesApiResponse,
  UseSiteQuery
} from 'awayto/hooks';
import { FileManagerProps } from './FileManager';

export interface FileContextType {
  getFiles: UseSiteQuery<FileServiceGetFilesApiArg, FileServiceGetFilesApiResponse>;
  fileContents?: IFile;
  fileManager: (props: FileManagerProps) => React.JSX.Element;
  getFileContents: ReturnType<UseFileContents>['getFileContents'];
}

export const FileContext = createContext<FileContextType | null>(null);

export default FileContext;
