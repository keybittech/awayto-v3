import React, { useState, useMemo, Suspense } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';

import CreateIcon from '@mui/icons-material/Create';

import { DataGrid } from '@mui/x-data-grid';

import { siteApi, useGrid, dayjs, IGroupUser, useStyles, targets } from 'awayto/hooks';

import ManageUserModal from './ManageUserModal';

export function ManageUsers(props: IComponent): React.JSX.Element {

  const classes = useStyles();

  const { data: groupUsersRequest, refetch: getGroupUsers } = siteApi.useGroupUsersServiceGetGroupUsersQuery();

  const [user, setUser] = useState<IGroupUser>({});
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const actions = useMemo(() => {
    const { length } = selected;
    const actions = length == 1 ? [
      <Tooltip key={'manage_user'} title="Edit">
        <Button
          {...targets(`manage users edit`, `edit a user's group information`)}
          color="info"
          key={'manage_user'}
          onClick={() => {
            const sel = groupUsersRequest?.groupUsers?.find(gu => gu.id === selected[0]);
            if (sel) {
              setUser(sel);
              setDialog('manage_user');
              setSelected([]);
            }
          }}
        >

          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ] : [];

    return [
      ...actions,
      // <IconButton key={'lock_user'} onClick={() => {
      //   api(lockUsersAction, { users: selected.map(u => ({ id: u.id })) }, { load: true });
      //   setToggle(!toggle);
      // }}><LockIcon /></IconButton>,
      // <IconButton key={'unlock_user'} onClick={() => {
      //   api(unlockUsersAction, { users: selected.map(u => ({ id: u.id })) }, { load: true });
      //   setToggle(!toggle);
      // }}><LockOpenIcon /></IconButton>,
    ];
  }, [selected, groupUsersRequest?.groupUsers]);

  const userGridProps = useGrid<IGroupUser>({
    rows: groupUsersRequest?.groupUsers || [],
    columns: [
      { flex: 1, headerName: 'First Name', field: 'firstName', renderCell: ({ row }) => row.userProfile?.firstName },
      { flex: 1, headerName: 'Last Name', field: 'lastName', renderCell: ({ row }) => row.userProfile?.lastName },
      { flex: 1, headerName: 'Email', field: 'email', renderCell: ({ row }) => row.userProfile?.email },
      { flex: 1, headerName: 'Role', field: 'roleName' },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.userProfile?.createdOn)) }
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Typography variant="button">Users</Typography>
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>,
  });

  return <>
    <Dialog onClose={setDialog} open={dialog === 'manage_user'} fullWidth maxWidth="xs">
      <Suspense>
        <ManageUserModal {...props} editRoleId={user.roleId} editUser={user.userProfile} closeModal={() => {
          getGroupUsers()
          setDialog('')
        }} />
      </Suspense>
    </Dialog>

    <DataGrid {...userGridProps} />
  </>
}

export default ManageUsers;
