import React, { useMemo, useState, Suspense } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import { useColorScheme } from '@mui/material';

import Box from '@mui/material/Box';
import Switch from '@mui/material/Switch';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Badge from '@mui/material/Badge';
import Menu from '@mui/material/Menu';
import Typography from '@mui/material/Typography';
import MenuList from '@mui/material/MenuList';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';

import CampaignIcon from '@mui/icons-material/Campaign';
import ThreePIcon from '@mui/icons-material/ThreeP';
import LogoutIcon from '@mui/icons-material/Logout';
import EventNoteIcon from '@mui/icons-material/EventNote';
import MoreTimeIcon from '@mui/icons-material/MoreTime';
import GroupIcon from '@mui/icons-material/Group';
import ApprovalIcon from '@mui/icons-material/Approval';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';

import { useSecure, siteApi, keycloak, useStyles, SiteRoles, targets } from 'awayto/hooks';

import UpcomingBookingsMenu from '../bookings/UpcomingBookingsMenu';
import PendingQuotesProvider from '../quotes/PendingQuotesProvider';
import PendingQuotesMenu from '../quotes/PendingQuotesMenu';
import FeedbackMenu from '../feedback/FeedbackMenu';

const {
  VITE_REACT_APP_APP_HOST_URL
} = import.meta.env;

interface TopbarProps extends IComponent {
  forceSiteMenu?: boolean;
}

