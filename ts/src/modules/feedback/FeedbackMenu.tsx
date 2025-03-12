import React, { useCallback, useState } from 'react';

import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Grid from '@mui/material/Grid';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import CampaignIcon from '@mui/icons-material/Campaign';

import { siteApi, targets, useStyles, useUtil } from 'awayto/hooks';

export function FeedbackMenu(_: IComponent): React.JSX.Element {

  const classes = useStyles();

  const { setSnack } = useUtil();

  const [postGroupFeedback] = siteApi.useGroupFeedbackServicePostGroupFeedbackMutation();
  const [postSiteFeedback] = siteApi.useFeedbackServicePostSiteFeedbackMutation();

  const [dialog, setDialog] = useState('');
  const [feedbackTarget, setFeedbackTarget] = useState('site');
  const [message, setMessage] = useState('');

  const handleSubmit = useCallback(function() {
    if (message) {
      const newFeedback = { feedback: { feedbackMessage: message } }
      if ('site' === feedbackTarget) {
        void postSiteFeedback({ postSiteFeedbackRequest: newFeedback });
      } else {
        void postGroupFeedback({ postGroupFeedbackRequest: newFeedback })
      }
      setMessage('');
      setDialog('');
      setSnack({ snackType: 'success', snackOn: 'Thanks for your feedback!' });
    }
  }, [message])

  return <>

    <Tooltip title="Feedback">
      <IconButton
        {...targets(`topbar feedback toggle`, `submit group or site feedback`)}
        disableRipple
        color="primary"
        onClick={() => setDialog('feedback')}
      >
        <CampaignIcon sx={classes.mdHide} />
        <Typography sx={classes.mdShow}>Feedback</Typography>
      </IconButton>
    </Tooltip>

    <Dialog
      open={dialog == 'feedback'}
      onClose={() => setDialog('')}
    >
      <Grid p={2} spacing={2} container direction="row">
        <Grid size={12}>
          <TextField
            {...targets(`feedback menu target selection`, `Group or Site`, `select if feedback should go to the group or website admins`)}
            select
            fullWidth
            value={feedbackTarget}
            variant="standard"
            onChange={e => setFeedbackTarget(e.target.value)}
          >
            <MenuItem key={`site-select-give-feedback`} value={'site'}>Site</MenuItem>
            <MenuItem key={`group-select-give-feedback`} value={'group'}>Group</MenuItem>
          </TextField>
        </Grid>

        <Grid size={12}>
          <TextField
            fullWidth
            multiline
            rows={4}
            slotProps={{
              input: {
                sx: {
                  maxLength: 300
                }
              }
            }}
            helperText={`${300 - message.length}/300`}
            value={message}
            onChange={e => setMessage(e.target.value)}
            onKeyDown={e => e.stopPropagation()}
          />
        </Grid>
        <Grid size={12}>
          <Button
            {...targets(`feedback comment submit`, `submit the feedback comment`)}
            fullWidth
            onClick={handleSubmit}
          >Submit Comment</Button>
        </Grid>
      </Grid>

    </Dialog>
  </>;
}

export default FeedbackMenu;
