import React, { useCallback, useState } from 'react';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';

import { siteApi, targets } from 'awayto/hooks';

interface FeedbackMenuProps extends IComponent {
  feedbackAnchorEl: null | HTMLElement;
  feedbackMenuId: string;
  isFeedbackOpen: boolean;
  handleMenuClose: () => void;
}

export function FeedbackMenu({ handleMenuClose, feedbackAnchorEl, feedbackMenuId, isFeedbackOpen }: FeedbackMenuProps): React.JSX.Element {

  const [postGroupFeedback] = siteApi.useGroupFeedbackServicePostGroupFeedbackMutation();
  const [postSiteFeedback] = siteApi.useFeedbackServicePostSiteFeedbackMutation();

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
      handleMenuClose && handleMenuClose();
    }
  }, [message])

  return <Menu
    anchorEl={feedbackAnchorEl}
    anchorOrigin={{
      vertical: 'bottom',
      horizontal: 'right',
    }}
    id={feedbackMenuId}
    keepMounted
    transformOrigin={{
      vertical: 'top',
      horizontal: 'right',
    }}
    open={!!isFeedbackOpen}
    onClose={handleMenuClose}
  >
    <Box p={1} sx={{ width: 300 }}>
      <Grid spacing={2} container direction="row">
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
            autoFocus
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
    </Box>
  </Menu>
}

export default FeedbackMenu;
