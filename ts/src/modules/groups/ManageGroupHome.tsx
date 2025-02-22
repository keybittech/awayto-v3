import React, { Suspense, useContext, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router';

import Tooltip from '@mui/material/Tooltip';
import Dialog from '@mui/material/Dialog';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import CreateIcon from '@mui/icons-material/Create';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

import { useUtil, SiteRoles, useStyles, siteApi } from 'awayto/hooks';

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

const { APP_GROUP_ADMIN, APP_GROUP_ROLES, APP_GROUP_SCHEDULE_KEYS, APP_GROUP_SERVICES, APP_GROUP_USERS } = SiteRoles;

export function ManageGroupHome(props: IComponent): React.JSX.Element {

  const { component } = useParams();

  const classes = useStyles();

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

  const menu = Object.keys(menuRoles).map(comp => {
    const selected = comp === component || (!component && comp === 'services');
    return group.name && <GroupSecure key={`menu_${comp}`} contentGroupRoles={menuRoles[comp]}>
      <Grid>
        <Button
          variant="text"
          color={selected ? "primary" : "secondary"}
          sx={{
            cursor: 'pointer',
            textDecoration: selected ? 'underline' : undefined
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

  const groupUrl = `https://${window.location.hostname}/app/join?groupCode=${group.code}`;

  const copyUrl = () => {
    window.navigator.clipboard.writeText(groupUrl).catch(console.error);
    setSnack({ snackType: 'success', snackOn: 'Invite URL copied to clipboard!' })
  }

  return <>

    <Card sx={{ mb: 2 }}>
      <CardHeader title={
        <>
          {group.displayName}
          <Tooltip key={'manage_group'} title="Edit">
            <IconButton color="info" onClick={() => {
              setDialog('manage_group');
            }}>
              <CreateIcon sx={classes.variableButtonIcon} />
            </IconButton>
          </Tooltip>
        </>
      } subheader={group.purpose} />
      <CardContent>

        <Grid container mb={2} justifyContent="space-between">
          <Grid flex={1}>
            <Typography fontWeight="bold">
              Group Code &nbsp;
              <Tooltip title="Copy Group Code">
                <IconButton size="small" color="info" onClick={copyCode}>
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
            </Typography>
            <Typography>{group.code}</Typography>
          </Grid>
          <Grid flex={1}>
            <Typography fontWeight="bold">
              Invite Url &nbsp;
              <Tooltip title="Copy Invite URL">
                <IconButton size="small" color="info" onClick={copyUrl}>
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
            </Typography>
            <Typography>{groupUrl}</Typography>
          </Grid>
          <Grid flex={1}>
            <Typography fontWeight="bold">
              Seats &nbsp;
              <Tooltip title="Add Seats">
                <Button sx={{ cursor: 'pointer' }} onClick={() => setDialog('add_seats')} color="info" size="small">Add</Button>
              </Tooltip>
            </Typography>
            <Typography>Test</Typography>
          </Grid>
        </Grid>

      </CardContent>
    </Card>

    <Grid container pb={2} spacing={2} justifyContent="flex-start" alignItems="center">
      <Grid>
        <Typography variant="button">Controls:</Typography>
      </Grid>
      {menu}
    </Grid>

    {viewPage}

    <Suspense>
      <Dialog open={dialog === 'manage_group'} fullWidth maxWidth="sm">
        <ManageGroupModal {...props} editGroup={group} closeModal={() => {
          setDialog('');
          void getUserProfileDetails();
        }} />
      </Dialog>

      <Dialog open={dialog === 'add_seats'} fullWidth maxWidth="sm">
        <GroupSeatModal {...props} closeModal={() => {
          setDialog('');
          console.log("did finish")
        }} />
      </Dialog>
    </Suspense>
  </>
}

export const roles = [];

export default ManageGroupHome;
