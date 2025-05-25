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

import { siteApi, useGrid, useStyles, dayjs, IGroupRole, targets } from 'awayto/hooks';

export function ManageRoles(_: IComponent): React.JSX.Element {

  const classes = useStyles();

  const { data: groupRolesRequest, refetch: getGroupRoles } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const { data: profileRequest, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const group = useMemo(() => Object.values(profileRequest?.userProfile?.groups || {}).find(g => g.active), [profileRequest?.userProfile]);

  const roleSet = useMemo<IGroupRole[]>(() => groupRolesRequest?.groupRoles ?? [], [groupRolesRequest?.groupRoles]);

  const [deleteGroupRole] = siteApi.useGroupRoleServiceDeleteGroupRoleMutation();

  const [editRole, setEditRole] = useState<IGroupRole>();
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_role'} title="Edit">
        <Button
          {...targets(`manage roles edit`, `edit the currently selected role`)}
          color="info"
          onClick={() => {
            const groupRole = roleSet.find(r => r.id === selected[0]);
            setEditRole(groupRole);
            setDialog('manage_role');
            setSelected([]);
          }}
        >
          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_role'} title="Delete">
        <Button
          {...targets(`manage roles delete`, `delete the currently selected role or roles`)}
          color="error"
          onClick={() => {
            deleteGroupRole({ ids: selected.join(',') }).unwrap().then(async () => {
              setSelected([]);
            }).catch(console.error)
          }}
        >
          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ]
  }, [selected]);

  const roleGridProps = useGrid({
    rows: roleSet,
    columns: [
      { flex: 1, headerName: 'Name', field: 'name', renderCell: ({ row }) => row.name },
      { flex: 1, headerName: 'Default', sortable: false, field: 'isDefault', renderCell: ({ row }) => row.roleId == group?.defaultRoleId ? 'Yes' : 'No' },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(row.createdOn) },
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Typography variant="button">Roles:</Typography>
      <Tooltip key={'manage_role'} title="Create">
        <Button
          {...targets(`manage roles create`, `create a new group role`)}
          color="info"
          onClick={() => {
            setEditRole(undefined);
            setDialog('manage_role')
          }}
        >
          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
          <GroupAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>
  })

  return <>
    <Dialog onClose={setDialog} open={dialog === 'manage_role'} fullWidth maxWidth="sm">
      <Suspense>
        <ManageRoleModal editRole={editRole} isDefault={group?.defaultRoleId == editRole?.roleId} closeModal={() => {
          setDialog('');
          void getGroupRoles(); // refresh roles on screen
          void getUserProfileDetails(); // refresh roles globally
        }} />
      </Suspense>
    </Dialog>

    <DataGrid {...roleGridProps} />
  </>
}

export default ManageRoles;
