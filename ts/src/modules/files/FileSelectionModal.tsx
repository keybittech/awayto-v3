import React from 'react';

import Grid from '@mui/material/Grid';
import List from '@mui/material/List';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import Alert from '@mui/material/Alert';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';

import { OrderedFiles, targets } from 'awayto/hooks';
import FileTypeIcon from './FileTypeIcon';

interface FileSelectionModalProps extends IComponent {
  fileGroups: OrderedFiles[];
}

export function FileSelectionModal({ closeModal, fileGroups }: FileSelectionModalProps): React.JSX.Element {
  return <Card sx={{ display: 'flex', flex: 1, flexDirection: 'column' }}>
    <CardHeader title="Select File" />
    <CardContent sx={{ display: 'flex', flex: 1, flexDirection: 'column', overflow: 'auto' }}>
      <Grid container>
        {!fileGroups.length ? <Alert variant="outlined" severity="info">
          No files could be found.
        </Alert> : <Grid container>
          {fileGroups.map((group, i) => {
            return <Grid size={{ xs: 12, md: 4 }} key={`file_group_${i}`}>
              <List
                subheader={
                  <ListSubheader>{group.name} files</ListSubheader>
                }
              >
                {group.files.map((f, j) => {
                  return !f.mimeType ? <></> : <ListItemButton
                    {...targets(`file modal select`, `select this file from the list`)}
                    key={`file_${i}_${j}`}
                    onClick={() => {
                      closeModal && closeModal(f);
                    }}>
                    <ListItemIcon>
                      <FileTypeIcon fileType={f.mimeType} />
                    </ListItemIcon>
                    <ListItemText>{f.name}</ListItemText>
                  </ListItemButton>;
                })}
              </List>
            </Grid>;

          })}
        </Grid>
        }
      </Grid>
    </CardContent>
    <CardActions>
      <Grid container justifyContent="space-between">
        <Button
          {...targets(`file modal cancel`, `cancel file selection`)}
          onClick={() => closeModal && closeModal()}
        >Cancel</Button>
      </Grid>
    </CardActions>
  </Card>
}

export default FileSelectionModal;
