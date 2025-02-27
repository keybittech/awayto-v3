import React, { useState, useMemo, Suspense, useContext } from 'react';
import { useNavigate } from 'react-router';

import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';

import ManageAccountsIcon from '@mui/icons-material/ManageAccounts';
import DomainAddIcon from '@mui/icons-material/DomainAdd';
import CreateIcon from '@mui/icons-material/Create';
import DeleteIcon from '@mui/icons-material/Delete';
import Logout from '@mui/icons-material/Logout';

import { DataGrid } from '@mui/x-data-grid';

import { useSecure, useGrid, useUtil, useStyles, keycloak, siteApi, dayjs, IGroup, SiteRoles } from 'awayto/hooks';

import ManageGroupModal from './ManageGroupModal';
import JoinGroupModal from './JoinGroupModal';

export function ManageGroups(props: IComponent): React.JSX.Element {
  const classes = useStyles();

  const { openConfirm } = useUtil();

  const hasRole = useSecure();
  const navigate = useNavigate();

  const [group, setGroup] = useState<IGroup>();
  const [dialog, setDialog] = useState('');
  const [selected, setSelected] = useState<string[]>([]);

  const { data: profileRequest, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const groups = useMemo(() => profileRequest?.userProfile?.groups || {}, [profileRequest?.userProfile]);

  const [deleteGroup] = siteApi.useGroupServiceDeleteGroupMutation();
  const [leaveGroup] = siteApi.useGroupUtilServiceLeaveGroupMutation();

  const actions = useMemo(() => {
    if (!groups) return [];
    const { length } = selected;
    const gr = groups[selected[0]];
    const grldr = gr?.ldr;
    const acts = length == 1 ? [
      !grldr && <Tooltip key={'leave_group'} title="Leave">
        <Button onClick={() => {
          openConfirm({
            isConfirming: true,
            confirmEffect: 'Leave the group ' + gr.name + ' and refresh the session.',
            confirmAction: () => {
              if (gr.code) {
                leaveGroup({ leaveGroupRequest: { code: gr.code } }).unwrap().then(() =>
                  keycloak.clearToken()
                ).catch(console.error);
              }
            }
          });
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Leave</Typography>
          <Logout sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>,
      // grldr && hasRole([SiteRoles.APP_GROUP_ADMIN]) && <Tooltip key={'view_group_details'} title="Details">
      //   <Button onClick={() => {
      //     navigate(`/group/${gr.name}/manage/users`)
      //   }}>
      //     <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Details</Typography>
      //     <ManageAccountsIcon sx={classes.variableButtonIcon} />
      //   </Button>
      // </Tooltip>,
      // grldr && <Tooltip key={'manage_group'} title="Edit">
      //   <Button onClick={() => {
      //     setGroup(groups[selected[0]]);
      //     setDialog('manage_group');
      //     setSelected([]);
      //   }}>
      //     <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
      //     <CreateIcon sx={classes.variableButtonIcon} />
      //   </Button>
      // </Tooltip>
    ] : [];

    return [
      ...acts,
      grldr && <Tooltip key={'delete_group'} title="Delete">
        <Button onClick={() => {
          openConfirm({
            isConfirming: true,
            confirmEffect: 'Delete the group ' + gr.name + ' and refresh the session.',
            confirmAction: () => {
              deleteGroup({ ids: selected.join(',') }).unwrap().then(() => keycloak.clearToken()).catch(console.error);
            }
          });
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ];
  }, [selected, groups]);

  const groupsGridProps = useGrid({
    rows: Object.values(groups || {}) as Required<IGroup>[],
    columns: [
      { flex: 1, headerName: 'Name', field: 'name' },
      { flex: 1, headerName: 'Code', field: 'code' },
      { flex: 1, headerName: 'Users', field: 'usersCount', renderCell: ({ row }) => row.usersCount || 0 },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.createdOn)) },
      ...(hasRole([SiteRoles.APP_GROUP_ADMIN]) ? [
        {
          flex: 1,
          headerName: '',
          field: 'id',
          renderCell: ({ row }: { row: IGroup }) => <Tooltip key={'view_group_details'} title="Details">
            <Button color="secondary" onClick={() => {
              navigate(`/group/manage/users`);
            }}>
              <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Details</Typography>
              <ManageAccountsIcon sx={classes.variableButtonIcon} />
            </Button>
          </Tooltip>
        }
      ] : [])
    ],
    selected,
    onSelected: p => setSelected(p as string[]),
    toolbar: () => <>
      <Typography variant="button">Groups:</Typography>
      <Tooltip key={'join_group'} title="Join">
        <Button key={'join_group_button'} onClick={() => {
          setGroup(undefined);
          setDialog('join_group');
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Join</Typography>
          <DomainAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
      {/* <Tooltip key={'create_group'} title="Create">
        <Button key={'create_group_button'} onClick={() => {
          setGroup(undefined);
          setDialog('create_group');
        }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
          <GroupAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip> */}
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>
  });

  return <>
    <Dialog open={dialog === 'create_group'} fullWidth maxWidth="sm">
      <Suspense>
        <ManageGroupModal {...props} editGroup={group} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Suspense>
    </Dialog>

    <Dialog open={dialog === 'join_group'} fullWidth maxWidth="sm">
      <Suspense>
        <JoinGroupModal {...props} editGroup={group} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Suspense>
    </Dialog>

    <Dialog open={dialog === 'manage_group'} fullWidth maxWidth="sm">
      <Suspense>
        <ManageGroupModal {...props} editGroup={group} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Suspense>
    </Dialog>

    <DataGrid {...groupsGridProps} />

  </>
}

export default ManageGroups;
