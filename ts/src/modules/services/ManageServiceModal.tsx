import React, { Suspense, useMemo, useEffect, useState, useCallback } from 'react';

import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';

import { useStyles, siteApi, useUtil, useSuggestions, IGroup, IService, IServiceTier, IPrompts, IGroupService } from 'awayto/hooks';
import { Checkbox, FormControlLabel } from '@mui/material';
import FormPicker from '../forms/FormPicker';
import SelectLookup from '../common/SelectLookup';
import ServiceTierAddons from './ServiceTierAddons';

const serviceSchema = {
  name: '',
  cost: 0,
  formId: '',
  surveyId: '',
  tiers: {}
} as IService;

const serviceTierSchema = {
  name: '',
  multiplier: 100,
  formId: '',
  surveyId: '',
  addons: {}
} as IServiceTier;

// const validCost = function(cost: string): boolean {
//   return /(^$|^\$?\d+(,\d{3})*(\.\d*)?$)/.test(cost);
// }

interface ManageServiceModalProps extends IComponent {
  showCancel?: boolean;
  editGroup?: IGroup;
  editGroupService: IGroupService;
  setEditGroupService?: React.Dispatch<React.SetStateAction<IGroupService>>;
  saveToggle?: number;
}

export function ManageServiceModal({ editGroup, editGroupService, setEditGroupService, saveToggle = 0, showCancel = true, closeModal, ...props }: ManageServiceModalProps) {

  const classes = useStyles();

  const { setSnack } = useUtil();

  const [newService, setNewService] = useState({ ...serviceSchema, ...editGroupService?.service });
  const [newServiceTier, setNewServiceTier] = useState({ ...serviceTierSchema });
  const [serviceTierAddonIds, setServiceTierAddonIds] = useState<string[]>([]);
  const [hasServiceFormOrSurvey, setHasServiceFormOrSurvey] = useState(!!newService.formId || !!newService.surveyId);
  const [hasTierFormOrSurvey, setHasTierFormOrSurvey] = useState(!!newServiceTier.formId || !!newServiceTier.surveyId);

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => editGroup || Object.values(profileRequest?.userProfile?.groups || {}).find(g => g.active), [profileRequest?.userProfile, editGroup]);

  const { data: existingServiceRequest } = siteApi.useServiceServiceGetServiceByIdQuery({ id: editGroupService?.id || newService?.id || '' }, { skip: !editGroupService?.id });
  const { data: groupServiceAddonsRequest, refetch: getGroupServiceAddons } = siteApi.useGroupServiceAddonsServiceGetGroupServiceAddonsQuery(undefined, { skip: !group?.id });

  const {
    comp: ServiceSuggestions,
    suggest: suggestServices
  } = useSuggestions('services');

  const {
    comp: TierSuggestions,
    suggest: suggestTiers
  } = useSuggestions('service_tiers');

  const {
    comp: AddonSuggestions,
    suggest: suggestAddons
  } = useSuggestions('service_tier_addons');

  const [postServiceAddon] = siteApi.useServiceAddonServicePostServiceAddonMutation();
  const [postGroupServiceAddon] = siteApi.useGroupServiceAddonsServicePostGroupServiceAddonMutation();
  const [deleteGroupServiceAddon] = siteApi.useGroupServiceAddonsServiceDeleteGroupServiceAddonMutation();
  const [patchService] = siteApi.useServiceServicePatchServiceMutation();
  const [postService] = siteApi.useServiceServicePostServiceMutation();
  const [postGroupService] = siteApi.useGroupServiceServicePostGroupServiceMutation();

  const serviceTiers = useMemo(() => Object.values(newService.tiers || {}), [newService.tiers]);

  const handleSubmit = useCallback(async () => {
    if (!newService.name || !Object.keys(newService?.tiers || {}).length) {
      setSnack({ snackOn: 'Provide the service name and at least 1 tier with at least 1 feature.', snackType: 'info' });
      return;
    }

    if (!setEditGroupService) {
      if (editGroupService?.serviceId) {
        await patchService({ patchServiceRequest: { service: newService } }).unwrap();
      } else {
        const { id: serviceId } = await postService({
          postServiceRequest: {
            service: {
              name: newService.name,
              cost: newService.cost,
              formId: newService.formId,
              surveyId: newService.surveyId,
            }
          }
        }).unwrap();
        const newServiceRef = {
          ...newService,
          id: serviceId
        };
        await patchService({ patchServiceRequest: { service: newServiceRef } }).unwrap();
        await postGroupService({ postGroupServiceRequest: { serviceId } }).unwrap();
        setNewService(newServiceRef);
      }
    }

    closeModal && closeModal(newService);
  }, [newService]);

  const useSuggestTiers = useCallback(() => {
    if (newService.name) {
      void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${newService.name.toLowerCase()} at ${group?.displayName}` });
    }
  }, [newService, group]);

  const useSuggestAddons = useCallback((prompt: string) => {
    void suggestAddons({ id: IPrompts.SUGGEST_FEATURE, prompt });
  }, []);

  // Onboarding handling
  useEffect(() => {
    if (setEditGroupService) {
      setEditGroupService({ service: newService });
    }
  }, [newService.name, newService.tiers]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  useEffect(() => {
    if (newServiceTier.name && serviceTierAddonIds.length) {
      const existingIds = Object.keys(newServiceTier.addons || {});
      const addons = [...existingIds, ...serviceTierAddonIds]
        .filter((v, i, a) => a.indexOf(v) === i)
        .reduce((m, id, i) => {
          if (!serviceTierAddonIds.includes(id)) return m;

          const name = groupServiceAddonsRequest?.groupServiceAddons?.find(gs => gs.serviceAddon?.id === id)?.serviceAddon?.name;
          if (!name) return m;
          return {
            ...m,
            [id]: {
              id,
              name,
              order: i + 1
            }
          }
        }, {});
      setNewServiceTier({ ...serviceTierSchema, ...newServiceTier, addons });
    }
  }, [newServiceTier.name, serviceTierAddonIds, groupServiceAddonsRequest]);

  useEffect(() => {
    if (group?.purpose) {
      void suggestServices({ id: IPrompts.SUGGEST_SERVICE, prompt: group.purpose });
    }
  }, [group?.purpose]);

  useEffect(() => {
    const serv = existingServiceRequest?.service;
    if (serv) {
      setNewService({ ...newService, ...serv });
    }
  }, [existingServiceRequest]);

  useEffect(() => {
    setHasServiceFormOrSurvey(!!(newService.formId || newService.surveyId));
  }, [newService]);

  useEffect(() => {
    setHasTierFormOrSurvey(!!(newServiceTier.formId || newServiceTier.surveyId));
  }, [newServiceTier]);

  return <Card>
    <CardHeader title={`${editGroupService ? 'Edit' : 'Create'} Service`} />
    <CardContent>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 1. Provide details</legend>
            <Typography variant="caption">Services relate to the work performed during appointments. They can be specific or more broad. For example, a "Tutoring" service where all consultants handle all subjects, versus two separate "Math Tutoring" and "English Tutoring" services.</Typography>
            <Box my={2}>
              <TextField
                fullWidth
                label="Service Name"
                required
                value={newService.name}
                onChange={e => setNewService({ ...newService, name: e.target.value })}
                onBlur={() => {
                  if (!newService.name || !group?.displayName) return;
                  useSuggestTiers();
                }}
                helperText={
                  <Suspense>
                    <ServiceSuggestions
                      staticSuggestions='Ex: Website Hosting, Yard Maintenance, Automotive Repair'
                      handleSuggestion={suggestedService => {
                        if (!group?.displayName) return;
                        void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${suggestedService.toLowerCase()} at ${group?.displayName}` });
                        setNewService({ ...newService, name: suggestedService });
                      }}
                    />
                  </Suspense>
                }
              />
            </Box>

            {/* <Box mb={4}>
              <TextField fullWidth label="Cost" helperText="Optional." value={newService.cost || ''} onChange={e => validCost(e.target.value) && setNewService({ ...newService, cost: /\.\d\d/.test(e.target.value) ? parseFloat(e.target.value).toFixed(2) : e.target.value })} />
            </Box> */}

            <FormControlLabel
              label="Include Service Forms"
              control={
                <Checkbox
                  checked={hasServiceFormOrSurvey}
                  onChange={() => setHasServiceFormOrSurvey(!hasServiceFormOrSurvey)}
                />
              }
            />

            {hasServiceFormOrSurvey && <Suspense>
              <Box my={2}>
                <FormPicker
                  formId={newService.formId}
                  label="Intake Form"
                  helperText="Optional. Shown during appointment creation."
                  onSelectForm={(formId: string) => {
                    setNewService({ ...newService, formId });
                  }}
                />
              </Box>
              <Box my={2}>
                <FormPicker
                  formId={newService.surveyId}
                  label="Survey Form"
                  helperText="Optional. Shown during post-appointment summary."
                  onSelectForm={(surveyId: string) => {
                    setNewService({ ...newService, surveyId });
                  }}
                />
              </Box>
            </Suspense>}
          </Box>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 2. Add a Tier</legend>
            <Typography variant="caption">Tiers describe the context and features that go along with a service. For example, a "bronze, silver, gold" ranking system, or subject categories like "English 1010, English 2010, etc.".</Typography>
            <Box my={1}>
              <Typography variant="body1"> At least 1 tier is required.</Typography>
            </Box>
            <Box my={2}>
              <TextField
                fullWidth
                label="Tier Name"
                required
                value={newServiceTier.name}
                onChange={e => setNewServiceTier({ ...newServiceTier, name: e.target.value })}
                onBlur={() => {
                  if (!newServiceTier.name || !newService.name) return;
                  useSuggestAddons(`${newServiceTier.name} ${newService.name}`);
                }}
                helperText={
                  <Suspense>
                    <TierSuggestions
                      hideSuggestions={!newService.name}
                      staticSuggestions='Ex: Basic, Mid-Tier, Advanced'
                      handleSuggestion={suggestedTier => {
                        useSuggestAddons(`${suggestedTier} ${newService.name}`);
                        setNewServiceTier({ ...newServiceTier, name: suggestedTier })
                      }}
                    />
                  </Suspense>
                }
              />
            </Box>

            <Box my={2} flexDirection="column" sx={{ display: 'flex', alignItems: 'baseline' }}>
              <Suspense>
                <SelectLookup
                  multiple
                  lookupName='Feature'
                  lookups={groupServiceAddonsRequest?.groupServiceAddons?.map(gsa => gsa.serviceAddon)}
                  lookupValue={serviceTierAddonIds}
                  helperText={
                    <AddonSuggestions
                      hideSuggestions={!newServiceTier.name}
                      staticSuggestions='Ex: 24-Hour Support, Premium Access, Domain Registration, 20GB Storage'
                      handleSuggestion={suggestedAddon => {
                        const existingId = groupServiceAddonsRequest?.groupServiceAddons?.find(gsa => gsa.serviceAddon?.name === suggestedAddon)?.serviceAddon?.id;
                        if (!existingId || (existingId && !serviceTierAddonIds.includes(existingId))) {
                          if (existingId) {
                            setServiceTierAddonIds([...serviceTierAddonIds, existingId])
                          } else {
                            postServiceAddon({ postServiceAddonRequest: { name: suggestedAddon } }).unwrap().then(({ id: serviceAddonId }) => {
                              postGroupServiceAddon({ serviceAddonId }).unwrap().then(async () => {
                                await getGroupServiceAddons();
                                if (serviceAddonId) {
                                  !serviceTierAddonIds.includes(serviceAddonId) && setServiceTierAddonIds([...serviceTierAddonIds, serviceAddonId]);
                                }
                              }).catch(console.error);
                            }).catch(console.error);
                          }
                        }
                      }}
                    />
                  }
                  lookupChange={(selectedAddonIds: string[]) => {
                    setServiceTierAddonIds([...selectedAddonIds]);
                  }}
                  createAction={async ({ name }) => {
                    return await postServiceAddon({ postServiceAddonRequest: { name } }).unwrap();
                  }}
                  deleteAction={async ({ serviceAddonId }) => {
                    await deleteGroupServiceAddon({ serviceAddonId }).unwrap();
                  }}
                  deleteActionIdentifier='serviceAddonId'
                  deleteComplete={(val: string) => {
                    const tiers = { ...newService.tiers };
                    Object.values(tiers).forEach(tier => {
                      if (tier.addons) {
                        delete tier.addons[val];
                      }
                    })
                    setNewService({ ...newService, tiers });
                  }}
                  refetchAction={getGroupServiceAddons}
                  attachAction={async ({ serviceAddonId }) => {
                    await postGroupServiceAddon({ serviceAddonId }).unwrap();
                  }}
                  attachName='serviceAddonId'
                  {...props}
                />
              </Suspense>
            </Box>

            <Box>
              <FormControlLabel
                label="Include Tier Forms"
                control={
                  <Checkbox
                    checked={hasTierFormOrSurvey}
                    onChange={() => setHasTierFormOrSurvey(!hasTierFormOrSurvey)}
                  />
                }
              />

              {hasTierFormOrSurvey && <Suspense>
                <>
                  <Box my={2}>
                    <FormPicker
                      formId={newServiceTier.formId}
                      label="Intake Form"
                      helperText="Optional. Shown during appointment creation."
                      onSelectForm={(formId: string) => {
                        setNewServiceTier({ ...newServiceTier, formId });
                      }}
                    />
                  </Box>
                  <Box my={2}>
                    <FormPicker
                      formId={newServiceTier.surveyId}
                      label="Survey Form"
                      helperText="Optional. Shown during post-appointment summary."
                      onSelectForm={(surveyId: string) => {
                        setNewServiceTier({ ...newServiceTier, surveyId });
                      }}
                    />
                  </Box>
                </>
              </Suspense>}
            </Box>

            {/* <Box>
              <Typography variant="h6">Multiplier</Typography>
              <Box sx={{ display: 'flex', alignItems: 'baseline' }}>
                <span>{newServiceTier.multiplier}x <span>&nbsp;</span> &nbsp;</span>
                <Slider value={parseFloat(newServiceTier.multiplier)} onChange={(e, val) => setNewServiceTier({ ...newServiceTier, multiplier: parseFloat(val.toString()).toFixed(2) })} step={.01} min={1} max={5} />
              </Box>
            </Box> */}
            <Grid container size="grow" justifyContent="space-between">
              {newServiceTier.id ? <Button
                color="error"
                onClick={() => {
                  const tiers = { ...newService.tiers };
                  if (newServiceTier.id) {
                    delete tiers[newServiceTier.id];
                    setNewService({ ...newService, tiers });
                    setNewServiceTier({ ...serviceTierSchema });
                    setServiceTierAddonIds([]);
                  }
                }}
              >
                Delete
              </Button> : <Box />}
              <Button
                color="secondary"
                onClick={() => {
                  let st = { ...newServiceTier };
                  if (st.name && newService.tiers) {
                    const et = serviceTiers.find(x => x.name == st.name);
                    if (et) {
                      if (et.id != st.id) {
                        setSnack({ snackType: "warning", snackOn: "This tier already exists. Click on " });
                        return
                      }
                      st = { ...st, id: et.id, createdOn: et.createdOn, order: et.order, addons: st.addons };
                    }

                    if (!st.id) {
                      const created = (new Date()).getTime().toString();
                      st.id = created;
                      st.createdOn = created;
                      st.order = serviceTiers.length + 1;
                    }
                    setNewServiceTier({ ...serviceTierSchema });
                    setServiceTierAddonIds([]);
                    setNewService({ ...newService, tiers: { ...newService.tiers, [st.id]: st } });
                    setHasTierFormOrSurvey(false);
                  } else {
                    void setSnack({ snackOn: 'Provide a unique tier name.', snackType: 'info' });
                  }
                }}
              >
                {newServiceTier.id ? 'Save Changes' : 'Add Service Tier'}
              </Button>
            </Grid>
          </Box>
        </Grid>
        <Grid size={12}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 3. Review</legend>
            <Box mb={2}>
              <Typography variant="button">Service Name</Typography>
              <Typography sx={{ mb: 1 }} variant="h2">{newService.name}</Typography>
              {!!newService.formId && <Chip color="info" size="small" label="Intake Form" />} &nbsp;
              {!!newService.surveyId && <Chip color="warning" size="small" label="Survey Form" />}
              {!(newService.surveyId || newService.formId) && <Chip size="small" label="No Forms" />}
            </Box>
            <Suspense>
              <ServiceTierAddons
                service={newService}
                showFormChips
                onClickHeader={(tier: IServiceTier) => {
                  if (tier.id) {
                    setNewServiceTier({ ...tier });
                    useSuggestAddons(`${tier.name} ${newService.name}`);
                    setServiceTierAddonIds(Object.keys(tier.addons || {}));
                  }
                }}
              />
            </Suspense>
          </Box>
        </Grid>
      </Grid>

    </CardContent>
    {!setEditGroupService && <CardActions>
      <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
        {showCancel && <Button onClick={closeModal}>Cancel</Button>}
        <Button disabled={!newService.name || newService.tiers && !Object.keys(newService.tiers).length} onClick={handleSubmit}>Save Service</Button>
      </Grid>
    </CardActions>}
  </Card >;
}

export default ManageServiceModal;
