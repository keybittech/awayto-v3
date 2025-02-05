import React, { useState, useCallback, useMemo } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';

import { useUtil, siteApi, IRole } from 'awayto/hooks';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';

declare global {
  interface IComponent {
    editRole?: IRole;
    isDefault?: boolean;
  }
}

export function ManageRoleModal({ editRole, isDefault, closeModal }: IComponent): React.JSX.Element {
  const { setSnack } = useUtil();
  const [postRole] = siteApi.useRoleServicePostRoleMutation();
  const [patchGroupRole] = siteApi.useGroupRoleServicePatchGroupRoleMutation();
  const [postGroupRole] = siteApi.useGroupRoleServicePostGroupRoleMutation();

  const [role, setRole] = useState({
    name: '',
    ...editRole
  } as Required<IRole>);

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
      const { id } = await postRole({ postRoleRequest: role }).unwrap();
      await postGroupRole({ postGroupRoleRequest: { roleId: id, name, defaultRole } }).unwrap().catch(console.error);
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
                  fullWidth
                  autoFocus
                  id="name"
                  label="Name"
                  name="name"
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
                    id={`default-role-selection`}
                    control={
                      <Checkbox
                        checked={defaultRole}
                        onChange={() => setDefaultRole(!defaultRole)}
                      />
                    }
                    label="Use as Default Role"
                  />
                </FormGroup>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent="space-between">
          <Button onClick={closeModal}>Cancel</Button>
          <Button onClick={() => void handleSubmit()}>Submit</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default ManageRoleModal;
