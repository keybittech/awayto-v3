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

import ExitToAppIcon from '@mui/icons-material/ExitToApp';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import CloseIcon from '@mui/icons-material/Close';

import { useUtil, siteApi, IGroup, IGroupSchedule, IGroupService, useStyles, refreshToken, keycloak } from 'awayto/hooks';
import { Breadcrumbs, Dialog } from '@mui/material';

import KbtIcon from '../../img/kbt-icon.png';
import ManageGroupModal from '../groups/ManageGroupModal';
import ManageGroupRolesModal from '../roles/ManageGroupRolesModal';
import ManageServiceModal from '../services/ManageServiceModal';
import ManageSchedulesModal from '../group_schedules/ManageSchedulesModal';

const {
  VITE_REACT_APP_APP_HOST_URL
} = import.meta.env;

export function Onboard(props: IComponent): React.JSX.Element {

  window.INT_SITE_LOAD = true;

  const navigate = useNavigate();
  // const location = useLocation();
  const classes = useStyles();

  const { setSnack, openConfirm } = useUtil();

  const [group, setGroup] = useState({} as IGroup);
  const [groupService, setGroupService] = useState(JSON.parse(localStorage.getItem('onboarding_service') || '{}') as IGroupService);
  const [groupSchedule, setGroupSchedule] = useState(JSON.parse(localStorage.getItem('onboarding_schedule') || '{}') as IGroupSchedule);
  const [hasCode, setHasCode] = useState(false);
  const [saveToggle, setSaveToggle] = useState(0);

  const [assist, setAssist] = useState('');
  const [groupCode, setGroupCode] = useState('');
  const [currentAccordion, setCurrentAccordion] = useState(0);
  // const [currPage, setCurrPage] = useState<string | false>(location.hash.replace('#state', '').includes('#') ? location.hash.substring(1).split('&')[0] : 'create_group');
  // const [expanded, setExpanded] = useState<boolean>(false);

  const groupRoleValues = useMemo(() => Object.values(group.roles || {}), [group.roles]);

  const { data: profileReq, refetch: getUserProfileDetails, isSuccess } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const [joinGroup] = siteApi.useGroupServiceJoinGroupMutation();
  const [attachUser] = siteApi.useGroupServiceAttachUserMutation();
  const [activateProfile] = siteApi.useUserProfileServiceActivateProfileMutation();
  const [completeOnboarding] = siteApi.useGroupServiceCompleteOnboardingMutation();

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

  const accordions = useMemo(() => [
    {
      name: 'Group',
      complete: Boolean(group.name && group.purpose && group.isValid),
      comp: () => <>
        <Typography variant="subtitle1">
          <p>Start by providing a unique name for your group. Group name can be changed later.</p>
          <p>If AI Suggestions are enabled, the group name and description will be used to generate custom suggestions for naming roles, services, and other elements on the site.</p>
          <p>Restrict who can join your group by adding an email to the list of allowed domains. For example, site.com is the domain for the email user@site.com. To ensure only these email accounts can join the group, enter site.com into the Allowed Email Domains and press Add. Multiple domains can be added. Leave empty to allow users with any email address.</p>
          <p>To make onboarding easier, we'll use the example of creating an online learning center. For this step, we give our group a name and description which reflect the group's purpose.</p>
        </Typography>
        <Alert sx={{ py: 0 }} icon={false} severity="info">
          Complete each section in order to progress. The save button will become active when the form is valid.
        </Alert>
      </>
    },
    {
      name: 'Roles',
      complete: Boolean(group.defaultRoleId),
      comp: () => <>
        <Typography variant="subtitle1">
          <p>Roles allow access to different functionality on the site. Each user is assigned 1 role. You have the Admin role.</p>
          <p>If AI is enabled, some role name suggestions have been provided based on your group details. You can add them to your group by clicking on them. Otherwise, click the dropdown to add your own roles.</p>
          <p>Once you've created some roles, set the default role as necessary. This role will automatically be assigned to new users who join your group. Normally you would choose the role which you plan to have the least amount of access.</p>
          <p>For example, our learning center might have Student and Tutor roles. By default, everyone that joins is a Student. If a Tutor joins the group, the Admin can manually change their role in the user list.</p>
        </Typography>
      </>
    },
    {
      name: 'Services',
      complete: Boolean(groupService.service?.name && Object.keys(groupService.service?.tiers || {}).length),
      comp: () => <>
        <Typography variant="subtitle1">
          <p>Services define the context of the appointments that happen within your group. You can add forms and tiers to distinguish the details of your service.</p>
          <p>Forms can be used to enhance the information collected before and after an appointment. Click on a form dropdown to add a new form.</p>
          <p>Each service should have at least 1 tier. The concept of the tiers is up to you, but it essentially allows for the distinction of level of service.</p>
          <p>For example, our learning center creates a service called Writing Tutoring, which has a single tier, General. The General tier has a few features: Feedback, Grammar Help, Brainstorming. Forms are used to get details about the student's paper and then ask how the appointment went afterwards.</p>
        </Typography>
      </>
    },
    {
      name: 'Schedule',
      complete: Boolean(groupSchedule.schedule?.name && groupSchedule.schedule?.startTime),
      comp: () => <>
        <Typography variant="subtitle1">
          <p>Next we create a group schedule. Start by providing basic details about the schedule and when it should be active.</p>
          <p>Some premade options are available to select common defaults. Try selecting a default and adjusting it to your exact specifications.</p>
          <p>For example, at our learning center, students and tutors meet in 30 minute sessions. Tutors work on an hours per week basis. So we create a schedule with a week duration, an hour bracket duration, and a booking slot of 30 minutes.</p>
        </Typography>
      </>
    },
    {
      name: 'Review',
      complete: false,
      comp: () => <>
        <Typography variant="subtitle1">
          <p>Make sure everything looks good, then create your group.</p>
        </Typography>
      </>
    },
  ], [group.name, group.purpose, group.isValid, group.defaultRoleId, groupService.service?.name, groupService.service?.tiers, groupSchedule.schedule?.name, groupSchedule.schedule?.startTime]);

  const changePage = (dir: number) => {
    const nextPage = dir + currentAccordion;
    if (nextPage >= 0 && nextPage < accordions.length) {
      setCurrentAccordion(nextPage)
    }
  }

  const accordionProps = useMemo(() => accordions[currentAccordion], [currentAccordion, accordions]);
  const Help = accordionProps.comp;

  const OnboardingProgress = useCallback(() => <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />}>
    {accordions.map((acc, i) => {
      const curr = i == currentAccordion;
      return <Typography
        key={`acc-progress-${i}`}
        color={acc.complete ? "success" : "primary"}
        sx={{ textDecoration: curr ? 'underline' : 'none' }}
      >{`${i + 1}. ${acc.name}`}</Typography>
    })}
  </Breadcrumbs>, [currentAccordion, accordionProps]);

  useEffect(() => {
    if (isSuccess) {
      const userGroups = Object.values(profileReq?.userProfile?.groups || {});
      if (userGroups.length) {
        const gr = userGroups.find(g => g.ldr) || userGroups.find(g => g.active);
        if (gr) {
          setGroup({ ...gr, isValid: true });
        }
      }
    }
  }, [profileReq, isSuccess]);

  return <>

    <Grid container spacing={2} p={2} sx={{ justifyContent: 'center', bgcolor: 'secondary.main', height: '100vh', alignContent: 'flex-start', overflow: 'auto' }}>

      <Grid container size={{ xs: 12, md: 8, xl: 7 }}>

        <Grid container size="grow" direction="row" sx={{ alignItems: 'center', backgroundColor: '#000', borderRadius: '4px', p: '8px 12px' }}>
          <Grid size="grow" justifyItems="center">
            <Typography>Onboarding Progress</Typography>
            <OnboardingProgress />
          </Grid>
          <Grid>
            <Tooltip title="Logout">
              <IconButton
                onClick={() => {
                  async function go() {
                    localStorage.clear();
                    await keycloak.logout({ redirectUri: VITE_REACT_APP_APP_HOST_URL });
                  }
                  void go();
                }}
              >
                <ExitToAppIcon color="primary" />
              </IconButton>
            </Tooltip>
          </Grid>
        </Grid>
        <Grid container size={12} direction="row" sx={{ alignItems: "center" }}>

          <Button
            sx={classes.onboardingProgress}
            color="warning"
            disableRipple
            disableElevation
            variant="contained"
            disabled={currentAccordion == 0}
            onClick={() => changePage(-1)}
          >
            Back
          </Button>
          <Grid size="grow" >
            <Alert severity="info" color="success" variant="standard" sx={{ alignItems: 'center' }} >
              <Typography>Need help?</Typography>
              <Link sx={{ cursor: 'pointer' }} onClick={() => { setAssist('demo'); }}>Watch the tutorial</Link> or&nbsp;
              <Link sx={{ cursor: 'pointer' }} onClick={() => { setAssist('help'); }}>read about the {accordionProps.name} page</Link>.
            </Alert>
          </Grid>
          <Button
            sx={classes.onboardingProgress}
            color="warning"
            disableRipple
            disableElevation
            variant="contained"
            disabled={!accordionProps.complete || currentAccordion + 1 == accordions.length}
            onClick={() => { setSaveToggle((new Date()).getTime()); }}
          >
            Next
          </Button>
        </Grid>
        <Grid size={12}>
          <Suspense>
            {hasCode ? <Grid size={12} p={2}>
              <TextField
                fullWidth
                sx={{ mb: 2 }}
                value={groupCode}
                required
                onChange={e => setGroupCode(e.target.value)}
                label="Group Code"
              />

              <Grid container justifyContent="space-between">
                <Button onClick={() => setHasCode(false)}>Cancel</Button>
                <Button onClick={joinGroupCb}>Join Group</Button>
              </Grid>
            </Grid> :
              currentAccordion === 0 ? <>
                <Grid container size="grow" spacing={2}>
                  {!group.id && <Button fullWidth disableRipple disableElevation variant="contained" color="primary" onClick={() => setHasCode(true)}>
                    I have a group code
                  </Button>}
                  {isSuccess && <ManageGroupModal
                    {...props}
                    showCancel={false}
                    editGroup={group}
                    setEditGroup={setGroup}
                    saveToggle={saveToggle}
                    closeModal={() => {
                      changePage(1);
                      setSaveToggle(0);
                    }}
                  />}
                </Grid>
              </> :
                currentAccordion == 1 ? <ManageGroupRolesModal
                  {...props}
                  showCancel={false}
                  editGroup={group}
                  setEditGroup={setGroup}
                  saveToggle={saveToggle}
                  closeModal={() => {
                    changePage(1);
                    setSaveToggle(0);
                  }}
                /> :
                  currentAccordion == 2 ? <ManageServiceModal
                    {...props}
                    showCancel={false}
                    editGroup={group}
                    editGroupService={groupService}
                    setEditGroupService={setGroupService}
                    saveToggle={saveToggle}
                    closeModal={(savedService: IGroupService) => {
                      changePage(1);
                      setSaveToggle(0);

                      localStorage.setItem('onboarding_service', JSON.stringify({ ...groupService, service: savedService }));
                    }}
                  /> :
                    currentAccordion == 3 ? <ManageSchedulesModal
                      {...props}
                      showCancel={false}
                      editGroup={group}
                      editGroupSchedule={groupSchedule}
                      setEditGroupSchedule={setGroupSchedule}
                      saveToggle={saveToggle}
                      closeModal={(savedSchedule: IGroupSchedule) => {
                        changePage(1);
                        setSaveToggle(0);
                        localStorage.setItem('onboarding_schedule', JSON.stringify(savedSchedule));
                      }}
                    /> :
                      currentAccordion == 4 ? <>
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
      </Grid>
    </Grid>

    <Dialog slotProps={{ paper: { elevation: 8 } }} open={!!assist} onClose={() => { setAssist(''); }} fullWidth maxWidth="md">
      <Grid size="grow" p={4}>
        <Typography ml={3} variant="body1">{'demo' == assist ? 'Onboarding' : accordionProps.name} Help</Typography>
        <IconButton
          aria-label="close"
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
          {'demo' == assist ? <video controls loop poster={KbtIcon} src="/demos/onboarding.mp4" width="100%" /> : <Help />}
        </DialogContent>
      </Grid>
    </Dialog>
  </>
}

export default Onboard;
