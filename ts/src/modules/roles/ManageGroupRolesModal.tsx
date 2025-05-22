import React, { useState, useCallback, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';

import { siteApi, useUtil, useSuggestions, IPrompts, IGroup, IGroupRole, targets, IValidationAreas, useValid } from 'awayto/hooks';
import SelectLookup from '../common/SelectLookup';

interface ManageGroupRolesModalProps extends IComponent {
  editGroup?: IGroup;
  validArea?: keyof IValidationAreas;
  showCancel?: boolean;
  saveToggle?: number;
}

export function ManageGroupRolesModal({ children, editGroup, validArea, saveToggle = 0, showCancel = true, closeModal }: ManageGroupRolesModalProps): React.JSX.Element {

  const { setSnack } = useUtil();
  const { setValid } = useValid();

  const [groupRole, setGroupRole] = useState('');
  const [groupRoleIds, setGroupRoleIds] = useState<string[]>(Object.keys(editGroup?.roles || {}).filter(rid => editGroup?.roles && editGroup.roles[rid].name !== 'Admin'));
  const [defaultRoleId, setDefaultRoleId] = useState(editGroup?.defaultRoleId && groupRoleIds.includes(editGroup.defaultRoleId) ? editGroup.defaultRoleId : '');

  const {
    comp: RoleSuggestions,
    suggest: suggestRoles
  } = useSuggestions('group_roles');

  const { data: groupRolesRequest, refetch: getGroupRoles } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const [patchGroupRoles] = siteApi.useGroupRoleServicePatchGroupRolesMutation();
  const [postGroupRole] = siteApi.useGroupRoleServicePostGroupRoleMutation();
  const [deleteGroupRole] = siteApi.useGroupRoleServiceDeleteGroupRoleMutation();

  const handleSubmitRole = useCallback(() => {
    if (groupRole.length) {
      postGroupRole({ postGroupRoleRequest: { name: groupRole } }).unwrap().then(() => {
        setGroupRole('');
      });
    }
  }, [groupRole]);

  const handleSubmitRoles = useCallback(() => {
    if (!groupRoleIds.length || !defaultRoleId) {
      setSnack({ snackType: 'error', snackOn: 'All fields are required.' });
      return;
    }

    const newRoles = Object.fromEntries(groupRoleIds.map(id => [id, groupRolesRequest?.groupRoles.find(r => r.id === id)] as [string, IGroupRole]));

    void patchGroupRoles({ patchGroupRolesRequest: { roles: newRoles, defaultRoleId } }).unwrap().then(() => {
      closeModal && closeModal({ roles: newRoles, defaultRoleId });
    });
  }, [groupRoleIds, defaultRoleId]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmitRoles();
    }
  }, [saveToggle]);

  // Onboarding handling
  useEffect(() => {
    if (validArea) {
      setValid({ area: validArea, schema: 'roles', valid: Boolean(defaultRoleId.length && groupRoleIds.length) });
    }
  }, [validArea, defaultRoleId, groupRoleIds]);

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
            <TextField
              {...targets(`group role entry`, `Add Group Role`, `input a role to be assigned to group members`)}
              fullWidth
              helperText={
                <RoleSuggestions
                  staticSuggestions='Ex: Consultant, Project Manager, Advisor, Business Analyst'
                  handleSuggestion={suggestedRole => {
                    // The currently suggested role in the user detail's role list
                    const existingId = groupRolesRequest?.groupRoles.find(r => r.name === suggestedRole)?.id;

                    // If the role is not in the user detail roles list, or it is, but it doesn't exist in the current list, continue
                    if (!existingId || (existingId && !groupRoleIds.includes(existingId))) {

                      // If the role is in the user details roles list
                      if (existingId) {
                        setGroupRoleIds([...groupRoleIds, existingId])
                      } else {
                        postGroupRole({ postGroupRoleRequest: { name: suggestedRole } }).unwrap().then(async ({ groupRoleId }) => {
                          await getGroupRoles();
                          !groupRoleIds.includes(groupRoleId) && setGroupRoleIds([...groupRoleIds, groupRoleId]);
                        }).catch(console.error);
                      }
                    }
                  }}
                />
              }
              onKeyDown={(e) => 'Enter' === e.key && handleSubmitRole()}
              onChange={e => setGroupRole(e.target.value)}
              value={groupRole}
              slotProps={{
                input: {
                  endAdornment: <Button
                    {...targets(`manage group roles modal add role`, `add the typed role to the group`)}
                    variant="text"
                    color="secondary"
                    onClick={() => handleSubmitRole}
                  >Add</Button>
                }
              }}
            />
            <Grid container>
              {groupRolesRequest?.groupRoles && groupRolesRequest?.groupRoles.map((gr, i) => <Box key={`group-role-selection-${i}`} mt={2} mr={2}>
                <Chip
                  {...targets(`group roles delete ${i}`, gr.name, `remove ${gr.name} from the list of group roles`)}
                  color="secondary"
                  onDelete={() => {
                    if (gr.id?.length) {
                      deleteGroupRole({ ids: gr.id });
                    }
                  }}
                />
              </Box>)}
            </Grid>




            {/* <SelectLookup */}
            {/*   multiple */}
            {/*   required */}
            {/*   helperText={ */}
            {/*     <RoleSuggestions */}
            {/*       staticSuggestions='Ex: Consultant, Project Manager, Advisor, Business Analyst' */}
            {/*       handleSuggestion={suggestedRole => { */}
            {/*         // The currently suggested role in the user detail's role list */}
            {/*         const existingId = groupRolesRequest?.groupRoles.find(r => r.name === suggestedRole)?.id; */}
            {/**/}
            {/*         // If the role is not in the user detail roles list, or it is, but it doesn't exist in the current list, continue */}
            {/*         if (!existingId || (existingId && !groupRoleIds.includes(existingId))) { */}
            {/**/}
            {/*           // If the role is in the user details roles list */}
            {/*           if (existingId) { */}
            {/*             setGroupRoleIds([...groupRoleIds, existingId]) */}
            {/*           } else { */}
            {/*             postGroupRole({ postGroupRoleRequest: { name: suggestedRole } }).unwrap().then(async ({ groupRoleId }) => { */}
            {/*               await getGroupRoles(); */}
            {/*               !groupRoleIds.includes(groupRoleId) && setGroupRoleIds([...groupRoleIds, groupRoleId]); */}
            {/*             }).catch(console.error); */}
            {/*           } */}
            {/*         } */}
            {/*       }} */}
            {/*     /> */}
            {/*   } */}
            {/*   lookupName='Group Role' */}
            {/*   lookups={groupRolesRequest?.groupRoles} */}
            {/*   lookupChange={(newIds: string[]) => { */}
            {/*     if (!newIds.length || !newIds.includes(defaultRoleId)) { */}
            {/*       setDefaultRoleId(''); */}
            {/*     } */}
            {/*     setGroupRoleIds(newIds); */}
            {/*   }} */}
            {/*   lookupValue={groupRoleIds} */}
            {/*   invalidValues={['admin']} */}
            {/*   refetchAction={getGroupRoles} */}
            {/*   createAction={async ({ name }) => { */}
            {/*     return await postGroupRole({ postGroupRoleRequest: { name } }).unwrap(); */}
            {/*   }} */}
            {/*   deleteAction={async ({ ids }) => { */}
            {/*     await deleteGroupRole({ ids }).unwrap(); */}
            {/*   }} */}
            {/*   deleteActionIdentifier='ids' */}
            {/*   {...props} */}
            {/* /> */}
          </Grid>
          {!!groupRoleIds.length && <Grid size={12}>
            <TextField
              {...targets(`manage group roles modal default role selection`, `Default Role`, `select the group's default role from the list`)}
              select
              fullWidth
              helperText={'Set the group default role. When members join the group, this role will be assigned.'}
              required
              onChange={e => setDefaultRoleId(e.target.value)}
              value={defaultRoleId}
            >
              {groupRoleIds.map(roleId =>
                <MenuItem key={`${roleId}_primary_role_select`} value={roleId}>
                  {groupRolesRequest?.groupRoles.find(role => role.id === roleId)?.name || ''}
                </MenuItem>
              )}
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
            onClick={handleSubmitRoles}
          >Save Roles</Button>
        </Grid>
      </CardActions>}
    </Card>
  </>
}

export default ManageGroupRolesModal;
