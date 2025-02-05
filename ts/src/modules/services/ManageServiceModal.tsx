import React, { Suspense, useMemo, useEffect, useState, useCallback } from 'react';

import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import CardActionArea from '@mui/material/CardActionArea';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Link from '@mui/material/Link';

import { useComponents, useStyles, siteApi, useUtil, useSuggestions, IGroup, IService, IServiceTier, IPrompts, IServiceAddon } from 'awayto/hooks';
import { Checkbox, FormControlLabel } from '@mui/material';
import { SettingsEthernetSharp } from '@mui/icons-material';

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

const validCost = function(cost: string): boolean {
  return /(^$|^\$?\d+(,\d{3})*(\.\d*)?$)/.test(cost);
}

declare global {
  interface IComponent {
    showCancel?: boolean;
    editGroup?: IGroup;
    editService?: IService;
  }
}

export function ManageServiceModal({ editGroup, editService, showCancel = true, closeModal, ...props }: IComponent) {

  const classes = useStyles();
  const { SelectLookup, ServiceTierAddons, FormPicker } = useComponents();

  const { setSnack } = useUtil();

  const [newService, setNewService] = useState({ ...serviceSchema, ...editService });
  const [newServiceTier, setNewServiceTier] = useState({ ...serviceTierSchema });
  const [serviceTierAddonIds, setServiceTierAddonIds] = useState<string[]>([]);
  const [hasServiceFormOrSurvey, setHasServiceFormOrSurvey] = useState(!!newService.formId || !!newService.surveyId);
  const [hasTierFormOrSurvey, setHasTierFormOrSurvey] = useState(!!newServiceTier.formId || !!newServiceTier.surveyId);

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => editGroup || Object.values(profileRequest?.userProfile?.groups || {}).find(g => g.active), [profileRequest?.userProfile, editGroup]);

  const { data: existingServiceRequest, refetch: getExistingService } = siteApi.useServiceServiceGetServiceByIdQuery({ id: editService?.id || newService?.id || '' }, { skip: !editService?.id });
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

  const handleSubmit = useCallback(async () => {
    if (!newService.name || !Object.keys(newService?.tiers || {}).length) {
      setSnack({ snackOn: 'Provide the service name and at least 1 tier with at least 1 feature.', snackType: 'info' });
      return;
    }

    if (editGroup) {
      closeModal && closeModal(newService);
      return;
    }

    if (newService?.id) {
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
      // await getExistingService();
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
    <CardHeader title={`${editService ? 'Edit' : 'Create'} Service`} />
    <CardContent>


      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 1. Provide details</legend>
            <Typography variant="caption">Services relate to the work performed during appointments. They can be specific or more broad. For example, a "Tutoring" service where all consultants handle all subjects, versus two separate "Math Tutoring" and "English Tutoring" services.</Typography>
            <Box my={2}>
              <TextField
                fullWidth
                label="Name"
                value={newService.name}
                onChange={e => setNewService({ ...newService, name: e.target.value })}
                onBlur={() => {
                  if (!newService.name || !group?.displayName) return;
                  useSuggestTiers();
                }}
                helperText={
                  <ServiceSuggestions
                    staticSuggestions='Ex: Website Hosting, Yard Maintenance, Automotive Repair'
                    handleSuggestion={suggestedService => {
                      if (!group?.displayName) return;
                      void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${suggestedService.toLowerCase()} at ${group?.displayName}` });
                      setNewService({ ...newService, name: suggestedService });
                    }}
                  />
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
        <Grid size={{ xs: 12, lg: 6 }}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 2. Add a Tier</legend>
            <Typography variant="caption">Tiers describe the context and features that go along with a service. For example, a "bronze, silver, gold" ranking system, or subject categories like "English 1010, English 2010, etc.". At least 1 tier is required.</Typography>
            <Box my={2}>
              <TextField
                fullWidth
                label="Name"
                value={newServiceTier.name}
                onChange={e => setNewServiceTier({ ...newServiceTier, name: e.target.value })}
                onBlur={() => {
                  if (!newServiceTier.name || !newService.name) return;
                  useSuggestAddons(`${newServiceTier.name} ${newService.name}`);
                }}
                helperText={
                  <TierSuggestions
                    hideSuggestions={!newService.name}
                    staticSuggestions='Ex: Basic, Mid-Tier, Advanced'
                    handleSuggestion={suggestedTier => {
                      useSuggestAddons(`${suggestedTier} ${newService.name}`);
                      setNewServiceTier({ ...newServiceTier, name: suggestedTier })
                    }}
                  />
                }
              />
            </Box>

            <Box my={2} flexDirection="column" sx={{ display: 'flex', alignItems: 'baseline' }}>
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
                createAction={postServiceAddon}
                createActionBodyKey='postServiceAddonRequest'
                deleteAction={deleteGroupServiceAddon}
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
                attachAction={postGroupServiceAddon}
                attachName='serviceAddonId'
                {...props}
              />
            </Box>

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

            {/* <Box>
              <Typography variant="h6">Multiplier</Typography>
              <Box sx={{ display: 'flex', alignItems: 'baseline' }}>
                <span>{newServiceTier.multiplier}x <span>&nbsp;</span> &nbsp;</span>
                <Slider value={parseFloat(newServiceTier.multiplier)} onChange={(e, val) => setNewServiceTier({ ...newServiceTier, multiplier: parseFloat(val.toString()).toFixed(2) })} step={.01} min={1} max={5} />
              </Box>
            </Box> */}
            <CardActionArea onClick={() => {
              if (newServiceTier.name && newService.tiers) {
                const existingTierNames = Object.values(newService.tiers).flatMap(t => t.name);
                if (!newServiceTier.id || !existingTierNames.includes(newServiceTier.name)) {
                  const created = (new Date()).getTime().toString();
                  newServiceTier.id = created;
                  newServiceTier.createdOn = created;
                  newServiceTier.order = Object.keys(newService.tiers).length + 1;
                }
                setNewServiceTier({ ...serviceTierSchema });
                setServiceTierAddonIds([]);
                setNewService({ ...newService, tiers: { ...newService.tiers, [newServiceTier.id]: newServiceTier } });
                setHasTierFormOrSurvey(false);
              } else {
                void setSnack({ snackOn: 'Provide a unique tier name.', snackType: 'info' });
              }
            }}>
              <Box m={2} sx={{ display: 'flex', alignItems: 'center' }}>
                <Typography color="secondary" variant="button">Add Tier to Service</Typography>
              </Box>
            </CardActionArea>
            <Typography variant="caption">Click to edit or remove existing tiers.</Typography>
            <Box sx={{ py: 1, display: 'flex', alignItems: 'flex-end', flexWrap: 'wrap' }}>
              {Object.values(newService.tiers || {}).sort((a, b) => new Date(a.createdOn!).getTime() - new Date(b.createdOn!).getTime()).map((tier, i) => {
                return <Box key={`service-tier-chip${i + 1}new`} m={1}>
                  <Chip
                    sx={classes.chipRoot}
                    label={
                      <Typography sx={classes.chipLabel}>
                        {`#${i + 1} ` + tier.name} {/*  + ' (' + (tier.multiplier || 100) / 100 + 'x)'} */}
                      </Typography>
                    }
                    onDelete={() => {
                      const tiers = { ...newService.tiers };
                      if (tier.id) {
                        delete tiers[tier.id];
                        setNewService({ ...newService, tiers });
                      }
                    }}
                    onClick={() => {
                      setNewServiceTier({ ...tier });
                      useSuggestAddons(`${tier.name} ${newService.name}`);
                      setServiceTierAddonIds(Object.keys(tier.addons || {}));
                    }}
                  />
                </Box>
              })}
            </Box>
          </Box>
        </Grid>
        <Grid size={12}>
          <Box p={2} component="fieldset" sx={classes.legendBox}>
            <legend>Step 3. Review</legend>
            <Box mb={2}>
              <Typography sx={{ mb: 1, mt: -2 }} variant="h2">{newService.name}</Typography>
              {!!newService.formId && <Chip color="info" size="small" label="Intake Form" />} &nbsp;
              {!!newService.surveyId && <Chip color="warning" size="small" label="Survey Form" />}
              {!(newService.surveyId || newService.formId) && <Chip size="small" label="No Forms" />}
            </Box>
            <Suspense>
              <ServiceTierAddons service={newService} showFormChips />
            </Suspense>
          </Box>
        </Grid>
      </Grid>

      <Grid container justifyContent={showCancel ? "space-between" : "flex-end"}>
        {showCancel && <Button onClick={closeModal}>Cancel</Button>}
        <Button disabled={!newService.name || newService.tiers && !Object.keys(newService.tiers).length} onClick={handleSubmit}>Save Service</Button>
      </Grid>

    </CardContent>
  </Card >;
}

export default ManageServiceModal;
