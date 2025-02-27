import React, { useState, useCallback, useEffect, useRef } from 'react';

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
import CircularProgress from '@mui/material/CircularProgress';

import { useDebounce, useUtil, refreshToken, siteApi, IGroup } from 'awayto/hooks';

const {
  VITE_REACT_APP_AI_ENABLED
} = import.meta.env;

interface ManageGroupModalProps extends IComponent {
  editGroup: IGroup;
  onValidChanged?: (valid: boolean) => void;
  showCancel?: boolean;
  saveToggle?: number;
}

export function ManageGroupModal({ children, editGroup, onValidChanged, showCancel = true, saveToggle = 0, closeModal }: ManageGroupModalProps): React.JSX.Element {

  const { setSnack } = useUtil();

  const [group, setGroup] = useState({
    name: '',
    displayName: '',
    purpose: '',
    allowedDomains: '',
    ai: false,
    ...editGroup
  } as Required<IGroup>);

  const debouncedName = useDebounce(group.displayName || '', 500);

  const [checkName, checkState] = siteApi.useLazyGroupUtilServiceCheckGroupNameQuery();

  const [groupValid, setGroupValid] = useState(false);
  const debouncedValidity = useDebounce(groupValid, 50);

  // const [editedPurpose, setEditedPurpose] = useState(false);
  const [allowedDomains, setAllowedDomains] = useState([] as string[]);
  const [allowedDomain, setAllowedDomain] = useState('');
  const purposeEdited = useRef(false);

  const [getUserProfileDetails] = siteApi.useLazyUserProfileServiceGetUserProfileDetailsQuery();
  const [postGroup] = siteApi.useGroupServicePostGroupMutation();
  const [patchGroup] = siteApi.useGroupServicePatchGroupMutation();

  const formatName = (name: string) => name.replace(/__+/g, '_')
    .replace(/\s/g, '_')
    .replace(/[\W]+/g, '_')
    .replace(/__+/g, '_')
    .replace(/__+/g, '').toLowerCase();

  const handleSubmit = useCallback(async () => {
    if (editGroup && editGroup.name == group.name && editGroup.purpose == group.purpose && editGroup.ai == group.ai) {
      closeModal && closeModal(group);
      return;
    }

    if (!group.displayName || !group.purpose) {
      setSnack({ snackType: 'error', snackOn: 'All fields are required.' });
      return;
    }

    group.allowedDomains = allowedDomains.join(',');

    if (!group.name) {
      group.name = formatName(debouncedName);
    }

    if (group.id) {
      await patchGroup({ patchGroupRequest: { ...group } }).unwrap().catch(console.error);
    } else {
      await postGroup({ postGroupRequest: { ...group } }).unwrap().then(resp => {
        group.id = resp.id;
      }).catch(console.error);
    }

    await refreshToken(61).then(async () => {
      await getUserProfileDetails();
      closeModal && closeModal(group);
    }).catch(console.error);
  }, [group, editGroup]);

  const badName = !checkState.isUninitialized && (!group.name || checkState.isFetching || checkState.isError);

  useEffect(() => {
    async function go() {
      const update: Partial<IGroup> = {};
      if (!debouncedName.length) { // must have a name to check
        update.isValid = false;
        update.name = '';
      } else if (editGroup && editGroup.displayName == debouncedName) { // don't check instantly when editing
        update.isValid = true
      } else { // else check as normal and update name if valid
        const name = formatName(debouncedName);
        const { isValid } = await checkName({ name }).unwrap();
        update.isValid = isValid;
        update.name = isValid ? name : '';
      }

      setGroup(g => ({ ...g, ...update }));
    }
    void go();
  }, [debouncedName, editGroup]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  useEffect(() => {
    setGroupValid(Boolean(
      !(checkState.isFetching || checkState.isLoading) && group.displayName == debouncedName && group.name && group.purpose && group.isValid
    ));
  }, [group, debouncedName, checkState]);

  useEffect(() => {
    if (onValidChanged) {
      onValidChanged(debouncedValidity);
    }
  }, [debouncedValidity]);

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
              value={group.displayName}
              error={badName}
              name="name"
              onChange={e => {
                setGroup({ ...group, displayName: e.target.value });
              }}
              multiline
              required
              helperText="Group names can only contain letters, numbers, and underscores. Max 50 characters."
              slotProps={{
                input: {
                  endAdornment: <>
                    {group.displayName && (debouncedName.length && !group.isValid) && !checkState.isUninitialized && !checkState.isFetching ? <Box
                      sx={{
                        color: '#000',
                        fontSize: '12px',
                        padding: '0 8px',
                        backgroundColor: 'rgb(255, 150, 150)',
                        border: '2px solid rgb(255, 100, 100)',
                      }}
                    >
                      Unavailable
                    </Box> : checkState.isFetching ? <CircularProgress color="info" size={16} /> : <></>
                    }
                  </>
                }
              }}
            />
          </Grid>

          <Grid size={12}>
            <TextField
              id={`group-purpose-entry`}
              fullWidth
              helperText={'Enter a short phrase about the function of your group (max. 100 characters).'}
              label={`Group Description`}
              required
              error={purposeEdited.current && (group.purpose.length == 0 || group.purpose.length > 100)}
              onChange={e => {
                if (e.target.value.length > 100) return;
                purposeEdited.current = true;
                setGroup({ ...group, purpose: e.target.value })
              }}
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
              slotProps={{
                input: {
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
                }
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

          {'1' == VITE_REACT_APP_AI_ENABLED && <Grid size={12}>
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
          </Grid>}
        </Grid>
      </CardContent>
      {!onValidChanged && <CardActions>
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
