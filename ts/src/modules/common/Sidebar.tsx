import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

import Grid from '@mui/material/Grid';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Button from '@mui/material/Button';

import GroupIcon from '@mui/icons-material/Group';
import AccountBoxIcon from '@mui/icons-material/AccountBox';
import ExitToAppIcon from '@mui/icons-material/ExitToApp';
import MoreTimeIcon from '@mui/icons-material/MoreTime';
import EventNoteIcon from '@mui/icons-material/EventNote';

import Icon from '../../img/kbt-icon.png';

import { useSecure, useStyles, SiteRoles, targets, logout } from 'awayto/hooks';

export function Sidebar(): React.JSX.Element {
  const secure = useSecure();
  const nav = useNavigate();
  const navigate = (loc: string) => {
    nav(loc);
  }
  const classes = useStyles();
  const location = useLocation();

  return <Grid container style={{ height: '100vh' }} alignContent="space-between">
    <Grid size={12} style={{ marginTop: '20px' }}>
      <Grid container justifyContent="center">
        <Button
          {...targets(`sidebar go home`, `go to the home page`)}
          onClick={() => navigate('/')}>
          <img src={Icon} width="64" alt="kbt-icon" />
        </Button>
      </Grid>
      <List component="nav">
        <ListItem sx={classes.menuIcon} onClick={() => navigate('/')} key={'home'}>
          <ListItemIcon><GroupIcon color={location.pathname === '/' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Home</ListItemText>
        </ListItem>
        {/* <ListItem sx={classes.menuIcon} onClick={() => navigate('/exchange')} button key={'exchange'}>
          <ListItemIcon><TtyIcon color={location.pathname === '/exchange' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Exchange</ListItemText>
        </ListItem> */}
        {/* {secure([SiteRoles.APP_GROUP_SERVICES]) && <ListItem sx={classes.menuIcon} onClick={() => navigate('/service')} button key={'service'}>
          <ListItemIcon><BusinessIcon color={location.pathname === '/service' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Service</ListItemText>
        </ListItem>} */}
        {secure([SiteRoles.APP_GROUP_SCHEDULES]) && <ListItem sx={classes.menuIcon} onClick={() => navigate('/schedule')} key={'schedule'}>
          <ListItemIcon><EventNoteIcon color={location.pathname === '/schedule' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Schedule</ListItemText>
        </ListItem>}
        {secure([SiteRoles.APP_GROUP_BOOKINGS]) && <ListItem sx={classes.menuIcon} onClick={() => navigate('/request')} key={'request'}>
          <ListItemIcon><MoreTimeIcon color={location.pathname === '/request' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Request</ListItemText>
        </ListItem>}
      </List>
    </Grid>
    <Grid size={12}>
      <List component="nav">
        <ListItem sx={classes.menuIcon} onClick={() => navigate('/profile')} key={'profile'}>
          <ListItemIcon><AccountBoxIcon color={location.pathname === '/profile' ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText sx={classes.menuText}>Profile</ListItemText>
        </ListItem>
        <ListItem sx={classes.menuIcon} onClick={() => {
          async function go() {
            localStorage.clear();
            await logout();
          }
          void go();
        }} key={'logout'}>
          <ListItemIcon><ExitToAppIcon color="primary" /></ListItemIcon>
          <ListItemText sx={classes.menuIcon}>Logout</ListItemText>
        </ListItem>
      </List>
    </Grid>
  </Grid>
}

export default Sidebar;
