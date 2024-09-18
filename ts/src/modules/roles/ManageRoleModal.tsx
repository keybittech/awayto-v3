import React, { useState, useCallback } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

import { useUtil, siteApi, IRole } from 'awayto/hooks';

declare global {
  interface IComponent {
    editRole?: IRole;
  }
}

export function ManageRoleModal({ editRole, closeModal }: IComponent): React.JSX.Element {
  const { setSnack } = useUtil();
  const [patchRole] = siteApi.useRoleServicePatchRoleMutation();
  const [postRole] = siteApi.useRoleServicePostRoleMutation();
  const [postGroupRole] = siteApi.useGroupRoleServicePostGroupRoleMutation();

  const [role, setRole] = useState({
    name: '',
    ...editRole
  } as Required<IRole>);

  const handleSubmit = useCallback(async () => {
    const { id, name } = role;

    if (!name) {
      setSnack({ snackType: 'error', snackOn: 'Roles must have a name.' });
      return;
    }

    if (id) {
      await patchRole({ patchRoleRequest: role }).unwrap();
    } else {
      const { id } = await postRole({ postRoleRequest: role }).unwrap();
      await postGroupRole({ postGroupRoleRequest: { role: { id, name } } }).unwrap();
    }

    if (closeModal)
      closeModal();
  }, [role]);

  return <>
    <Card>
      <CardContent>
        <Typography variant="button">Manage role</Typography>
        <Grid container direction="row" spacing={2}>
          <Grid item xs={12}>
            <Grid container direction="column" spacing={4} justifyContent="space-evenly" >
              <Grid item>
                <Typography variant="h6">Role</Typography>
              </Grid>
              <Grid item>
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
