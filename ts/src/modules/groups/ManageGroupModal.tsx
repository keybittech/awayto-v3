import React, { useState, useCallback, useEffect } from 'react';

import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';

import { useDebounce, useUtil, refreshToken, siteApi, IGroup } from 'awayto/hooks';

declare global {
  interface IComponent {
    showCancel?: boolean;
    editGroup?: IGroup;
    setEditGroup?: React.Dispatch<React.SetStateAction<IGroup>>;
    saveToggle?: number;
  }
}

export function ManageGroupModal({ children, editGroup, setEditGroup, showCancel = true, saveToggle = 0, closeModal }: IComponent): React.JSX.Element {

  const { setSnack } = useUtil();

  const [group, setGroup] = useState({
    name: '',
    displayName: '',
    purpose: '',
    allowedDomains: '',
    ai: false,
    ...editGroup
  } as Required<IGroup>);

  const debouncedName = useDebounce(group.name || '', 1000);
  const [debouncedPrev, setDebouncedPrev] = useState(debouncedName);

  const [checkName, checkState] = siteApi.useLazyGroupServiceCheckGroupNameQuery();

  const [editedPurpose, setEditedPurpose] = useState(false);
  const [allowedDomains, setAllowedDomains] = useState([] as string[]);
  const [allowedDomain, setAllowedDomain] = useState('');

  const [getUserProfileDetails] = siteApi.useLazyUserProfileServiceGetUserProfileDetailsQuery();
  const [postGroup] = siteApi.useGroupServicePostGroupMutation();
  const [patchGroup] = siteApi.useGroupServicePatchGroupMutation();

  const handleSubmit = useCallback(async () => {
    if (!group.name || !group.purpose) {
      setSnack({ snackType: 'error', snackOn: 'All fields are required.' });
      return;
    }

    group.allowedDomains = allowedDomains.join(',');

    const newGroup = {
      displayName: group.name,
      name: group.name.replace(/__+/g, '_')
        .replace(/\s/g, '_')
        .replace(/[\W]+/g, '_')
        .replace(/__+/g, '_')
        .replace(/__+/g, '').toLowerCase(),
      purpose: group.purpose,
      allowedDomains: allowedDomains.join(','),
      ai: group.ai
    };

    if (group.id) {
      await patchGroup({ patchGroupRequest: newGroup }).unwrap().catch(console.error);
    } else {
      await postGroup({ postGroupRequest: newGroup }).unwrap().then(resp => {
        group.id = resp.id;
      }).catch(console.error);
    }

    await refreshToken(61).then(async () => {
      await getUserProfileDetails();
      closeModal && closeModal(group);
    }).catch(console.error);
  }, [group, editGroup]);

  const badName = group.name != debouncedName || checkState.isFetching || checkState.isError;

  useEffect(() => {
    async function go() {
      if (debouncedName != debouncedPrev) {
        const { data: check } = await checkName({ name: debouncedName });
        if (check?.isValid) {
          setDebouncedPrev(debouncedName);
        }
        setGroup({ ...group, isValid: !!check?.isValid });
      }
    }
    void go();
  }, [debouncedName, debouncedPrev]);

  // Onboarding handling
  useEffect(() => {
    if (setEditGroup) {
      setEditGroup({ ...group, isValid: !badName });
    }
  }, [group, badName]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  return <>
    <Card>
      <CardHeader title={`${editGroup ? 'Edit' : 'Create'} Group`}></CardHeader>
      <CardContent>
        {!!children && children}

        <Grid container spacing={4}>
          <Grid size={12}>
            <TextField
              fullWidth
              id="name"
              label="Group Name"
              value={group.name}
              error={badName}
              name="name"
              onChange={e => { setGroup({ ...group, name: e.target.value }); }}
              multiline
              required
              helperText="Group names can only contain letters, numbers, and underscores. Max 50 characters."
            />
          </Grid>

          <Grid size={12}>
            <TextField
              id={`group-purpose-entry`}
              fullWidth
              inputProps={{ maxLength: 100 }}
              helperText={'Enter a short phrase about the function of your group (max. 100 characters).'}
              label={`Group Description`}
              required
              error={editedPurpose && !!group.purpose && group.purpose.length > 100}
              onBlur={() => setEditedPurpose(true)}
              onFocus={() => setEditedPurpose(false)}
              onChange={e => setGroup({ ...group, purpose: e.target.value })}
              value={group.purpose}
            />
          </Grid>

          <Grid size={12}>
            <TextField
              id={`group-allowed-domains-entry`}
              fullWidth
              helperText={`These email domains will be allowed to join the group. Leaving this empty means anyone can join.`}
              label={`Allowed Email Domains`}
              onChange={e => setAllowedDomain(e.target.value)}
              value={allowedDomain}
              InputProps={{
                endAdornment: <Button
                  variant="text"
                  color="secondary"
                  onClick={() => {
                    if (!/[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*/.test(allowedDomain)) {
                      setSnack({ snackType: 'info', snackOn: 'Must be an email domain, like DOMAIN.COM' })
                    } else {
                      setAllowedDomains([...allowedDomains, allowedDomain])
                      setAllowedDomain('');
                    }
                  }}
                >Add</Button>
              }}
            />
            <Grid container>
              {allowedDomains.map((ad, i) => <Box key={`allowed-domain-selection-${i}`} mt={2} mr={2}>
                <Chip
                  label={ad}
                  color="secondary"
                  onDelete={() => {
                    setAllowedDomains(allowedDomains.filter(da => da !== ad))
                  }}
                />
              </Box>)}
            </Grid>
          </Grid>

          <Grid size={12}>
            <FormGroup>
              <FormControlLabel
                id={`group-disable-ai-entry`}
                control={
                  <Checkbox
                    checked={group.ai}
                    onChange={() => setGroup({ ...group, ai: !group.ai })}
                  />
                }
                label="Use AI Suggestions"
              />
              <Typography variant="caption">AI suggestions will be seen by all group members. This functionality can be toggled on/off in group settings. Group name and description are used to generate suggestions.</Typography>
            </FormGroup>
          </Grid>
        </Grid>
      </CardContent>
      {!setEditGroup && <CardActions>
        <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
          {showCancel && <Button onClick={closeModal}>Cancel</Button>}
          <Button color="info" size="large" disabled={!editGroup?.id && (group.purpose.length > 100 || badName)} onClick={handleSubmit}>
            Save Group
          </Button>
        </Grid>
      </CardActions>}
    </Card>
  </>
}

export default ManageGroupModal;
