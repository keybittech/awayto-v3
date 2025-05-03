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
  uploadId: string;
  files: IFile[];
  setFiles: React.Dispatch<React.SetStateAction<IFile[]>>;
}

function FileManager({ uploadId, files, setFiles }: FileManagerProps): React.JSX.Element {

  const [posting, setPosting] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);

  const { openConfirm, setSnack } = useUtil();

  const { postFileContents } = useFileContents();

  const actions = useMemo(() => {
    return [
      <Button
        {...targets(`use files delete`, `delete currently selected file or files`)}
        key={`delete_files_action`}
        color="error"
        onClick={deleteFiles}
      >Delete</Button>,
    ];
  }, [selected]);

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    if (files && event.target.files && event.target.files.length > 0) {

      const newFiles = Array.from(event.target.files);
      const newFilesNames = newFiles.map(nf => nf.name);
      const fileOverwrites = files.filter(ef => ef.name && newFilesNames.includes(ef.name)) as Required<IFile>[];

      const existingFileIds: string[] = [];
      const existingFileNames: string[] = [];
      files.forEach(f => {
        if (!f.uuid || !f.name) return;
        existingFileIds.push(f.uuid);
        existingFileNames.push(f.name);
      });

      if (newFiles.length + (existingFileIds.length - fileOverwrites.length) > 5) {
        setSnack({ snackType: 'warning', snackOn: 'No more than 5 files may be uploaded in total.' });
        return
      }

      openConfirm({
        isConfirming: true,
        confirmEffect: `Upload ${newFiles.length} files` + (fileOverwrites.length ? `, overwriting ${fileOverwrites.length}` : '') + '.',
        confirmAction: () => {
          setPosting(true);
          postFileContents(uploadId, newFiles, existingFileIds, fileOverwrites.map(fo => fo.uuid)).then(newFileIds => {
            if (!newFileIds.length) return;
            setFiles(oldFiles => {
              for (const f of fileOverwrites) {
                const newIdx = newFiles.findIndex(ff => f.name == ff.name)
                const updatedId = newFileIds[newIdx];
                const existingIdx = oldFiles.findIndex(ff => f.name == ff.name);

                oldFiles[existingIdx].uuid = updatedId;
              }

              const uploadedFiles: IFile[] = newFiles.filter(f => !existingFileNames.includes(f.name)).map(f => {
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
          }).finally(() => {
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

      <Grid container size={12}>
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
        {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
      </Grid>
      <Grid container size={12}>
        <Typography variant="caption">
          <p>Submit up to 5 files. 32MB total size limit.</p>
          <p>Allowed Extensions: {allowedFileExt}</p>
        </Typography>
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
