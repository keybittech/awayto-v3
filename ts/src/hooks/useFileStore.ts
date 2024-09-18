import { useState, useEffect } from 'react';
import { FileStoreContext, FileStoreStrategies, FileStoreStrategy, FileSystemFileStoreStrategy } from './file_store';

/**
 * `useFileStore` is used to access various types of pre-determined file stores. All stores allow CRUD operations for user-bound files. Internally default instantiates {@link AWSS3FileStoreStrategy}, but you can also pass a {@link FileStoreStrategies} to `useFileStore` for other supported stores.
 * 
 * ```
 * import { useFileStore } from 'awayto/hooks';
 * 
 * const files = useFileStore();
 * 
 * const file: File = ....
 * const fileName: string = '...';
 * 
 * // Make sure the filestore has connected
 * if (files)
 *  await files.post(file, fileName)
 * 
 * ```
 * 
 * @category Hooks
 */
export const useFileStore = (strategyName: FileStoreStrategies | void): FileStoreContext => {

  if (!strategyName)
    strategyName = FileStoreStrategies.FILE_SYSTEM;

  const [fileStore, setFileStore] = useState<FileStoreContext>(new FileStoreContext(new FileSystemFileStoreStrategy()));

  let strategy: FileStoreStrategy;

  useEffect(() => {
    function setup() {
      if (!strategy) {
        switch (strategyName) {
          // case FileStoreStrategies.XYZ: {
            
          // }
          default:
            strategy = new FileSystemFileStoreStrategy();
            break;
        }

        setFileStore(new FileStoreContext(strategy));
      }
    }
    void setup();
  }, [])  

  return fileStore;
}
