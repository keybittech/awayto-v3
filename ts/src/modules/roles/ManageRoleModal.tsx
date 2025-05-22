import React, { useState, useCallback } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';

import { useUtil, siteApi, IGroupRole, targets } from 'awayto/hooks';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';

interface ManageRoleModalProps extends IComponent {
  editRole?: IGroupRole;
  isDefault?: boolean;
}

export function ManageRoleModal({ editRole, isDefault, closeModal }: ManageRoleModalProps): React.JSX.Element {
  const { setSnack } = useUtil();
  const [patchGroupRole] = siteApi.useGroupRoleServicePatchGroupRoleMutation();
  const [postGroupRole] = siteApi.useGroupRoleServicePostGroupRoleMutation();

  const [role, setRole] = useState({
    name: '',
    ...editRole
  } as Required<IGroupRole>);

  const [defaultRole, setDefaultRole] = useState(isDefault);

  const handleSubmit = useCallback(async () => {
    const { id, name } = role;

    if (!name) {
      setSnack({ snackType: 'error', snackOn: 'Roles must have a name.' });
      return;
    }

    if (id) {
      await patchGroupRole({ patchGroupRoleRequest: { roleId: id, name, defaultRole } }).unwrap().catch(console.error);
    } else {
      await postGroupRole({ postGroupRoleRequest: { name, defaultRole } }).unwrap().catch(console.error);
    }

    if (closeModal)
      closeModal();
  }, [role, defaultRole]);

  return <>
    <Card>
      <CardContent>
        <Typography variant="button">Manage role</Typography>
        <Grid container direction="row" spacing={2}>
          <Grid size={12}>
            <Grid container direction="column" spacing={4} justifyContent="space-evenly" >
              <Grid>
                <Typography variant="h6">Role</Typography>
              </Grid>
              <Grid>
                <TextField
                  {...targets(`manage role modal name`, `Role Name`, `edit the role's name`)}
                  fullWidth
                  autoFocus
                  value={role.name}
                  onKeyDown={e => {
                    if ('Enter' === e.key) {
                      void handleSubmit();
                    }
                  }}
                  onChange={e => setRole({ ...role, name: e.target.value })} />
              </Grid>
              <Grid>
                <FormGroup>
                  <FormControlLabel
                    {...targets(`manage role modal default`, `Use as Default`, `toggle whether this role should be applied to new members of the group automatically`)}
                    control={
                      <Checkbox
                        checked={defaultRole}
                        onChange={() => setDefaultRole(!defaultRole)}
                      />
                    }
                  />
                </FormGroup>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent="space-between">
          <Button
            {...targets(`manage role modal close`, `close the role details editing modal`)}
            onClick={closeModal}
          >Cancel</Button>
          <Button
            {...targets(`manage role modal submit`, `submit the role details for editing or creation`)}
            onClick={() => void handleSubmit()}
          >Submit</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default ManageRoleModal;
