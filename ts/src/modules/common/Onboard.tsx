import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import DialogContent from '@mui/material/DialogContent';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import CardHeader from '@mui/material/CardHeader';
import Tooltip from '@mui/material/Tooltip';
import Alert from '@mui/material/Alert';
import Link from '@mui/material/Link';

import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import CloseIcon from '@mui/icons-material/Close';

import { useUtil, siteApi, IGroup, IGroupSchedule, IGroupService, useStyles, refreshToken, keycloak, targets, useAppSelector } from 'awayto/hooks';
import { Breadcrumbs, CircularProgress, Dialog } from '@mui/material';

import OnboardingVideo from './OnboardingVideo';
import ManageGroupModal from '../groups/ManageGroupModal';
import ManageGroupRolesModal from '../roles/ManageGroupRolesModal';
import ManageServiceModal from '../services/ManageServiceModal';
import ManageSchedulesModal from '../group_schedules/ManageSchedulesModal';
import JoinGroupModal from '../groups/JoinGroupModal';

const {
  VITE_REACT_APP_APP_HOST_URL
} = import.meta.env;

export function Onboard(_: IComponent): React.JSX.Element {

  window.INT_SITE_LOAD = true;
  console.log('window initialized');

  const navigate = useNavigate();
  // const location = useLocation();
  const classes = useStyles();

  const { onboarding } = useAppSelector(state => state.valid);

  const { setSnack, openConfirm } = useUtil();

  const [group, setGroup] = useState({} as IGroup);
  const [groupService, setGroupService] = useState(JSON.parse(localStorage.getItem('onboarding_service') || '{}') as IGroupService);
  const [groupSchedule, setGroupSchedule] = useState(JSON.parse(localStorage.getItem('onboarding_schedule') || '{}') as IGroupSchedule);
  const [saveToggle, setSaveToggle] = useState(0);

  const [assist, setAssist] = useState('');
  const pages = [
    { id: 0, name: 'group', complete: onboarding.group },
    { id: 1, name: 'roles', complete: onboarding.roles },
    { id: 2, name: 'service', complete: onboarding.service },
    { id: 3, name: 'schedule', complete: onboarding.schedule },
    { id: 4, name: 'review', complete: false },
  ];

  const loadedPage = location.hash.replace('#state', '').includes('#') ? location.hash.substring(1).split('&')[0] : 'group';
  const lp = pages.find(p => p.name == loadedPage);

  const [currentPage, setCurrentPage] = useState(lp ? lp.id : 0);

  const groupRoleValues = useMemo(() => Object.values(group.roles || {}), [group.roles]);

  const { data: profileReq, refetch: getUserProfileDetails, isUninitialized, isLoading } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const [completeOnboarding] = siteApi.useGroupUtilServiceCompleteOnboardingMutation();

  const reloadProfile = async (): Promise<void> => {
    await refreshToken(61).then(async () => {
      await getUserProfileDetails().unwrap();
      navigate('/');
    }).catch(console.error);
  }

  const changePage = (dir: number) => {
    const nextPage = dir + currentPage;
    if (nextPage >= 0 && nextPage < 5) {
      setCurrentPage(nextPage);
      navigate('#' + pages[nextPage].name);
    }
  }

  useEffect(() => {
    if (!isLoading) {
      const userGroups = Object.values(profileReq?.userProfile?.groups || {});
      if (userGroups.length) {
        const gr = userGroups.find(g => g.ldr) || userGroups.find(g => g.active);
        if (gr) {
          setGroup({ ...gr, isValid: true });
        }
      }
    }
  }, [profileReq, isLoading]);

  return <>

    <Grid container spacing={2} p={2} sx={{ justifyContent: 'center', bgcolor: 'secondary.main', height: '100vh', alignContent: 'flex-start', overflow: 'auto' }}>

      <Grid container size={{ xs: 12, md: 10, lg: 9, xl: 8 }}>

        <Grid container size="grow" direction="row" sx={{ alignItems: 'center', backgroundColor: '#000', borderRadius: '4px', p: '8px 12px' }}>
          <Grid container size="grow" justifyItems="center">
            <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />}>
              {['Group', 'Roles', 'Services', 'Schedule', 'Review'].map((pg, i) => {
                const curr = i == currentPage;
                return <span id={`acc-progress-${i}`} key={`acc-progress-${i}`}><Typography
                  color={pages[i].complete ? "success" : "primary"}
                  sx={{
                    fontWeight: 'bold',
                    textDecoration: curr ? 'underline' : 'none',
                  }}
                >{pg}</Typography></span>
              })}
            </Breadcrumbs>
          </Grid>
          <Grid>
            <Tooltip title="Logout">
              <Button
                {...targets(`site logout`, `logout of the website`)}
                sx={{ fontSize: '10px' }}
                onClick={() => {
                  async function go() {
                    localStorage.clear();
                    await keycloak.logout({ redirectUri: VITE_REACT_APP_APP_HOST_URL });
                  }
                  void go();
                }}
              >
                Logout
              </Button>
            </Tooltip>
          </Grid>
        </Grid>
        <Grid size={12}>
          {currentPage === 0 ? !isUninitialized && ((profileReq?.userProfile?.groups && group.name) || !profileReq?.userProfile?.groups) && <ManageGroupModal
            showCancel={false}
            editGroup={group}
            saveToggle={saveToggle}
            validArea='onboarding'
            closeModal={(g: IGroup) => {
              changePage(1);
              setSaveToggle(0);
              setGroup({ ...group, ...g });
            }}
          /> : currentPage == 1 ? !isUninitialized && ((profileReq?.userProfile?.groups && group.name) || !profileReq?.userProfile?.groups) && <ManageGroupRolesModal
            showCancel={false}
            editGroup={group}
            validArea='onboarding'
            saveToggle={saveToggle}
            closeModal={(g: IGroup) => {
              changePage(1);
              setSaveToggle(0);
              setGroup({ ...group, ...g });
            }}
          /> : currentPage == 2 ? <ManageServiceModal
            showCancel={false}
            groupDisplayName={group.displayName}
            groupPurpose={group.purpose}
            editGroupService={groupService}
            validArea='onboarding'
            saveToggle={saveToggle}
            closeModal={(savedService: IGroupService) => {
              changePage(1);
              setSaveToggle(0);
              setGroupService({ ...groupService, service: savedService });
            }}
          /> : currentPage == 3 ? <ManageSchedulesModal
            showCancel={false}
            editGroupSchedule={groupSchedule}
            validArea='onboarding'
            saveToggle={saveToggle}
            closeModal={(savedSchedule: IGroupSchedule) => {
              changePage(1);
              setSaveToggle(0);
              setGroupSchedule(savedSchedule);
            }}
          /> : currentPage == 4 ? <>
            <Card>
              <CardHeader title="Review" />
              <CardContent>
                <Typography variant="caption">Group Name</Typography> <Typography mb={2} variant="h5">{group.displayName}</Typography>
                <Typography variant="caption">Roles</Typography> <Typography mb={2} variant="h5">{groupRoleValues.map(r => r.name).join(', ')}</Typography>
                <Typography variant="caption">Default Role</Typography> <Typography mb={2} variant="h5">{groupRoleValues.find(r => r.id === group.defaultRoleId)?.name || ''}</Typography>
                <Typography variant="caption">Service Name</Typography> <Typography mb={2} variant="h5">{groupService.service?.name}</Typography>
                <Typography variant="caption">Schedule Name</Typography> <Typography mb={2} variant="h5">{groupSchedule.schedule?.name}</Typography>
              </CardContent>
              <CardActionArea onClick={() => {
                if (!pages[0].complete || !pages[1].complete || !pages[2].complete || !pages[3].complete) {
                  setSnack({ snackType: 'error', snackOn: 'All pages must be completed' });
                  return;
                }
                openConfirm({
                  isConfirming: true,
                  confirmEffect: `Create the group ${group.displayName}.`,
                  confirmAction: submit => {
                    if (submit) {
                      completeOnboarding({
                        completeOnboardingRequest: {
                          service: groupService.service!,
                          schedule: groupSchedule.schedule!
                        }
                      }).unwrap().then(() => {
                        localStorage.removeItem('onboarding_service');
                        localStorage.removeItem('onboarding_schedule');
                        void reloadProfile();
                      }).catch(console.error);
                    }
                  }
                });
              }}>
                <Box m={2} sx={{ display: 'flex', alignItems: 'center' }}>
                  <Typography color="secondary" variant="button">Create Group</Typography>
                </Box>
              </CardActionArea>
            </Card>
          </> : <></>}

        </Grid>
        <Grid container size={12} direction="row" justifyContent="space-between" sx={{ alignItems: "center" }}>
          <Grid height={{ sx: 'unset', sm: '100%' }}>
            <Button
              {...targets(`onboarding previous page`, `return to the previous page`)}
              sx={classes.onboardingProgress}
              color="warning"
              disableRipple
              disableElevation
              variant="contained"
              disabled={currentPage == 0}
              onClick={() => changePage(-1)}
            >
              Back
            </Button>
          </Grid>
          <Grid size={{ xs: 12, sm: 'grow' }} order={{ xs: 3, sm: 2 }}>
            <Alert action={
              !group.id && <Button
                {...targets(`use group code`, `toggle group code entry to join a group instead of creating one`)}
                sx={{ fontSize: '10px' }}
                onClick={() => { setAssist('join_group') }}
              >
                Use Group Code
              </Button>
            } severity="info" color="success" variant="standard" sx={{ alignItems: 'center' }}>
              <Typography sx={{ display: 'inline' }}>Need help?</Typography>&nbsp;
              <Link sx={{ cursor: 'pointer' }} onClick={() => { setAssist('demo'); }}>Watch the tutorial</Link>
            </Alert>
          </Grid>
          <Grid height={{ sx: 'unset', sm: '100%' }} order={{ xs: 2, sm: 3 }}>
            <Button
              {...targets(`onboarding next page`, `go to the next page`)}
              sx={classes.onboardingProgress}
              color="warning"
              disableRipple
              disableElevation
              variant="contained"
              disabled={saveToggle > 0 || !pages[currentPage].complete || currentPage + 1 == 5}
              onClick={() => { setSaveToggle((new Date()).getTime()); }}
            >
              {saveToggle == 0 ? 'Next' : <CircularProgress size={16} />}
            </Button>
          </Grid>
        </Grid>
      </Grid>
    </Grid >


    <Dialog open={assist == 'join_group'} fullWidth maxWidth="sm">
      <JoinGroupModal closeModal={(joined: boolean) => {
        setAssist('');
        if (joined) {
          void reloadProfile();
        }
      }} />
    </Dialog>

    <Dialog slotProps={{ paper: { elevation: 8 } }} open={assist == 'demo'} onClose={() => { setAssist(''); }} fullWidth maxWidth="md">
      <Grid size="grow" p={4}>
        <Typography ml={3} variant="body1">Onboarding Help</Typography>
        <IconButton
          {...targets(`close assist modal`, `close the assistance modal`)}
          onClick={() => { setAssist(''); }}
          sx={(theme) => ({
            position: 'absolute',
            right: 8,
            top: 8,
            color: theme.palette.grey[500],
          })}
        >
          <CloseIcon />
        </IconButton>
        <DialogContent sx={{ m: 3 }} dividers>
          <OnboardingVideo />
        </DialogContent>
      </Grid>
    </Dialog>
  </>
}

export default Onboard;