export function Topbar(props: TopbarProps): React.JSX.Element {

  const { exchangeId } = useParams();
  const classes = useStyles();
  const navigate = useNavigate();
  const hasRole = useSecure();

  const location = useLocation();

  const { mode, setMode } = useColorScheme();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const pendingQuotes = useMemo(() => Object.values(profileRequest?.userProfile?.quotes || {}), [profileRequest?.userProfile]);
  const upcomingBookings = useMemo(() => Object.values(profileRequest?.userProfile?.bookings || {}), [profileRequest?.userProfile]);

  const mobileMenuId = 'mobile-app-bar-menu';
  const pendingQuotesMenuId = 'pending-requests-menu';
  const feedbackMenuId = 'feedback-menu';
  const upcomingBookingsMenuId = 'upcoming-bookings-menu';

  const [mobileMoreAnchorEl, setMobileMoreAnchorEl] = useState<null | HTMLElement>(null);
  const [pendingQuotesAnchorEl, setPendingQuotesAnchorEl] = useState<null | HTMLElement>(null);
  const [feedbackAnchorEl, setFeedbackAnchorEl] = useState<null | HTMLElement>(null);
  const [upcomingBookingsAnchorEl, setUpcomingBookingsAnchorEl] = useState<null | HTMLElement>(null);

  const isMobileMenuOpen = Boolean(mobileMoreAnchorEl);
  const isPendingQuotesOpen = Boolean(pendingQuotesAnchorEl);
  const isFeedbackOpen = Boolean(feedbackAnchorEl);
  const isUpcomingBookingsOpen = Boolean(upcomingBookingsAnchorEl);

  const handleMenuClose = () => {
    setUpcomingBookingsAnchorEl(null);
    setPendingQuotesAnchorEl(null);
    setFeedbackAnchorEl(null);
    setMobileMoreAnchorEl(null);
  };

  const handleNavAndClose = (e: React.SyntheticEvent, path: string) => {
    e.preventDefault();
    handleMenuClose();
    navigate(path);
  };

  return <Grid size={12} sx={{ height: '60px' }} container>
    <Grid sx={{ display: { xs: 'flex', md: true ? 'flex' : !props.forceSiteMenu ? 'none' : 'flex' } }}>
      <Tooltip title="Menu">
        <Button
          {...targets(`topbar open menu`, `show main menu`)}
          role="navigation"
          aria-controls={mobileMenuId}
          aria-haspopup="true"
          onClick={e => setMobileMoreAnchorEl(e.currentTarget)}
        >
          {/* <img src={Icon} alt="kbt-icon" width={28} /> */}
          <Typography alignSelf="center" fontWeight="bold">MENU</Typography>
        </Button>
      </Tooltip>
    </Grid>
    <Grid container sx={{ flexGrow: 1, justifyContent: 'right', alignItems: 'center' }}>
      <Suspense>
        <Grid>
          <UpcomingBookingsMenu
            {...props}
            upcomingBookingsAnchorEl={upcomingBookingsAnchorEl}
            upcomingBookingsMenuId={upcomingBookingsMenuId}
            isUpcomingBookingsOpen={isUpcomingBookingsOpen}
            handleMenuClose={handleMenuClose}
          />

          <Tooltip sx={{ display: !!exchangeId ? 'none' : 'flex' }} title="View Appointments">
            <IconButton
              disableRipple
              color="primary"
              {...targets(`topbar exchange toggle`, `view ${upcomingBookings.length} exchanges`)}
              aria-controls={upcomingBookingsMenuId}
              aria-haspopup="true"
              onClick={e => setUpcomingBookingsAnchorEl(e.currentTarget)}
            >
              <Badge badgeContent={upcomingBookings.length} color="error">
                <ThreePIcon sx={classes.mdHide} />
                <Typography sx={classes.mdShow}>Upcoming</Typography>
              </Badge>
            </IconButton>
          </Tooltip>
        </Grid>


        <Grid>
          {/** PENDING REQUESTS MENU */}
          <Suspense>
            <PendingQuotesProvider>
              <PendingQuotesMenu
                {...props}
                pendingQuotesAnchorEl={pendingQuotesAnchorEl}
                pendingQuotesMenuId={pendingQuotesMenuId}
                isPendingQuotesOpen={isPendingQuotesOpen}
                handleMenuClose={handleMenuClose}
              />
            </PendingQuotesProvider>
          </Suspense>
          {hasRole([SiteRoles.APP_GROUP_SCHEDULES]) && <Tooltip title="Approve Appointments">
            <IconButton
              disableRipple
              color="primary"
              {...targets(`topbar pending toggle`, `show ${pendingQuotes.length} pending exchange requests`)}
              aria-controls={pendingQuotesMenuId}
              aria-haspopup="true"
              onClick={e => setPendingQuotesAnchorEl(e.currentTarget)}
            >
              <Badge badgeContent={pendingQuotes.length} color="error">
                <ApprovalIcon sx={classes.mdHide} />
                <Typography sx={classes.mdShow}>Approve</Typography>
              </Badge>
            </IconButton>
          </Tooltip>}
        </Grid>

        <FeedbackMenu
          feedbackAnchorEl={feedbackAnchorEl}
          feedbackMenuId={feedbackMenuId}
          isFeedbackOpen={isFeedbackOpen}
          handleMenuClose={handleMenuClose}
        />
        <Grid>
          <Tooltip title="Feedback">
            <IconButton
              disableRipple
              color="primary"
              {...targets(`topbar feedback toggle`, `submit group or site feedback`)}
              aria-controls={feedbackMenuId}
              aria-haspopup="true"
              onClick={e => setFeedbackAnchorEl(e.currentTarget)}
            >
              <CampaignIcon sx={classes.mdHide} />
              <Typography sx={classes.mdShow}>Feedback</Typography>
            </IconButton>
          </Tooltip>
        </Grid>
      </Suspense>
    </Grid>

    {/** MOBILE MENU */}
    <Menu
      role="menubar"
      anchorEl={mobileMoreAnchorEl}
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'left',
      }}
      id={mobileMenuId}
      keepMounted
      transformOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      open={isMobileMenuOpen}
      onClose={() => setMobileMoreAnchorEl(null)}
    >
      <Box sx={{ width: 250 }}>
        <MenuList>

          <MenuItem
            {...targets(`main menu go home`, `go to the home page`)}
            onClick={e => handleNavAndClose(e, '/')}
          >
            <ListItemIcon><GroupIcon color={location.pathname === '/' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Home</ListItemText>
          </MenuItem>

          <MenuItem
            {...targets(`main menu profile`, `go to your profile page`)}
            onClick={e => handleNavAndClose(e, '/profile')}
          >
            <ListItemIcon><AccountCircleIcon color={location.pathname === '/profile' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Profile</ListItemText>
          </MenuItem>

          <MenuItem>
            <ListItemText>
              Dark
              <Switch
                {...targets(`main menu toggle color mode`, `toggle to switch between dark and light mode`)}
                checked={mode == 'light'}
                onChange={e => {
                  setMode(e.target.checked ? 'light' : 'dark');
                }}
              />
              Light
            </ListItemText>
          </MenuItem>

          <Divider />

          {/* <MenuItem aria-label="navigate to exchange" onClick={() => handleNavAndClose('/exchange')}>
            <ListItemIcon><TtyIcon color={location.pathname === '/exchange' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Exchange</ListItemText>
          </MenuItem> */}

          {/* <MenuItem hidden={!hasRole([SiteRoles.APP_GROUP_SERVICES])} aria-label="navigate to create service" onClick={() => handleNavAndClose('/service')}>
            <ListItemIcon><BusinessIcon color={location.pathname === '/service' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Service</ListItemText>
          </MenuItem> */}

          <MenuItem
            {...targets(`main menu manage personal schedules`, `view your personal schedules`)}
            hidden={!hasRole([SiteRoles.APP_GROUP_SCHEDULES])}
            onClick={e => handleNavAndClose(e, '/schedule')}
          >
            <ListItemIcon><EventNoteIcon color={location.pathname === '/schedule' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Schedule</ListItemText>
          </MenuItem>

          <MenuItem
            {...targets(`main menu request quote`, `go to make a request for service`)}
            hidden={!hasRole([SiteRoles.APP_GROUP_BOOKINGS])}
            onClick={e => handleNavAndClose(e, '/request')}
          >
            <ListItemIcon><MoreTimeIcon color={location.pathname === '/request' ? "secondary" : "primary"} /></ListItemIcon>
            <ListItemText>Request</ListItemText>
          </MenuItem>

          <Divider />

          <MenuItem
            {...targets(`main menu logout`, `logout of the website`)}
            onClick={() => {
              async function go() {
                localStorage.clear();
                await keycloak.logout({ redirectUri: VITE_REACT_APP_APP_HOST_URL });
              }
              void go();
            }}
          >
            <ListItemIcon><LogoutIcon color="error" /></ListItemIcon>
            <ListItemText>Logout</ListItemText>
          </MenuItem>

        </MenuList>
      </Box>
    </Menu>
  </Grid>;
}

export default Topbar;
