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
import Typography from '@mui/material/Typography';

import { siteApi, useUtil, useSuggestions, IPrompts, IGroup, IGroupRole, targets, IValidationAreas, useValid } from 'awayto/hooks';

interface ManageGroupRolesModalProps extends IComponent {
  editGroup?: IGroup;
  validArea?: keyof IValidationAreas;
  showCancel?: boolean;
  saveToggle?: number;
}

export function ManageGroupRolesModal({ children, editGroup, validArea, saveToggle = 0, showCancel = true, closeModal }: ManageGroupRolesModalProps): React.JSX.Element {

  const { setSnack } = useUtil();
  const { setValid } = useValid();

  const [groupRoleInput, setGroupRoleInput] = useState('');
  const [defaultRoleId, setDefaultRoleId] = useState(editGroup?.defaultRoleId || '');

  const {
    comp: RoleSuggestions,
    suggest: suggestRoles
  } = useSuggestions('group_roles');

  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const [patchGroupRoles] = siteApi.useGroupRoleServicePatchGroupRolesMutation();
  const [postGroupRole] = siteApi.useGroupRoleServicePostGroupRoleMutation();
  const [deleteGroupRole] = siteApi.useGroupRoleServiceDeleteGroupRoleMutation();

  const handleSubmitRole = useCallback((name?: string) => {
    if (name?.length) {
      postGroupRole({ postGroupRoleRequest: { name } });
      setGroupRoleInput('');
    }
  }, []);

  const handleSubmitRoles = useCallback(() => {
    if (!groupRolesRequest?.groupRoles.length || !defaultRoleId) {
      setSnack({ snackType: 'error', snackOn: 'All fields are required.' });
      return;
    }

    const newRoles = groupRolesRequest.groupRoles.reduce((m, groupRole) => {
      if (!groupRole.id) return m;
      return {
        ...m,
        [groupRole.id]: groupRole
      };
    }, {} as Record<string, IGroupRole>);

    void patchGroupRoles({ patchGroupRolesRequest: { roles: newRoles, defaultRoleId } }).unwrap().then(() => {
      closeModal && closeModal({ roles: newRoles, defaultRoleId });
    });
  }, [editGroup, defaultRoleId, groupRolesRequest?.groupRoles]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmitRoles();
    }
  }, [saveToggle]);

  // Onboarding handling
  useEffect(() => {
    if (validArea) {
      setValid({ area: validArea, schema: 'roles', valid: Boolean(defaultRoleId.length && groupRolesRequest?.groupRoles.length) });
    }
  }, [validArea, defaultRoleId, groupRolesRequest?.groupRoles]);

  useEffect(() => {
    if (editGroup?.name && editGroup?.purpose) {
      suggestRoles({ id: IPrompts.SUGGEST_ROLE, prompt: `${editGroup.name}!$${editGroup.purpose}` });
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
                    const existingId = groupRolesRequest?.groupRoles?.find(r => r.name === suggestedRole)?.id;

                    if (!existingId) {
                      handleSubmitRole(suggestedRole);
                    }
                  }}
                />
              }
              onKeyDown={e => 'Enter' === e.key && handleSubmitRole(groupRoleInput)}
              onChange={e => setGroupRoleInput(e.target.value)}
              value={groupRoleInput}
              slotProps={{
                input: {
                  endAdornment: <Button
                    {...targets(`manage group roles modal add role`, `add the typed role to the group`)}
                    variant="text"
                    color="secondary"
                    onClick={_ => handleSubmitRole(groupRoleInput)}
                  >Add</Button>
                }
              }}
            />
            {groupRolesRequest?.groupRoles?.length && <Box ml={2} mt={2}>
              <Grid container spacing={2} size="grow">
                {groupRolesRequest?.groupRoles.map((gr, i) =>
                  <Chip
                    key={`group-role-selection-${i}`}
                    {...targets(`group roles delete ${i}`, gr.name, `remove ${gr.name} from the list of group roles`)}
                    color="secondary"
                    onClick={() => {
                      if (gr.id?.length) {
                        deleteGroupRole({ ids: gr.id });
                      }
                    }}
                  />
                )}
              </Grid>
              <Typography variant="caption">Click to remove</Typography>
            </Box>}
          </Grid>
          {groupRolesRequest?.groupRoles && groupRolesRequest?.groupRoles.length && <Grid size={12}>
            <TextField
              {...targets(`manage group roles modal default role selection`, `Default Role`, `select the group's default role from the list`)}
              select
              fullWidth
              helperText={'Set the group default role. When members join the group, this role will be assigned.'}
              required
              onChange={e => setDefaultRoleId(e.target.value)}
              value={defaultRoleId}
              slotProps={{
                input: {
                  endAdornment: <Button
                    {...targets(`manage group roles modal update default role`, `update the default group role to the selected one`)}
                    variant="text"
                    color="secondary"
                    sx={{ pr: 4 }}
                    onClick={_ => {
                      patchGroupRoles({ patchGroupRolesRequest: { defaultRoleId } });
                    }}
                  >Update</Button>
                }
              }}
            >
              {groupRolesRequest?.groupRoles.map(groupRole => {
                return <MenuItem key={`${groupRole.roleId}_primary_role_select`} value={groupRole.roleId}>
                  {groupRole.name}
                </MenuItem>
              })}
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
