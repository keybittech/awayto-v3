import React, { useState, useMemo, Suspense } from 'react';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';

import DomainAddIcon from '@mui/icons-material/DomainAdd';
import CreateIcon from '@mui/icons-material/Create';
import DeleteIcon from '@mui/icons-material/Delete';

import { DataGrid } from '@mui/x-data-grid';

import { useGrid, siteApi, useUtil, useStyles, useComponents, dayjs, IGroupService } from 'awayto/hooks';

export function ManageServices(props: IProps): React.JSX.Element {
  const classes = useStyles();

  const { openConfirm } = useUtil();
  const { ManageServiceModal } = useComponents();

  const [deleteGroupService] = siteApi.useGroupServiceServiceDeleteGroupServiceMutation();

  const { data: groupServicesRequest, refetch: getGroupServices } = siteApi.useGroupServiceServiceGetGroupServicesQuery();

  const [groupService, setGroupService] = useState<IGroupService>();
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_service'} title="Edit">
        <Button onClick={() => {
          setGroupService(groupServicesRequest?.groupServices.find(gs => gs.id === selected[0])?.service);
          setDialog('manage_service');
          setSelected([]);
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_service'} title="Delete">
        <Button onClick={() => {
          openConfirm({
            isConfirming: true,
            confirmEffect: 'Are you sure you want to delete these services? This cannot be undone.',
            confirmAction: () => {
              deleteGroupService({ ids: selected.join(',') }).unwrap().then(() => {
                void getGroupServices();
                setSelected([]);
              }).catch(console.error);
            }
          });
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ]
  }, [selected, groupServicesRequest]);

  const serviceGridProps = useGrid({
    rowId: 'serviceId',
    rows: groupServicesRequest?.groupServices || [],
    columns: [
      { flex: 1, headerName: 'Name', field: 'name', renderCell: ({ row }) => row.service?.name },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.service?.createdOn)) }
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Typography variant="button">Services:</Typography>
      <Tooltip key={'create_service'} title="Create">
        <Button onClick={() => {
          setGroupService(undefined);
          setDialog('manage_service');
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
          <DomainAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>
  })

  return <>
    <Dialog open={dialog === 'manage_service'} fullWidth maxWidth="lg">
      <Suspense>
        <Grid container>
          <Grid size={12} sx={{ maxHeight: '80vh', overflowY: 'scroll' }}>
            <ManageServiceModal {...props} editService={groupService} closeModal={() => {
              setDialog('')
              void getGroupServices();
            }} />
          </Grid>
        </Grid>
      </Suspense>
    </Dialog>
    <DataGrid {...serviceGridProps} />
  </>
}

export default ManageServices;
