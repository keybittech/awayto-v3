import React, { Suspense, useContext, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router';

import Dialog from '@mui/material/Dialog';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';

import CreateIcon from '@mui/icons-material/Create';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import AddIcon from '@mui/icons-material/Add';

import { useUtil, SiteRoles, siteApi, targets, useStyles } from 'awayto/hooks';

import GroupContext, { GroupContextType } from './GroupContext';
import GroupSecure from './GroupSecure';
import ManageSchedules from '../group_schedules/ManageSchedules';
import ManageRoles from '../roles/ManageRoles';
import ManageUsers from '../users/ManageUsers';
import ManageRoleActions from '../roles/ManageRoleActions';
import ManageFeedback from '../feedback/ManageFeedback';
import ManageForms from '../forms/ManageForms';
import ManageServices from '../services/ManageServices';
import ManageGroupModal from './ManageGroupModal';
import GroupSeatModal from './GroupSeatModal';
import { useTheme } from '@mui/material';

const { APP_GROUP_ADMIN, APP_GROUP_ROLES, APP_GROUP_SCHEDULE_KEYS, APP_GROUP_SERVICES, APP_GROUP_USERS } = SiteRoles;

export function ManageGroupHome(props: IComponent): React.JSX.Element {

  const theme = useTheme();
  const classes = useStyles();
  const { component } = useParams();

  const [dialog, setDialog] = useState('');

  const { setSnack } = useUtil();

  const [getUserProfileDetails] = siteApi.useLazyUserProfileServiceGetUserProfileDetailsQuery();

  const {
    group
  } = useContext(GroupContext) as GroupContextType;

  const navigate = useNavigate();

  const menuRoles: Record<string, SiteRoles[]> = {
    services: [APP_GROUP_SERVICES],
    schedules: [APP_GROUP_SCHEDULE_KEYS],
    roles: [APP_GROUP_ROLES],
    users: [APP_GROUP_USERS],
    permissions: [APP_GROUP_ADMIN],
    feedback: [APP_GROUP_ADMIN],
    forms: [APP_GROUP_ADMIN],
  }

  useEffect(() => {
    if (component && !Object.keys(menuRoles).includes(component)) {
      navigate('/');
    }
  }, [component]);

  const menu = Object.keys(menuRoles).map(comp => {
    const selected = comp === component || (!component && comp === 'services');
    return group.name && <GroupSecure key={`menu_${comp}`} contentGroupRoles={menuRoles[comp]}>
      <Grid>
        <Button
          {...targets(`manage group home nav ${comp}`, `navigate to the group ${comp} management page`)}
          variant="underline"
          color={selected ? "primary" : "secondary"}
          sx={{
            ...classes.variableText,
            cursor: 'pointer',
            borderBottom: selected ? `2px solid ${theme.palette.primary.main}` : undefined
          }}
          onClick={() => {
            navigate(`/group/manage/${comp}`);
          }}
        >
          {comp}
        </Button>
      </Grid>
    </GroupSecure>
  });

  const viewPage = useMemo(() => {
    switch (component) {
      case 'schedules':
        return <ManageSchedules {...props} />
      case 'roles':
        return <ManageRoles {...props} />
      case 'users':
        return <ManageUsers {...props} />
      case 'permissions':
        return <ManageRoleActions {...props} />
      case 'feedback':
        return <ManageFeedback {...props} />
      case 'forms':
        return <ManageForms {...props} />
      case 'services':
      default:
        return <ManageServices {...props} />
    }
  }, [component]);

  const copyCode = () => {
    if (!group.code) return;
    window.navigator.clipboard.writeText(group.code).catch(console.error);
    setSnack({ snackType: 'success', snackOn: 'Group code copied to clipboard!' })
  }

  const groupUrl = `https://${window.location.host}/app/join?groupCode=${group.code}`;

  const copyUrl = () => {
    window.navigator.clipboard.writeText(groupUrl).catch(console.error);
    setSnack({ snackType: 'success', snackOn: 'Invite URL copied to clipboard!' })
  }

  return <>
    <Grid container spacing={2}>
      <Grid size={{ xs: 6, md: 3 }}>
        <TextField
          {...targets(`manage group home group name input`, `Group Info`, `details about the group`)}
          fullWidth
          variant="standard"
          value={group.displayName}
          helperText={group.purpose}
          slotProps={{
            inputLabel: {
              shrink: true
            },
            input: {
              readOnly: true,
              endAdornment: <IconButton
                {...targets(`manage group edit`, `edit the group details`)}
                color="info"
                onClick={() => {
                  setDialog('manage_group');
                }}
              >
                <CreateIcon />
              </IconButton>
            }
          }}
        />
      </Grid>
      <Grid size={{ xs: 6, md: 3 }}>
        <TextField
          {...targets(`manage group home code input`, `Group Code`, `the group code used to join groups`)}
          fullWidth
          variant="standard"
          value={group.code}
          slotProps={{
            inputLabel: {
              shrink: true
            },
            input: {
              readOnly: true,
              endAdornment: <IconButton
                {...targets(`manage group home copy code`, `copy the group code to your system clipboard`)}
                color="info"
                onClick={copyCode}
              >
                <ContentCopyIcon />
              </IconButton>
            }
          }}
        />
      </Grid>
      <Grid size={{ xs: 6, md: 3 }}>
        <TextField
          {...targets(`manage group home inviteurl input`, `Invite URL`, `the group invite url`)}
          fullWidth
          variant="standard"
          value={groupUrl}
          slotProps={{
            inputLabel: {
              shrink: true
            },
            input: {
              readOnly: true,
              endAdornment: <IconButton
                {...targets(`manage group home copy join url`, `copy a group join url to your system clipboard`)}
                color="info"
                onClick={copyUrl}
              >
                <ContentCopyIcon />
              </IconButton>
            }
          }}
        />
      </Grid>
      <Grid size={{ xs: 6, md: 3 }}>
        <TextField
          {...targets(`manage group home seats input`, `Seats`, `number of available group seats`)}
          fullWidth
          variant="standard"
          value={group.seatsBalance || 0}
          slotProps={{
            inputLabel: {
              shrink: true
            },
            input: {
              readOnly: true,
              endAdornment: <IconButton
                {...targets(`manage group home add seats`, `open the group seats modal to add more seats to your group`)}
                sx={{ cursor: 'pointer' }}
                onClick={() => setDialog('add_seats')}
                color="info"
              >
                <AddIcon />
              </IconButton>
            }
          }}
        />
      </Grid>
    </Grid>

    <Grid container py={2} spacing={2} justifyContent="flex-start" alignItems="center">
      {menu}
    </Grid>

    {viewPage}

    <Suspense>
      <Dialog onClose={setDialog} open={dialog === 'manage_group'} fullWidth maxWidth="sm">
        <ManageGroupModal {...props} editGroup={group} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Dialog>

      <Dialog onClose={setDialog} open={dialog === 'add_seats'} fullWidth maxWidth="sm">
        <GroupSeatModal {...props} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Dialog>
    </Suspense>
  </>
}

export const roles = [];

export default ManageGroupHome;
