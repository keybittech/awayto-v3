import React, { useState, useCallback, useMemo, useEffect } from 'react';

import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';

import { siteApi, useUtil, useSuggestions, IPrompts, IGroup, IRole, targets, IValidationAreas, useValid } from 'awayto/hooks';
import SelectLookup from '../common/SelectLookup';

interface ManageGroupRolesModalProps extends IComponent {
  editGroup?: IGroup;
  validArea?: keyof IValidationAreas;
  showCancel?: boolean;
  saveToggle?: number;
}

export function ManageGroupRolesModal({ children, editGroup, validArea, saveToggle = 0, showCancel = true, closeModal, ...props }: ManageGroupRolesModalProps): React.JSX.Element {

  console.log({ editGroup });

  const { setSnack } = useUtil();
  const { setValid } = useValid();

  const {
    comp: RoleSuggestions,
    suggest: suggestRoles
  } = useSuggestions('group_roles');

  const { data: profileRequest, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const [patchGroupRoles] = siteApi.useGroupRoleServicePatchGroupRolesMutation();
  const [postRole] = siteApi.useRoleServicePostRoleMutation();
  const [deleteRole] = siteApi.useRoleServiceDeleteRoleMutation();

  const [roleIds, setRoleIds] = useState<string[]>(Object.keys(editGroup?.roles || {}).filter(rid => editGroup?.roles && editGroup.roles[rid].name !== 'Admin'));
  const [defaultRoleId, setDefaultRoleId] = useState(editGroup?.defaultRoleId && roleIds.includes(editGroup.defaultRoleId) ? editGroup.defaultRoleId : '');

  const roleValues = useMemo(() => Object.values(profileRequest?.userProfile?.roles || {}), [profileRequest]);

  const newRoles = useMemo(() => Object.fromEntries(roleIds.map(id => [id, roleValues.find(r => r.id === id)] as [string, IRole])), [roleIds, roleValues]);

  const handleSubmit = useCallback(() => {
    if (!roleIds.length || !defaultRoleId) {
      setSnack({ snackType: 'error', snackOn: 'All fields are required.' });
      return;
    }

    void patchGroupRoles({ patchGroupRolesRequest: { roles: newRoles, defaultRoleId } }).unwrap().then(() => {
      closeModal && closeModal({ roles: newRoles, defaultRoleId });
    });
  }, [roleIds, newRoles, defaultRoleId]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  // Onboarding handling
  useEffect(() => {
    if (validArea) {
      setValid({ area: validArea, schema: 'roles', valid: Boolean(defaultRoleId.length && roleIds.length) });
    }
  }, [validArea, defaultRoleId, roleIds]);

  useEffect(() => {
    if (editGroup?.name && editGroup?.purpose) {
      void suggestRoles({ id: IPrompts.SUGGEST_ROLE, prompt: `${editGroup.name}!$${editGroup.purpose}` });
    }
  }, [editGroup]);

  return <>
    <Card>
      <CardHeader title="Edit Roles"></CardHeader>
      <CardContent>
        {!!children && children}

        <Grid container spacing={4}>
          <Grid size={12}>
            <SelectLookup
              multiple
              required
              helperText={
                <RoleSuggestions
                  staticSuggestions='Ex: Consultant, Project Manager, Advisor, Business Analyst'
                  handleSuggestion={suggestedRole => {
                    // The currently suggested role in the user detail's role list
                    const existingId = roleValues.find(r => r.name === suggestedRole)?.id;

                    // If the role is not in the user detail roles list, or it is, but it doesn't exist in the current list, continue
                    if (!existingId || (existingId && !roleIds.includes(existingId))) {

                      // If the role is in the user details roles list
                      if (existingId) {
                        setRoleIds([...roleIds, existingId])
                      } else {
                        postRole({ postRoleRequest: { name: suggestedRole } }).unwrap().then(async ({ id: newRoleId }) => {
                          await getUserProfileDetails();
                          !roleIds.includes(newRoleId) && setRoleIds([...roleIds, newRoleId]);
                        }).catch(console.error);
                      }
                    }
                  }}
                />
              }
              lookupName='Group Role'
              lookups={roleValues}
              lookupChange={(newIds: string[]) => {
                if (!newIds.length || !newIds.includes(defaultRoleId)) {
                  setDefaultRoleId('');
                }
                setRoleIds(newIds);
              }}
              lookupValue={roleIds}
              invalidValues={['admin']}
              refetchAction={getUserProfileDetails}
              createAction={async ({ name }) => {
                return await postRole({ postRoleRequest: { name } }).unwrap();
              }}
              deleteAction={async ({ ids }) => {
                await deleteRole({ ids }).unwrap();
              }}
              deleteActionIdentifier='ids'
              {...props}
            />
          </Grid>
          {!!roleIds.length && <Grid size={12}>
            <TextField
              {...targets(`manage group roles modal default role selection`, `Default Role`, `select the group's default role from the list`)}
              select
              fullWidth
              helperText={'Set the group default role. When members join the group, this role will be assigned.'}
              required
              onChange={e => setDefaultRoleId(e.target.value)}
              value={defaultRoleId}
            >
              {roleIds.map(roleId => <MenuItem key={`${roleId}_primary_role_select`} value={roleId}>{roleValues.find(role => role.id === roleId)?.name || ''}</MenuItem>)}
            </TextField>
          </Grid>}
          <Grid size="grow">
            <Alert severity="info">Your Admin role is created automatically. Only create roles for your members.</Alert>
          </Grid>
        </Grid>
      </CardContent>
      {validArea != 'onboarding' && <CardActions>
        <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
          {showCancel && <Button
            {...targets(`manage group roles modal close`, `close the group role management modal`)}
            onClick={closeModal}
          >Cancel</Button>}
          <Button
            {...targets(`manage group roles modal submit`, `submit the current list of group roles for editing or creation`)}
            disabled={!defaultRoleId}
            onClick={handleSubmit}
          >Save Roles</Button>
        </Grid>
      </CardActions>}
    </Card>
  </>
}

export default ManageGroupRolesModal;
