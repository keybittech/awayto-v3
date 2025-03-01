import React, { useCallback, useState, useEffect, Suspense, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import DialogContent from '@mui/material/DialogContent';
import TextField from '@mui/material/TextField';
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

import { useUtil, siteApi, IGroup, IGroupSchedule, IGroupService, useStyles, refreshToken, keycloak, targets } from 'awayto/hooks';
import { Breadcrumbs, CircularProgress, Dialog } from '@mui/material';

import ManageGroupModal from '../groups/ManageGroupModal';
import ManageGroupRolesModal from '../roles/ManageGroupRolesModal';
import ManageServiceModal from '../services/ManageServiceModal';
import ManageSchedulesModal from '../group_schedules/ManageSchedulesModal';

const {
  VITE_REACT_APP_APP_HOST_URL
} = import.meta.env;

export function Onboard(_: IComponent): React.JSX.Element {

  window.INT_SITE_LOAD = true;
  console.log('window initialized');

  const navigate = useNavigate();
  // const location = useLocation();
  const classes = useStyles();

  const { setSnack, openConfirm } = useUtil();

  const [group, setGroup] = useState({} as IGroup);
  const [groupService, setGroupService] = useState(JSON.parse(localStorage.getItem('onboarding_service') || '{}') as IGroupService);
  const [groupSchedule, setGroupSchedule] = useState(JSON.parse(localStorage.getItem('onboarding_schedule') || '{}') as IGroupSchedule);
  const [hasCode, setHasCode] = useState(false);
  const [saveToggle, setSaveToggle] = useState(0);
  const [formValidity, setFormValidity] = useState('00000');

  const [assist, setAssist] = useState('');
  const [groupCode, setGroupCode] = useState('');
  const [currentPage, setCurrentPage] = useState(0);
  const [currPage, setCurrPage] = useState<string | false>(location.hash.replace('#state', '').includes('#') ? location.hash.substring(1).split('&')[0] : 'create_group');

  const groupRoleValues = useMemo(() => Object.values(group.roles || {}), [group.roles]);

  const { data: profileReq, refetch: getUserProfileDetails, isUninitialized, isLoading } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const [joinGroup] = siteApi.useGroupUtilServiceJoinGroupMutation();
  const [attachUser] = siteApi.useGroupUtilServiceAttachUserMutation();
  const [activateProfile] = siteApi.useUserProfileServiceActivateProfileMutation();
  const [completeOnboarding] = siteApi.useGroupUtilServiceCompleteOnboardingMutation();

  const reloadProfile = async (): Promise<void> => {
    await refreshToken(61).then(async () => {
      await getUserProfileDetails().unwrap();
      navigate('/');
    }).catch(console.error);
  }

  const joinGroupCb = useCallback(() => {
    if (!groupCode || !/^[a-zA-Z0-9]{8}$/.test(groupCode)) {
      setSnack({ snackType: 'warning', snackOn: 'Invalid group code.' });
      return;
    }
    joinGroup({ joinGroupRequest: { code: groupCode } }).unwrap().then(async () => {
      await attachUser({ attachUserRequest: { code: groupCode } }).unwrap().catch(console.error);
      await activateProfile().unwrap().catch(console.error);
      reloadProfile && await reloadProfile().catch(console.error);
    }).catch(console.error);
  }, [groupCode]);

  const updateValidity = (idx: number, valid: boolean) => {
    setFormValidity(fv => {
      let nfv = [...fv];
      nfv[idx] = valid ? '1' : '0';
      return nfv.join('');
    });
  }

  const changePage = (dir: number) => {
    const nextPage = dir + currentPage;
    if (nextPage >= 0 && nextPage < 5) {
      setCurrentPage(nextPage);
      navigate('#');
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
                  color={formValidity[i] == '1' ? "success" : "primary"}
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
          <Suspense>
            {hasCode ? <Card>
              <CardHeader title="Join with Group Code" />
              <CardContent sx={{ textAlign: 'right' }}>
                <TextField
                  {...targets(`onboard group code input`, `Group Code`, `provide a group code to join a group with`)}
                  fullWidth
                  sx={{ mb: 2 }}
                  value={groupCode}
                  required
                  margin="none"
                  onChange={e => setGroupCode(e.target.value)}
                  slotProps={{
                    input: {
                      endAdornment: <Button
                        {...targets(`group join submit`, `join a group using the provided code`)}
                        color="info"
                        onClick={joinGroupCb}
                      >Join</Button>
                    }
                  }}
                />
              </CardContent>
            </Card> :
              currentPage === 0 ? <>
                <Grid container size="grow" spacing={2}>
                  {!isUninitialized && ((profileReq?.userProfile?.groups && group.name) || !profileReq?.userProfile?.groups) && <ManageGroupModal
                    showCancel={false}
                    editGroup={group}
                    saveToggle={saveToggle}
                    onValidChanged={valid => { updateValidity(0, valid) }}
                    closeModal={() => {
                      changePage(1);
                      setSaveToggle(0);
                      getUserProfileDetails();
                    }}
                  />}
                </Grid>
              </> :
                currentPage == 1 ? <ManageGroupRolesModal
                  showCancel={false}
                  editGroup={group}
                  onValidChanged={valid => { updateValidity(1, valid) }}
                  saveToggle={saveToggle}
                  closeModal={() => {
                    changePage(1);
                    setSaveToggle(0);
                    getUserProfileDetails();
                  }}
                /> :
                  currentPage == 2 ? <ManageServiceModal
                    showCancel={false}
                    groupDisplayName={group.displayName}
                    groupPurpose={group.purpose}
                    editGroupService={groupService}
                    onValidChanged={valid => { updateValidity(2, valid) }}
                    saveToggle={saveToggle}
                    closeModal={(savedService: IGroupService) => {
                      changePage(1);
                      setSaveToggle(0);
                      const newGroupService = { ...groupService, service: savedService };
                      localStorage.setItem('onboarding_service', JSON.stringify(newGroupService));
                      setGroupService(newGroupService);
                    }}
                  /> :
                    currentPage == 3 ? <ManageSchedulesModal
                      showCancel={false}
                      editGroupSchedule={groupSchedule}
                      onValidChanged={valid => { updateValidity(3, valid) }}
                      saveToggle={saveToggle}
                      closeModal={(savedSchedule: IGroupSchedule) => {
                        changePage(1);
                        setSaveToggle(0);
                        localStorage.setItem('onboarding_schedule', JSON.stringify(savedSchedule));
                        setGroupSchedule(savedSchedule);
                      }}
                    /> :
                      currentPage == 4 ? <>
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
                                    reloadProfile && reloadProfile().catch(console.error);
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

          </Suspense>
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
                onClick={() => setHasCode(!hasCode)}>
                {!hasCode ? 'Use Group Code' : 'Cancel'}
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
              disabled={saveToggle > 0 || formValidity[currentPage] != '1' || currentPage + 1 == 5}
              onClick={() => { setSaveToggle((new Date()).getTime()); }}
            >
              {saveToggle == 0 ? 'Next' : <CircularProgress size={16} />}
            </Button>
          </Grid>
        </Grid>
      </Grid>
    </Grid >

    <Dialog slotProps={{ paper: { elevation: 8 } }} open={!!assist} onClose={() => { setAssist(''); }} fullWidth maxWidth="md">
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
