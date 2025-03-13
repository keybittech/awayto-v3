import React, { Suspense, useMemo, useEffect, useState, useCallback, useContext } from 'react';

import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Button from '@mui/material/Button';

import { useStyles, siteApi, useUtil, useSuggestions, IService, IServiceTier, IPrompts, IGroupService, useDebounce, targets, useValid, IValidationAreas } from 'awayto/hooks';
import FormPicker from '../forms/FormPicker';
import SelectLookup from '../common/SelectLookup';
import ServiceTierAddons from './ServiceTierAddons';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';

const serviceSchema = {
  name: '',
  tiers: {}
} as IService;

const serviceTierSchema = {
  name: '',
  addons: {}
} as IServiceTier;

// const validCost = function(cost: string): boolean {
//   return /(^$|^\$?\d+(,\d{3})*(\.\d*)?$)/.test(cost);
// }

interface ManageServiceModalProps extends IComponent {
  showCancel?: boolean;
  groupDisplayName?: string;
  groupPurpose?: string;
  validArea?: keyof IValidationAreas;
  editGroupService?: IGroupService;
  saveToggle?: number;
}

export function ManageServiceModal({ groupDisplayName, groupPurpose, editGroupService, validArea, saveToggle = 0, showCancel = true, closeModal, ...props }: ManageServiceModalProps): React.JSX.Element {

  const classes = useStyles();

  const { setSnack } = useUtil();
  const { setValid } = useValid();

  const [newService, setNewService] = useState({
    ...serviceSchema,
    ...editGroupService?.service,
    ...(JSON.parse(localStorage.getItem(`${validArea}_service`) || '{}') as IGroupService).service
  });
  const debouncedService = useDebounce(newService, 150);
  const [newServiceTier, setNewServiceTier] = useState({ ...serviceTierSchema });
  const [serviceTierAddonIds, setServiceTierAddonIds] = useState<string[]>([]);
  const [hasServiceFormOrSurvey, setHasServiceFormOrSurvey] = useState(!!newService.formId || !!newService.surveyId);
  const [hasTierFormOrSurvey, setHasTierFormOrSurvey] = useState(!!newServiceTier.formId || !!newServiceTier.surveyId);

  const { data: existingServiceRequest } = siteApi.useServiceServiceGetServiceByIdQuery({ id: newService.id || '' }, { skip: !newService.id });
  const { data: groupServiceAddonsRequest, refetch: getGroupServiceAddons } = siteApi.useGroupServiceAddonsServiceGetGroupServiceAddonsQuery();

  const {
    getGroupUserSchedules
  } = useContext(GroupScheduleContext) as GroupScheduleContextType

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

  const serviceTiers = useMemo(() => Object.values(newService.tiers || {}), [newService.tiers]);

  const handleSubmit = useCallback(async () => {
    if (newServiceTier.id?.length) {
      setSnack({ snackType: 'warning', snackOn: 'Please save or delete the ' + newServiceTier.name + ' tier before saving the service.' });
      return;
    }

    if (!newService.name || !Object.keys(newService?.tiers || {}).length) {
      setSnack({ snackType: 'info', snackOn: 'Provide the service name and at least 1 tier with at least 1 feature.' });
      return;
    }

    if (validArea != 'onboarding') {
      if (editGroupService?.serviceId) {
        await patchService({ patchServiceRequest: { service: newService } }).unwrap();
      } else {
        const { id } = await postService({
          postServiceRequest: {
            service: newService
          }
        }).unwrap();
        setNewService({ ...newService, id });
      }

      await getGroupUserSchedules.refetch().unwrap();
    }

    closeModal && closeModal(newService);
  }, [newService, newServiceTier.id]);

  const useSuggestTiers = useCallback(() => {
    if (debouncedService.name && groupDisplayName) {
      void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${debouncedService.name.toLowerCase()} at ${groupDisplayName}` });
    }
  }, [debouncedService, groupDisplayName]);

  const useSuggestAddons = useCallback((prompt: string) => {
    void suggestAddons({ id: IPrompts.SUGGEST_FEATURE, prompt });
  }, []);

  // Onboarding handling
  useEffect(() => {
    if (validArea) {
      localStorage.setItem(`${validArea}_service`, JSON.stringify({ service: debouncedService }));
      setValid({ area: validArea, schema: 'service', valid: Boolean(debouncedService.name && Object.keys(debouncedService.tiers || {}).length) });
    }
  }, [validArea, debouncedService]);

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
    if (groupPurpose) {
      void suggestServices({ id: IPrompts.SUGGEST_SERVICE, prompt: groupPurpose });
    }
  }, [groupPurpose]);

  useEffect(() => {
    const serv = existingServiceRequest?.service;
    if (serv) {
      setNewService({ ...newService, ...serv });
    }
  }, [existingServiceRequest]);

  useEffect(() => {
    setHasServiceFormOrSurvey(Boolean(newService.formId || newService.surveyId));
  }, [newService]);

  useEffect(() => {
    setHasTierFormOrSurvey(Boolean(newServiceTier.formId || newServiceTier.surveyId));
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
                {...targets(`manage service modal service name`, `Service Name`, `edit the name of the service`)}
                fullWidth
                required
                value={newService.name}
                onChange={e => setNewService({ ...newService, name: e.target.value })}
                onBlur={() => {
                  if (!newService.name || !groupDisplayName) return;
                  useSuggestTiers();
                }}
                helperText={
                  <Suspense>
                    <ServiceSuggestions
                      staticSuggestions='Ex: Website Hosting, Yard Maintenance, Automotive Repair'
                      handleSuggestion={suggestedService => {
                        if (!groupDisplayName) return;
                        void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${suggestedService.toLowerCase()} at ${groupDisplayName}` });
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
              {...targets(`manage service modal service forms`, `Include Service Forms`, `toggle if forms should be included with the service`)}
              control={
                <Checkbox
                  checked={hasServiceFormOrSurvey}
                  onChange={() => {
                    if (newService.surveyId || newService.formId && hasServiceFormOrSurvey) {
                      delete newService.surveyId;
                      delete newService.formId;
                      setNewService({ ...newService });
                    }
                    setHasServiceFormOrSurvey(!hasServiceFormOrSurvey)
                  }}
                />
              }
            />

            {hasServiceFormOrSurvey && <Suspense>
              <Box my={2}>
                <FormPicker
                  formId={newService.formId}
                  label="Service Intake Form"
                  helperText="Optional. Shown during appointment creation."
                  onSelectForm={(formId?: string) => {
                    setNewService({ ...newService, formId });
                  }}
                />
              </Box>
              <Box my={2}>
                <FormPicker
                  formId={newService.surveyId}
                  label="Service Survey Form"
                  helperText="Optional. Shown during post-appointment summary."
                  onSelectForm={(surveyId?: string) => {
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
                {...targets(`manage service modal service tier name`, `Tier Name`, `edit the name of the service tier`)}
                fullWidth
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
                {...targets(`manage service modal tier forms`, `Include Tier Forms`, `toggle if forms should be included with the current service tier`)}
                control={
                  <Checkbox
                    checked={hasTierFormOrSurvey}
                    onChange={() => {
                      if (newServiceTier.surveyId || newServiceTier.formId && hasTierFormOrSurvey) {
                        delete newServiceTier.surveyId;
                        delete newServiceTier.formId;
                        setNewServiceTier({ ...newServiceTier });
                      }
                      setHasTierFormOrSurvey(!hasTierFormOrSurvey)
                    }}
                  />
                }
              />

              {hasTierFormOrSurvey && <Suspense>
                <>
                  <Box my={2}>
                    <FormPicker
                      formId={newServiceTier.formId}
                      label="Tier Intake Form"
                      helperText="Optional. Shown during appointment creation."
                      onSelectForm={(formId?: string) => {
                        setNewServiceTier({ ...newServiceTier, formId });
                      }}
                    />
                  </Box>
                  <Box my={2}>
                    <FormPicker
                      formId={newServiceTier.surveyId}
                      label="Tier Survey Form"
                      helperText="Optional. Shown during post-appointment summary."
                      onSelectForm={(surveyId?: string) => {
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
                {...targets(`manage service modal delete service tier`, `delete the selected service tier`)}
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
                {...targets(`manage service modal edit tier`, `add service tier to the service or save changes to the currently selected tier`)}
                color="secondary"
                onClick={() => {
                  let st = { ...newServiceTier };
                  if (st.name && newService.tiers) {
                    const et = serviceTiers.find(x => x.name == st.name);
                    if (et) {
                      if (et.id != st.id) {
                        setSnack({ snackType: "warning", snackOn: "This tier already exists. You can edit it by clicking on its header cell in the table below." });
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
    {validArea != 'onboarding' && <CardActions>
      <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
        {showCancel && <Button
          {...targets(`manage service modal close`, `close the service management modal`)}
          color="error"
          onClick={closeModal}
        >Cancel</Button>}
        <Button
          {...targets(`manage service modal submit`, `submit this service for editing or creation`)}
          color="info"
          disabled={!newService.name || newService.tiers && !Object.keys(newService.tiers).length}
          onClick={handleSubmit}
        >Save Service</Button>
      </Grid>
    </CardActions>}
  </Card >;
}

export default ManageServiceModal;
