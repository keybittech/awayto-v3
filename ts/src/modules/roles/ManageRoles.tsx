import React, { useState, useMemo, Suspense } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';

import CreateIcon from '@mui/icons-material/Create';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import DeleteIcon from '@mui/icons-material/Delete';

import { DataGrid } from '@mui/x-data-grid';

import ManageRoleModal from './ManageRoleModal';

import { siteApi, useGrid, useStyles, dayjs, IGroupRole } from 'awayto/hooks';

export function ManageRoles(): React.JSX.Element {

  const classes = useStyles();

  const { data: groupRolesRequest, refetch: getGroupRoles } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const { data: profileRequest, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const roleSet = useMemo<IGroupRole[]>(() => groupRolesRequest?.groupRoles ?? [], [groupRolesRequest?.groupRoles]);

  const [deleteRole] = siteApi.useRoleServiceDeleteRoleMutation();
  const [deleteGroupRole] = siteApi.useGroupRoleServiceDeleteGroupRoleMutation();

  const [editRole, setEditRole] = useState<IGroupRole>();
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_role'} title="Edit">
        <Button onClick={() => {
          const groupRole = roleSet.find(r => r.id === selected[0]);
          setEditRole(groupRole);
          setDialog('manage_role');
          setSelected([]);
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_role'} title="Delete">
        <Button onClick={() => {
          async function go() {
            const selectedGroupRoleNames = groupRolesRequest?.groupRoles.
              filter(gr => gr.id && selected.includes(gr.id)).
              map(gr => gr.role?.name || '') || []

            if (selectedGroupRoleNames.length) {

              await deleteGroupRole({ ids: selected.join(',') }).unwrap();

              const userRoleIds = Object.values(profileRequest?.userProfile.roles || {}).
                filter(ur => ur.name && selectedGroupRoleNames.includes(ur.name)).
                map(ur => ur.id || '') || []

              if (userRoleIds.length) {
                await deleteRole({ ids: userRoleIds.join(',') }).unwrap();
                void getUserProfileDetails();
              }
              void getGroupRoles();
              setSelected([]);
            }
          }
          void go();
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ]
  }, [selected]);

  const roleGridProps = useGrid({
    rows: roleSet,
    columns: [
      { flex: 1, headerName: 'Name', field: 'name', renderCell: ({ row }) => row.role?.name },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.createdOn)) }
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Typography variant="button">Roles:</Typography>
      <Tooltip key={'manage_role'} title="Create">
        <Button onClick={() => {
          setEditRole(undefined);
          setDialog('manage_role')
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
          <GroupAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>
  })

  return <>
    <Dialog open={dialog === 'manage_role'} fullWidth maxWidth="sm">
      <Suspense>
        <ManageRoleModal editRole={editRole?.role} closeModal={() => {
          setDialog('');
          groupRolesRequest?.groupRoles?.length ? void getGroupRoles() : void getUserProfileDetails();
        }} />
      </Suspense>
    </Dialog>

    <DataGrid {...roleGridProps} />
  </>
}

export default ManageRoles;
