import React, { useState, useCallback, useMemo } from 'react';

import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import { DataGrid } from '@mui/x-data-grid';
import { styled } from '@mui/material/styles';

import { IFile, nid, targets, useFileContents, useGrid, useUtil } from 'awayto/hooks';

const {
  VITE_REACT_APP_ALLOWED_FILE_EXT,
} = import.meta.env;

const allowedFileExt = "." + VITE_REACT_APP_ALLOWED_FILE_EXT.split(" ").join(", .");

const VisuallyHiddenInput = styled('input')({
  clip: 'rect(0 0 0 0)',
  clipPath: 'inset(50%)',
  height: 1,
  overflow: 'hidden',
  position: 'absolute',
  bottom: 0,
  left: 0,
  whiteSpace: 'nowrap',
  width: 1,
});

export interface FileManagerProps extends IComponent {
  files: IFile[];
  setFiles: React.Dispatch<React.SetStateAction<IFile[]>>;
}

function FileManager({ files, setFiles }: FileManagerProps): React.JSX.Element {

  const [posting, setPosting] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);

  const { openConfirm } = useUtil();

  const { postFileContents } = useFileContents();

  const actions = useMemo(() => {
    return [
      <Button
        {...targets(`use files delete`, `delete currently selected file or files`)}
        color="error"
        onClick={deleteFiles}
      >Delete</Button>,
    ];
  }, [selected]);

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    if (files && event.target.files && event.target.files.length > 0) {

      const newFiles = Array.from(event.target.files);
      const existingFiles = files.map(f => f.name);
      const fileOverwrites = newFiles.filter(f => existingFiles.includes(f.name))

      openConfirm({
        isConfirming: true,
        confirmEffect: `Upload ${newFiles.length} files` + (fileOverwrites.length ? `, overwriting ${fileOverwrites.length}` : '') + '.',
        confirmAction: () => {
          setPosting(true);
          postFileContents(newFiles).then(newFileIds => {
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
            });
            setPosting(false);
          });
        }
      });
    }
  }, [files]);

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
      <Grid container size="grow" alignItems="baseline">
        <Button
          {...targets(`use files add`, `add files to the current list`)}
          color="info"
          component="label"
          role={undefined}
          tabIndex={-1}
          loading={posting}
        >
          Add File
          <VisuallyHiddenInput
            type="file"
            multiple
            onChange={handleFileChange}
          />
        </Button>
        <Box ml={1}>
          <Typography variant="caption">{allowedFileExt}</Typography>
        </Box>
        {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
      </Grid>
    </>
  });

  function deleteFiles() {
    if (files && selected.length) {
      setFiles && setFiles([...files.filter(f => f.id && !selected.includes(f.id))]);
    }
  }

  return <>
    <DataGrid {...fileGridProps} />
  </>
}

export default FileManager;
