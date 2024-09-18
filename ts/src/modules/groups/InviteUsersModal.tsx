import React, { useState, useCallback } from 'react';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';

import { siteApi, useUtil, UserEmail } from 'awayto/hooks';
import { TextField } from '@mui/material';

export function InviteUsersModal({ closeModal }: IProps): React.JSX.Element {

  const { setSnack } = useUtil();
  const [inviteGroupUser] = siteApi.useGroupServiceInviteGroupUsersMutation();

  const [email, setEmail] = useState('');
  const [users, setUsers] = useState<UserEmail[]>([]);

  const handleAdd = useCallback(() => {
    setUsers([...users, { email } as UserEmail]);
    setEmail('');
  }, [users, email]);

  const handleSubmit = useCallback(() => {
    if (!users.length) {
      setSnack({ snackType: 'error', snackOn: 'Please provide at least 1 email.' });
      return;
    }

    inviteGroupUser({ inviteGroupUsersRequest: { users } }).unwrap().then(() => {
      if (closeModal)
        closeModal();
    }).catch(console.error);
  }, [users]);

  return <>
    <Card>
      <CardContent>
        <Typography variant="button">Invite Users</Typography>
      </CardContent>
      <CardContent>
        <Grid container>
          <Grid item xs={12}>
            <Grid container>
              <Grid item xs={12}>
                <TextField
                  label="Email"
                  type="email"
                  placeholder="Type an email and press enter..."
                  fullWidth
                  value={email}
                  onChange={e => {
                    setEmail(e.currentTarget.value)
                  }}
                  InputProps={{
                    onKeyDown: e => {
                      if ('Enter' === e.key && e.currentTarget.validity.valid) {
                        handleAdd();
                      }
                    }
                  }}
                />
              </Grid>
              <Grid item xs={12}>
                <ul>
                  {users.map(({ email }, i) => <li key={`group_invite_email_${i}`}>{email}</li>)}
                </ul>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent="flex-end">
          <Button onClick={closeModal}>Cancel</Button>
          <Button onClick={handleSubmit}>Submit</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default InviteUsersModal;
