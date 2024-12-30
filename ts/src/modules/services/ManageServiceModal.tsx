import React, { Suspense, useMemo, useEffect, useState, useCallback } from 'react';

import TextField from '@mui/material/TextField';
import Dialog from '@mui/material/Dialog';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import CardActionArea from '@mui/material/CardActionArea';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';

import { useComponents, useStyles, siteApi, useUtil, useSuggestions, IGroup, IService, IServiceTier, IPrompts } from 'awayto/hooks';
import Link from '@mui/material/Link';

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
  const { SelectLookup, ServiceTierAddons, ManageFormModal } = useComponents();

  const { setSnack } = useUtil();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const group = useMemo(() => editGroup || Object.values(profileRequest?.userProfile?.groups || {}).find(g => g.active), [profileRequest?.userProfile, editGroup]);

  const { data: existingServiceRequest } = siteApi.useServiceServiceGetServiceByIdQuery({ id: editService?.id || '' }, { skip: !editService?.id });
  const { data: groupServiceAddonsRequest, refetch: getGroupServiceAddons } = siteApi.useGroupServiceAddonsServiceGetGroupServiceAddonsQuery(undefined, { skip: !group?.id });
  const { data: groupFormsRequest, refetch: getGroupForms, isSuccess: groupFormsLoaded } = siteApi.useGroupFormServiceGetGroupFormsQuery(undefined, { skip: !group?.id });

  const [newService, setNewService] = useState({ ...serviceSchema, ...editService });
  const [newServiceTier, setNewServiceTier] = useState({ ...serviceTierSchema });
  const [serviceTierAddonIds, setServiceTierAddonIds] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');
  const [hasServiceForm, setHasServiceForm] = useState(false);
  const [hasTierForm, setHasTierForm] = useState(false);

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
    suggest: suggestAddon
  } = useSuggestions('service_tier_addons');

  const [postServiceAddon] = siteApi.useServiceAddonServicePostServiceAddonMutation();
  const [postGroupServiceAddon] = siteApi.useGroupServiceAddonsServicePostGroupServiceAddonMutation();
  const [deleteGroupServiceAddon] = siteApi.useGroupServiceAddonsServiceDeleteGroupServiceAddonMutation();
  const [patchService] = siteApi.useServiceServicePatchServiceMutation();
  const [postService] = siteApi.useServiceServicePostServiceMutation();
  const [postGroupService] = siteApi.useGroupServiceServicePostGroupServiceMutation();

  const handleSubmit = useCallback(() => {
    if (!newService.name || !Object.keys(newService?.tiers || {}).length) {
      setSnack({ snackOn: 'Provide the service name and at least 1 tier with at least 1 feature.', snackType: 'info' });
      return;
    }

    if (!editGroup) {
      if (newService?.id) {
        patchService({ patchServiceRequest: { service: newService } }).unwrap().then(() => {
          closeModal && closeModal({ ...newService });
        }).catch(console.error);
      } else {
        postService({ postServiceRequest: { service: newService } }).unwrap().then(async ({ id: serviceId }) => {
          await postGroupService({ postGroupServiceRequest: { serviceId } }).unwrap();
          closeModal && closeModal({ ...newService, id: serviceId });
        }).catch(console.error);
      }
    } else {
      closeModal && closeModal(newService);
    }
  }, [newService]);

  useEffect(() => {
    if (group?.purpose) {
      void suggestServices({ id: IPrompts.SUGGEST_SERVICE, prompt: group.purpose });
    }
  }, [group?.purpose]);

  useEffect(() => {
    if (existingServiceRequest && existingServiceRequest.service) {
      setNewService({ ...newService, ...existingServiceRequest.service });
    }
  }, [existingServiceRequest]);

  return <>
    <Dialog open={dialog === 'manage_form'} fullWidth maxWidth="lg">
      <Suspense>
        <ManageFormModal {...props} closeModal={() => {
          setDialog('')
          void getGroupForms();
        }} />
      </Suspense>
    </Dialog>

    <Card>
      <CardHeader title={`${editService ? 'Edit' : 'Create'} Service`} />
      <CardContent>
        <Grid container>
          <Grid item xs={12}>
            <Box mb={4}>
              <TextField
                fullWidth
                label="Name"
                value={newService.name}
                onChange={e => setNewService({ ...newService, name: e.target.value })}
                onBlur={() => {
                  if (!newService.name || !group?.displayName) return;
                  void suggestTiers({ id: IPrompts.SUGGEST_TIER, prompt: `${newService.name.toLowerCase()} at ${group?.displayName}` });
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

            <Box mb={4}>
              <Link style={{ cursor: 'pointer' }} onClick={() => setHasServiceForm(!hasServiceForm)}>
                {hasServiceForm ? '- Remove' : '+ Add'} Service Forms
              </Link>
            </Box>
            {!hasServiceForm ? <>


            </> : <>

              {groupFormsLoaded && <Box mb={4}>
                <TextField
                  select
                  fullWidth
                  value={newService.formId || 'unset'}
                  label="Intake Form"
                  helperText="Optional. Shown during appointment creation."
                  onChange={e => {
                    e.target.value && setNewService({ ...newService, formId: e.target.value === 'unset' ? '' : e.target.value });
                  }}
                >
                  <MenuItem key="create-form" onClick={() => setDialog('manage_form')}>Add a form to this list</MenuItem>
                  <MenuItem key="unset-selection" value="unset">&nbsp;</MenuItem>
                  {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>)}
                </TextField>
              </Box>}

              {groupFormsLoaded && <Box>
                <TextField
                  select
                  fullWidth
                  value={newService.surveyId || 'unset'}
                  label="Survey Form"
                  helperText="Optional. Shown during post-appointment summary."
                  onChange={e => {
                    e.target.value && setNewService({ ...newService, surveyId: e.target.value === 'unset' ? '' : e.target.value });
                  }}
                >
                  <MenuItem key="create-form" onClick={() => setDialog('manage_form')}>Add a form to this list</MenuItem>
                  <MenuItem key="unset-selection" value="unset">&nbsp;</MenuItem>
                  {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>)}
                </TextField>
              </Box>}
            </>}
          </Grid>
        </Grid>
      </CardContent>

      <CardHeader title="Add Tiers" />

      <CardContent sx={{ padding: '0 15px' }}>
        <Grid container>
          <Grid item xs={12}>
            <Box mb={4}>
              <TextField
                fullWidth
                label="Name"
                value={newServiceTier.name}
                onChange={e => setNewServiceTier({ ...newServiceTier, name: e.target.value })}
                onBlur={() => {
                  if (!newServiceTier.name || !newService.name) return;
                  void suggestAddon({ id: IPrompts.SUGGEST_FEATURE, prompt: `${newServiceTier.name} ${newService.name}` });
                }}
                helperText={
                  <TierSuggestions
                    staticSuggestions='Ex: Basic, Mid-Tier, Advanced'
                    handleSuggestion={suggestedTier => {
                      void suggestAddon({ id: IPrompts.SUGGEST_FEATURE, prompt: `${suggestedTier} ${newService.name}` });
                      setNewServiceTier({ ...newServiceTier, name: suggestedTier })
                    }}
                  />
                }
              />
            </Box>

            <Box mb={4} flexDirection="column" sx={{ display: 'flex', alignItems: 'baseline' }}>
              <SelectLookup
                multiple
                lookupName='Feature'
                lookups={groupServiceAddonsRequest?.groupServiceAddons?.map(gsa => gsa.serviceAddon)}
                lookupValue={serviceTierAddonIds}
                helperText={
                  <AddonSuggestions
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
                lookupChange={(val: string[]) => {
                  const gsa = groupServiceAddonsRequest?.groupServiceAddons?.filter(s => val.includes(s.serviceAddon?.id || "")).map(s => s.serviceAddon?.id || "");
                  if (gsa) setServiceTierAddonIds(gsa);
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

            <Box mb={4}>
              <Link style={{ cursor: 'pointer' }} onClick={() => setHasTierForm(!hasTierForm)}>
                {hasTierForm ? '- Remove' : '+ Add'} Tier Forms
              </Link>
            </Box>
            {!hasTierForm ? <>


            </> : <>

              {groupFormsLoaded && <Box mb={4}>
                <TextField
                  select
                  fullWidth
                  value={newServiceTier.formId}
                  label="Intake Form"
                  helperText="Optional. Shown during appointment creation."
                  onChange={e => {
                    e.target.value && setNewServiceTier({ ...newServiceTier, formId: e.target.value === 'unset' ? '' : e.target.value });
                  }}
                >
                  <MenuItem key="create-form" onClick={() => setDialog('manage_form')}>Add a form to this list</MenuItem>
                  <MenuItem key="unset-selection" value="unset">&nbsp;</MenuItem>
                  {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>)}
                </TextField>
              </Box>}

              {groupFormsLoaded && <Box mb={4}>
                <TextField
                  select
                  fullWidth
                  value={newServiceTier.surveyId}
                  label="Survey Form"
                  helperText="Optional. Shown during post-appointment summary."
                  onChange={e => {
                    e.target.value && setNewServiceTier({ ...newServiceTier, surveyId: e.target.value === 'unset' ? '' : e.target.value });
                  }}
                >
                  <MenuItem key="create-form" onClick={() => setDialog('manage_form')}>Add a form to this list</MenuItem>
                  <MenuItem key="unset-selection" value="unset">&nbsp;</MenuItem>
                  {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>)}
                </TextField>
              </Box>}
            </>}

            {/* <Box>
              <Typography variant="h6">Multiplier</Typography>
              <Box sx={{ display: 'flex', alignItems: 'baseline' }}>
                <span>{newServiceTier.multiplier}x <span>&nbsp;</span> &nbsp;</span>
                <Slider value={parseFloat(newServiceTier.multiplier)} onChange={(e, val) => setNewServiceTier({ ...newServiceTier, multiplier: parseFloat(val.toString()).toFixed(2) })} step={.01} min={1} max={5} />
              </Box>
            </Box> */}
          </Grid>
        </Grid>
      </CardContent>
      <CardActionArea onClick={() => {
        if (newServiceTier.name && serviceTierAddonIds.length && newService.tiers && !Object.values(newService.tiers).some(ti => ti.name === newServiceTier.name)) {
          const created = (new Date()).getTime().toString();
          newServiceTier.id = created;
          newServiceTier.createdOn = created;
          newServiceTier.order = Object.keys(newService.tiers).length + 1;
          newServiceTier.addons = serviceTierAddonIds.reduce((m, id, i) => {
            const addon = groupServiceAddonsRequest?.groupServiceAddons?.find(gs => gs.serviceAddon?.id === id)?.serviceAddon || { name: '' };
            return {
              ...m,
              [id]: addon && {
                id,
                name: addon.name,
                order: i + 1
              }
            }
          }, {});
          newService.tiers[newServiceTier.id] = newServiceTier;
          setNewServiceTier({ ...serviceTierSchema });
          setServiceTierAddonIds([]);
          setNewService({ ...newService });
        } else {
          void setSnack({ snackOn: 'Provide a unique tier name and at least 1 feature.', snackType: 'info' });
        }
      }}>
        <Box m={2} sx={{ display: 'flex', alignItems: 'center' }}>
          <Typography color="secondary" variant="button">Add Tier to Service</Typography>
        </Box>
      </CardActionArea>
      {newService.tiers && Object.keys(newService.tiers).length > 0 ? <>
        <CardHeader title="Tiers" subheader='The following table will be shown during booking.' />
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'flex-end', flexWrap: 'wrap' }}>
            {Object.values(newService.tiers).sort((a, b) => new Date(a.createdOn!).getTime() - new Date(b.createdOn!).getTime()).map((tier, i) => {
              return <Box key={`service-tier-chip${i + 1}new`} m={1}>
                <Chip
                  sx={classes.chipRoot}
                  label={
                    <Typography sx={classes.chipLabel}>
                      {`#${i + 1} ` + tier.name + ' (' + (tier.multiplier || 100) / 100 + 'x)'}
                    </Typography>
                  }
                  onDelete={() => {
                    const tiers = { ...newService.tiers };
                    if (tier.id) {
                      delete tiers[tier.id];
                      setNewService({ ...newService, tiers });
                    }
                  }}
                />
              </Box>
            })}
          </Box>
        </CardContent>
        <CardContent sx={{ padding: '0 15px' }}>
          <Suspense>
            <ServiceTierAddons service={newService} />
          </Suspense>
        </CardContent>
      </> : <CardContent>No tiers added yet.</CardContent>}

      <CardActions>

        <Grid container justifyContent={showCancel ? "space-between" : "flex-end"}>
          {showCancel && <Button onClick={closeModal}>Cancel</Button>}
          <Button disabled={!newService.name || newService.tiers && !Object.keys(newService.tiers).length} onClick={handleSubmit}>Save Service</Button>
        </Grid>
      </CardActions>
    </Card>
  </>;
}

export default ManageServiceModal;
