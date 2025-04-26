import React, { useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import { useColorScheme } from '@mui/material';

import Switch from '@mui/material/Switch';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Badge from '@mui/material/Badge';
import Menu from '@mui/material/Menu';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';
import ListSubheader from '@mui/material/ListSubheader';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Avatar from '@mui/material/Avatar';

import ThreePIcon from '@mui/icons-material/ThreeP';
import LogoutIcon from '@mui/icons-material/Logout';
import HomeIcon from '@mui/icons-material/HomeWork';
import ApprovalIcon from '@mui/icons-material/Approval';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import DoneIcon from '@mui/icons-material/Done';

import { useSecure, siteApi, useStyles, SiteRoles, targets, logout, SiteRoleDetails, useUtil } from 'awayto/hooks';

import UpcomingBookingsMenu from '../bookings/UpcomingBookingsMenu';
import PendingQuotesProvider from '../quotes/PendingQuotesProvider';
import PendingQuotesMenu from '../quotes/PendingQuotesMenu';
import FeedbackMenu from '../feedback/FeedbackMenu';

import Icon from '../../img/kbt-icon.png';

const {
  VITE_REACT_APP_PROJECT_TITLE,
} = import.meta.env;

interface TopbarProps extends IComponent {
  forceSiteMenu?: boolean;
}

export function Topbar(props: TopbarProps): React.JSX.Element {

  const { exchangeId } = useParams();
  const { openConfirm } = useUtil();
  const classes = useStyles();
  const navigate = useNavigate();
  const secure = useSecure();

  const location = useLocation();

  const { mode, setMode } = useColorScheme();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const pendingQuotes = useMemo(() => Object.values(profileRequest?.userProfile?.quotes || {}), [profileRequest?.userProfile]);
  const upcomingBookings = useMemo(() => Object.values(profileRequest?.userProfile?.bookings || {}), [profileRequest?.userProfile]);

  const mobileMenuId = 'mobile-app-bar-menu';
  const pendingQuotesMenuId = 'pending-requests-menu';
  const upcomingBookingsMenuId = 'upcoming-bookings-menu';

  const [mobileMoreAnchorEl, setMobileMoreAnchorEl] = useState<null | HTMLElement>(null);
  const [pendingQuotesAnchorEl, setPendingQuotesAnchorEl] = useState<null | HTMLElement>(null);
  const [upcomingBookingsAnchorEl, setUpcomingBookingsAnchorEl] = useState<null | HTMLElement>(null);

  const isMobileMenuOpen = Boolean(mobileMoreAnchorEl);
  const isPendingQuotesOpen = Boolean(pendingQuotesAnchorEl);
  const isUpcomingBookingsOpen = Boolean(upcomingBookingsAnchorEl);

  const handleMenuClose = () => {
    setUpcomingBookingsAnchorEl(null);
    setPendingQuotesAnchorEl(null);
    setMobileMoreAnchorEl(null);
  };

  const handleNavAndClose = (e: React.SyntheticEvent, path: string) => {
    e.preventDefault();
    handleMenuClose();
    navigate(path);
  };

  const roleActions = useMemo(() => {
    const userRoleBits = profileRequest?.userProfile.roleBits;
    if (!userRoleBits) {
      return [<></>];
    }

    const actions = [];
    for (let r in SiteRoles) {
      const roleNum = parseInt(r)
      if (roleNum > 0 && roleNum != SiteRoles.APP_GROUP_BOOKINGS && (userRoleBits & roleNum) > 0) {
        const rd = SiteRoleDetails[roleNum as SiteRoles];

        const ActionIcon = rd.icon;

        actions.push(<MenuItem
          {...targets(`available role actions ${rd.description}`, `perform the ${rd.description} action`)}
          key={`role_menu_option_${roleNum}`}
          onClick={e => handleNavAndClose(e, rd.resource)}
        >
          <ListItemIcon><ActionIcon color={location.pathname == rd.resource ? "secondary" : "primary"} /></ListItemIcon>
          <ListItemText>{rd.name}</ListItemText>
        </MenuItem>);
      }
    }

    if (actions.length) {
      actions.unshift(<Divider key={"action-divider"} />);
    }
    return actions;
  }, [location, profileRequest?.userProfile.roleBits]);

  const isSelected = (opt: string) => {
    return { color: location.pathname === opt ? 'secondary.main' : 'primary.main' }
  };

  return <Grid size={12} sx={{ height: '60px' }} container>
    <Grid sx={{ display: { xs: 'flex', md: true ? 'flex' : !props.forceSiteMenu ? 'none' : 'flex' } }}>
      <Tooltip title="Menu">
        <Button
          {...targets(`topbar open menu`, `show main menu`)}
          aria-controls={isMobileMenuOpen ? mobileMenuId : undefined}
          aria-haspopup="true"
          aria-expanded={isMobileMenuOpen ? 'true' : undefined}
          onClick={e => setMobileMoreAnchorEl(e.currentTarget)}
        >
          <Typography alignSelf="center" fontWeight="bold">MENU</Typography>
        </Button>
      </Tooltip>
      {/** MOBILE MENU */}
      <Menu
        anchorEl={mobileMoreAnchorEl}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        id={mobileMenuId}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
        open={isMobileMenuOpen}
        onClose={() => setMobileMoreAnchorEl(null)}
        MenuListProps={{
          'aria-labelledby': 'topbar-open-menu'
        }}
      >
        <ListSubheader sx={{ mb: -1, bgcolor: 'inherit' }}>
          <Grid container spacing={1} alignItems="center" sx={{ mb: -2 }}>
            <Avatar src={Icon} sx={{ width: 24, height: 24 }} />
            <Typography variant="h6">{VITE_REACT_APP_PROJECT_TITLE}</Typography>
          </Grid>
          <Typography variant="button" color="secondary" fontSize="1.2rem" letterSpacing={2}>
            <strong>Preview</strong>
          </Typography>
        </ListSubheader>
        <MenuItem
          {...targets(`main menu go home`, `go to the home page`)}
          onClick={e => handleNavAndClose(e, '/')}
        >
          <ListItemIcon><HomeIcon sx={isSelected('/')} /></ListItemIcon>
          <ListItemText>Home</ListItemText>
        </MenuItem>

        <MenuItem
          {...targets(`main menu profile`, `go to your profile page`)}
          onClick={e => handleNavAndClose(e, '/profile')}
        >
          <ListItemIcon><AccountCircleIcon sx={isSelected('/profile')} /></ListItemIcon>
          <ListItemText>Profile</ListItemText>
        </MenuItem>

        <MenuItem
          {...targets(`main menu toggle color mode`, `toggle to switch between dark and light mode`)}
          onClick={e => {
            e.preventDefault();
            setMode('light' == mode ? 'dark' : 'light');
          }}
        >
          Dark
          <Switch
            checked={mode == 'light'}
          />
          Light
        </MenuItem>

        {roleActions}

        <Divider />

        <MenuItem
          {...targets(`main menu logout`, `logout of the website`)}
          onClick={logout}
        >
          <ListItemIcon><LogoutIcon color="error" /></ListItemIcon>
          <ListItemText>Logout</ListItemText>
        </MenuItem>

      </Menu>
    </Grid>
    <Grid container sx={{ flexGrow: 1, justifyContent: 'right', alignItems: 'center' }}>

      {exchangeId ? <Tooltip title="Go to Exchange Summary">
        <Button
          {...targets(`exchange summary navigate`, `go to exchange summary`)}
          color="success"
          sx={{ backgroundColor: '#203040' }}
          onClick={() => {
            openConfirm({
              isConfirming: true,
              confirmEffect: `Continue to the Exchange summary.`,
              confirmSideEffect: {
                approvalAction: 'All Done',
                approvalEffect: 'Continue to the Exchange summary.',
                rejectionAction: 'Keep Chatting',
                rejectionEffect: 'Return to the chat.',
              },
              confirmAction: approval => {
                if (approval) {
                  navigate(`/exchange/${exchangeId}/summary`);
                }
              }
            });
          }}
          variant="outlined"
          startIcon={<DoneIcon />}
        >
          Go to Summary
        </Button>
      </Tooltip> : <>
        {secure([SiteRoles.APP_GROUP_SCHEDULES]) && <Grid>
          {/** PENDING REQUESTS MENU */}
          <Tooltip title="Approve Appointments">
            <span>
              <IconButton
                {...targets(`topbar pending toggle`, `show ${pendingQuotes.length} pending exchange requests`)}
                disabled={!pendingQuotes.length}
                disableRipple
                aria-controls={isPendingQuotesOpen ? pendingQuotesMenuId : undefined}
                aria-haspopup="true"
                aria-expanded={isPendingQuotesOpen ? 'true' : undefined}
                onClick={e => setPendingQuotesAnchorEl(e.currentTarget)}
              >
                <Badge badgeContent={pendingQuotes.length} color="error">
                  <ApprovalIcon sx={classes.mdHide} />
                  <Typography sx={classes.mdShow}>Approve</Typography>
                </Badge>
              </IconButton>
            </span>
          </Tooltip>
          <PendingQuotesProvider>
            <PendingQuotesMenu
              pendingQuotesAnchorEl={pendingQuotesAnchorEl}
              pendingQuotesMenuId={pendingQuotesMenuId}
              isPendingQuotesOpen={isPendingQuotesOpen}
              handleMenuClose={handleMenuClose}
            />
          </PendingQuotesProvider>
        </Grid>}

        <Grid>
          <Tooltip sx={{ display: !!exchangeId ? 'none' : 'flex' }} title="View Appointments">
            <span>
              <IconButton
                {...targets(`topbar exchange toggle`, `view ${upcomingBookings.length} exchanges`)}
                disabled={!upcomingBookings.length}
                disableRipple
                aria-controls={isUpcomingBookingsOpen ? upcomingBookingsMenuId : undefined}
                aria-haspopup="true"
                aria-expanded={isUpcomingBookingsOpen ? 'true' : undefined}
                onClick={e => setUpcomingBookingsAnchorEl(e.currentTarget)}
              >
                <Badge badgeContent={upcomingBookings.length} color="error">
                  <ThreePIcon sx={classes.mdHide} />
                  <Typography sx={classes.mdShow}>Upcoming</Typography>
                </Badge>
              </IconButton>
            </span>
          </Tooltip>
          <UpcomingBookingsMenu
            upcomingBookingsAnchorEl={upcomingBookingsAnchorEl}
            upcomingBookingsMenuId={upcomingBookingsMenuId}
            isUpcomingBookingsOpen={isUpcomingBookingsOpen}
            handleMenuClose={handleMenuClose}
          />
        </Grid>
        <Grid>
          <FeedbackMenu />
        </Grid>
      </>}
    </Grid>
  </Grid>;
}

export default Topbar;
