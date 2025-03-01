import React, { useState, useCallback } from 'react';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';

import { siteApi, targets, useUtil } from 'awayto/hooks';
import { TextField } from '@mui/material';

export function JoinGroupModal({ closeModal }: IComponent): React.JSX.Element {

  const { setSnack } = useUtil();
  const [joinGroup] = siteApi.useGroupUtilServiceJoinGroupMutation();
  const [code, setCode] = useState('');

  const handleSubmit = useCallback(() => {
    if (!code) {
      setSnack({ snackType: 'error', snackOn: 'Please provide at least 1 code.' });
      return;
    }

    joinGroup({ joinGroupRequest: { code } }).unwrap().then(() => {
      if (closeModal)
        closeModal();
    }).catch(console.error);
  }, [code]);

  return <>
    <Card>
      <CardContent>
        <Typography variant="button">Join a Group</Typography>
      </CardContent>
      <CardContent>
        <Grid container>
          <Grid size={12}>
            <Grid container>
              <Grid size={12}>
                <TextField
                  {...targets(`join group input code`, `Code`, `input the group code for the group you want to join`)}
                  type="code"
                  placeholder="Type a code and press enter..."
                  fullWidth
                  value={code}
                  onChange={e => {
                    setCode(e.currentTarget.value)
                  }}
                />
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>
        <Grid container justifyContent="flex-end">
          <Button
            {...targets(`join group modal close`, `close the join group modal`)}
            onClick={closeModal}
          >Cancel</Button>
          <Button
            {...targets(`join group modal submit`, `submit the form to join a group using its group code`)}
            onClick={handleSubmit}
          >Submit</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default JoinGroupModal;
