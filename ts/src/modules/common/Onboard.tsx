import React, { useCallback, useState, useEffect, Suspense, useRef, useMemo, useContext } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

import Grid from '@mui/material/Grid';
import Avatar from '@mui/material/Avatar';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import CardHeader from '@mui/material/CardHeader';
import CheckIcon from '@mui/icons-material/Check';
import Alert from '@mui/material/Alert';

import { useComponents, useUtil, useAccordion, siteApi, IGroup, IGroupSchedule, IGroupService, IService, refreshToken, keycloak } from 'awayto/hooks';

declare global {
  interface IProps {
    reloadProfile?(): Promise<void>;
  }
}

export function Onboard({ reloadProfile, ...props }: IProps): React.JSX.Element {
  window.INT_SITE_LOAD = true;

  const location = useLocation();
  const navigate = useNavigate();

  const { setSnack, openConfirm } = useUtil();

  const { ManageGroupModal, ManageGroupRolesModal, ManageServiceModal, ManageSchedulesModal, AccordionWrap } = useComponents();

  const [group, setGroup] = useState({} as IGroup);
  const [groupService, setGroupService] = useState(JSON.parse(localStorage.getItem('onboarding_service') || '{}') as IGroupService);
  const [groupSchedule, setGroupSchedule] = useState(JSON.parse(localStorage.getItem('onboarding_schedule') || '{}') as IGroupSchedule);
  const [hasCode, setHasCode] = useState(false);

  const [groupCode, setGroupCode] = useState('');
  const [expanded, setExpanded] = useState<string | false>(location.hash.replace('#state', '').includes('#') ? location.hash.substring(1).split('&')[0] : 'create_group');

  const groupRoleValues = useMemo(() => Object.values(group.roles || {}), [group.roles]);

  const onboardTop = useRef<HTMLDivElement>(null);

  const { data: profileReq } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const [joinGroup] = siteApi.useGroupServiceJoinGroupMutation();
  const [attachUser] = siteApi.useGroupServiceAttachUserMutation();
  const [activateProfile] = siteApi.useUserProfileServiceActivateProfileMutation();
  const [completeOnboarding] = siteApi.useGroupServiceCompleteOnboardingMutation();

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

  const setStep = (panel: string) => {
    setExpanded(panel);
    navigate('#' + panel);
    onboardTop.current?.scrollIntoView({ behavior: 'auto', block: 'end', inline: 'nearest' });
  }

  const openTip = (panel: string) => () => {
    if (panel && panel !== expanded) {
      setStep(panel);
    }
  };

  const stepAvatar = (name: string, num: number, complete?: boolean) => () => <Avatar
    sx={{
      bgcolor: complete ? 'green' : expanded === name ? 'secondary.main' : 'primary.contrastText',
      color: 'white'
    }}>
    {complete ? <CheckIcon /> : num}
  </Avatar>;

  const CreateGroup = useAccordion('Group', false, expanded === 'create_group', openTip('create_group'), stepAvatar('create_group', 1, !!group.name));
  const CreateRoles = useAccordion('Roles', false, expanded === 'create_roles', openTip(group.name ? 'create_roles' : ''), stepAvatar('create_roles', 2, !!group.defaultRoleId));
  const CreateService = useAccordion('Service', false, expanded === 'create_service', openTip(group.defaultRoleId ? 'create_service' : ''), stepAvatar('create_service', 3, !!groupService?.service?.name));
  const CreateSchedule = useAccordion('Schedule', false, expanded === 'create_schedule', openTip(groupService.service?.name ? 'create_schedule' : ''), stepAvatar('create_schedule', 4, !!groupSchedule.schedule?.name));
  const Review = useAccordion('Review', false, expanded === 'review', openTip(groupSchedule.schedule?.name ? 'review' : ''), stepAvatar('review', 5));

  useEffect(() => {
    const userGroups = Object.values(profileReq?.userProfile?.groups || {});
    if (userGroups.length) {
      const gr = userGroups.find(g => g.ldr) || userGroups.find(g => g.active);
      if (gr) {
        setGroup(gr);
      }
    }
  }, [profileReq]);

  return <>
    <Box sx={{ bgcolor: 'primary.main' }}>
      <Grid container sx={{ placeItems: { xs: 'start', sm: 'center' }, minHeight: '100vh', height: '100%' }}>

        <Grid size={{ xs: 12, sm: 8, md: 6 }} p={4}>
          <Grid container spacing={2}>

            <Grid size={12}>
              <AccordionWrap {...CreateGroup}>
                <Typography variant="subtitle1">
                  <p>Start by providing a unique name for your group; a url-safe version is generated alongside. Group name can be changed later.</p>
                  <p>If AI Suggestions are enabled, the group name and description will be used to generate custom suggestions for naming roles, services, and other elements on the site.</p>
                  <p>Restrict who can join your group by adding an email to the list of allowed domains. For example, site.com is the domain for the email user@site.com. To ensure only these email accounts can join the group, enter site.com into the Allowed Email Domains and press Add. Multiple domains can be added. Leave empty to allow users with any email address.</p>
                  <p>To make onboarding easier, we'll use the example of creating an online learning center. For this step, we give our group a name and description which reflect the group's purpose.</p>
                </Typography>
                <Alert sx={{ py: 0 }} icon={false} severity="info">
                  Complete each section in order to progress. The save button will become active when the form is valid.
                </Alert>
              </AccordionWrap>

              <AccordionWrap {...CreateRoles}>
                <Typography variant="subtitle1">
                  <p>Roles allow access to different functionality on the site. Each user is assigned 1 role. You have the Admin role.</p>
                  <p>Some role name suggestions have been provided based on your group details. You can add them to your group by clicking on them. Otherwise, click the dropdown to add your own roles.</p>
                  <p>Once you've created some roles, set the default role as necessary. This role will automatically be assigned to new users who join your group. Normally you would choose the role which you plan to have the least amount of access.</p>
                  <p>For example, our learning center might have Student and Tutor roles. By default, everyone that joins is a Student. If a Tutor joins the group, the Admin can manually change their role in the user list.</p>
                </Typography>
              </AccordionWrap>

              <AccordionWrap {...CreateService}>
                <Typography variant="subtitle1">
                  <p>Services define the context of the appointments that happen within your group. You can add forms and tiers to distinguish the details of your service.</p>
                  <p>Forms can be used to enhance the information collected before and after an appointment. Click on a form dropdown to add a new form.</p>
                  <p>Each service should have at least 1 tier. The concept of the tiers is up to you, but it essentially allows for the distinction of level of service.</p>
                  <p>For example, our learning center creates a service called Writing Tutoring, which has a single tier, General. The General tier has a few features: Feedback, Grammar Help, Brainstorming. Forms are used to get details about the student's paper and then ask how the appointment went afterwards.</p>
                </Typography>
              </AccordionWrap>

              <AccordionWrap {...CreateSchedule}>
                <Typography variant="subtitle1">
                  <p>Next we create a group schedule. Start by providing basic details about the schedule and when it should be active.</p>
                  <p>Some premade options are available to select common defaults. Try selecting a default and adjusting it to your exact specifications.</p>
                  <p>For example, at our learning center, students and tutors meet in 30 minute sessions. Tutors work on an hours per week basis. So we create a schedule with a week duration, an hour bracket duration, and a booking slot of 30 minutes.</p>
                </Typography>
              </AccordionWrap>

              <AccordionWrap {...Review}>
                <Typography variant="subtitle1">
                  <p>Make sure everything looks good, then create your group.</p>
                </Typography>
              </AccordionWrap>
            </Grid>

          </Grid>
        </Grid>

        <Grid size={{ xs: 12, sm: 4, md: 6 }}>
          <Suspense>

            <Grid container sx={{ p: 4, height: '100vh', bgcolor: 'secondary.main', placeContent: 'center' }}>
              <Grid size={12} sx={{ maxHeight: '80vh', overflowY: 'scroll' }}>
                <Box ref={onboardTop} />
                {hasCode ? <Grid size={12} p={2}>
                  <TextField
                    fullWidth
                    sx={{ mb: 2 }}
                    value={groupCode}
                    onChange={e => setGroupCode(e.target.value)}
                    label="Group Code"
                  />

                  <Grid container justifyContent="space-between">
                    <Button onClick={() => setHasCode(false)}>Cancel</Button>
                    <Button onClick={joinGroupCb}>Join Group</Button>
                  </Grid>
                </Grid> :
                  expanded === 'create_group' ? <>
                    <ManageGroupModal
                      {...props}
                      showCancel={false}
                      editGroup={group}
                      closeModal={(newGroup: IGroup) => {
                        setGroup({ ...group, ...newGroup });
                        setStep('create_roles');
                      }}
                    />
                    <Button sx={{ mt: 2 }} fullWidth onClick={() => setHasCode(true)}>I have a group code</Button>
                  </> :
                    expanded === 'create_roles' ? <ManageGroupRolesModal
                      {...props}
                      showCancel={false}
                      editGroup={group}
                      closeModal={({ roles, defaultRoleId }: IGroup) => {
                        setGroup({ ...group, roles, defaultRoleId });
                        setStep('create_service');
                      }}
                    /> :
                      expanded === 'create_service' ? <ManageServiceModal
                        {...props}
                        showCancel={false}
                        editGroup={group}
                        editService={groupService.service}
                        closeModal={(savedService: IService) => {
                          setGroupService({ service: savedService });
                          setStep('create_schedule');
                          localStorage.setItem('onboarding_service', JSON.stringify({ service: savedService }));
                        }}
                      /> :
                        expanded === 'create_schedule' ? <ManageSchedulesModal
                          {...props}
                          showCancel={false}
                          editGroup={group}
                          editGroupSchedule={groupSchedule}
                          closeModal={(savedSchedule: IGroupSchedule) => {
                            setGroupSchedule(savedSchedule);
                            setStep('review');
                            localStorage.setItem('onboarding_schedule', JSON.stringify(savedSchedule));
                          }}
                        /> :
                          expanded === 'review' ? <>
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

              </Grid>
            </Grid>

          </Suspense>
        </Grid>
      </Grid>
    </Box>
  </>
}

export default Onboard;
