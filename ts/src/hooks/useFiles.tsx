import React, { useState, useCallback, useRef, useMemo } from 'react';

import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import { DataGrid } from '@mui/x-data-grid';

import { IFile } from './api';
import { useUtil } from './useUtil';
import { useFileContents } from './useFileContents';
import { useGrid } from './useGrid';
import { nid, targets } from './util';

const {
  VITE_REACT_APP_ALLOWED_FILE_EXT,
} = import.meta.env;

const allowedFileExt = "." + VITE_REACT_APP_ALLOWED_FILE_EXT.split(" ").join(", .");

export interface FileManagerProps extends IComponent {
  files: IFile[];
  setFiles: React.Dispatch<React.SetStateAction<IFile[]>>;
}

function FileManager({ files, setFiles }: FileManagerProps): React.JSX.Element {

  const fileSelectRef = useRef<HTMLInputElement>(null);

  const [selected, setSelected] = useState<string[]>([]);

  const { openConfirm } = useUtil();

  const { postFileContents } = useFileContents();

  const actions = useMemo(() => {
    return [
      <Button
        {...targets(`use files delete`, `delete currently selected file or files`)}
        key={'delete_selected_files'}
        onClick={deleteFiles}
      >Delete</Button>,
    ];
  }, [selected]);

  const fileGridProps = useGrid({
    rows: files || [],
    noPagination: true,
    columns: [
      {
        flex: 1,
        headerName: 'Name',
        field: 'name'
      },
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Grid container alignItems="baseline">
        <Button
          {...targets(`use files add`, `add files to the current list`)}
          onClick={addFiles}
        >Add File</Button>
        <Box ml={1}>
          <Typography variant="caption">{allowedFileExt}</Typography>
        </Box>
        {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
      </Grid>
    </>
  });

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    if (files && event.target.files && event.target.files.length > 0) {

      const newFiles = Array.from(event.target.files);
      const existingFiles = files.map(f => f.name);
      const fileOverwrites = newFiles.filter(f => existingFiles.includes(f.name))

      openConfirm({
        isConfirming: true,
        confirmEffect: `Upload ${newFiles.length} files` + (fileOverwrites.length ? `, overwriting ${fileOverwrites.length}` : '') + '.',
        confirmAction: async () => {
          const newFileIds = await postFileContents(newFiles);

          setFiles && setFiles(oldFiles => {
            for (const f of fileOverwrites) {

              const newIdx = newFiles.findIndex(ff => f.name == ff.name)
              const updatedId = newFileIds[newIdx];

              const existingIdx = oldFiles.findIndex(ff => f.name == ff.name);

              oldFiles[existingIdx].uuid = updatedId;
            }

            const uploadedFiles: IFile[] = newFiles.filter(f => !existingFiles.includes(f.name)).map(f => {
              const idx = newFiles.findIndex(ff => f.name == ff.name);
              return {
                id: nid('random') as string,
                uuid: newFileIds[idx],
                name: f.name,
                mimeType: f.type
              }
            });

            return [...oldFiles, ...uploadedFiles];
          })

        }

      });
    }
  }, [files]);

  function addFiles() {
    if (fileSelectRef.current) {
      fileSelectRef.current.value = '';
      fileSelectRef.current.click();
    }
  }

  function deleteFiles() {
    if (files && selected.length) {
      setFiles && setFiles([...files.filter(f => f.id && !selected.includes(f.id))]);
    }
  }

  return <>
    <input
      type="file"
      accept={allowedFileExt}
      multiple
      onChange={e => { handleFileChange(e).catch(console.error) }}
      ref={fileSelectRef}
      style={{ display: 'none' }}
    />

    <DataGrid {...fileGridProps} />
  </>
}


export function useFiles(): {
  files: IFile[];
  comp: (props: IComponent) => React.JSX.Element;
} {
  const [files, setFiles] = useState<IFile[]>([]);

  const comp = useCallback((_: IComponent) => {
    return <FileManager files={files} setFiles={setFiles} />;
  }, [files]);

  return { files, comp };
}

export default useFiles;
