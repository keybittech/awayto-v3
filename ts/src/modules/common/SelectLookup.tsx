import React, { useCallback, useEffect, useState } from 'react';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import MenuItem from '@mui/material/MenuItem';
import CircularProgress from '@mui/material/CircularProgress';
import InputAdornment from '@mui/material/InputAdornment';
import Grid from '@mui/material/Grid';
import ClearIcon from '@mui/icons-material/Clear';
import CheckIcon from '@mui/icons-material/Check';
import { useUtil, ILookup, isStringArray } from 'awayto/hooks';

type ActionBody = Record<string, string>;

// type MutatorFn<T, R> = (p: T) => QueryActionCreatorResult<SiteQuery<T, R>>

interface SelectLookupProps extends IComponent {
  multiple?: boolean;
  required?: boolean;
  disabled?: boolean;
  noEmptyValue?: boolean;
  lookups?: (ILookup | undefined)[];
  lookupName: string;
  helperText: React.JSX.Element | string;
  lookupChange(value: string | string[]): void;
  lookupValue: string | string[];
  defaultValue?: string | string[];
  invalidValues?: string[];
  refetchAction?: (props?: { [prop: string]: string }) => void;
  attachAction?: (p: ActionBody) => Promise<void>;
  createAction?: (p: ActionBody) => Promise<{ id: string }>;
  deleteAction?: (p: ActionBody) => Promise<void>;
  deleteComplete?(value: string | string[]): void;
  parentUuidName?: string;
  parentUuid?: string;
  attachName?: string;
  deleteActionIdentifier?: string
}

export function SelectLookup({ lookupChange, required = false, disabled = false, invalidValues = [], attachAction, attachName, refetchAction, parentUuidName, parentUuid, lookups, lookupName, helperText, lookupValue, multiple = false, noEmptyValue = false, createAction, deleteAction, deleteComplete, deleteActionIdentifier }: SelectLookupProps): React.JSX.Element {

  const { setSnack } = useUtil();

  const [addingNew, setAddingNew] = useState<boolean | undefined>();
  const [newLookup, setNewLookup] = useState({ name: '' } as ILookup);
  const [lookupUpdater, setLookupUpdater] = useState(null as unknown as string);

  if (!lookupName || !lookupChange) return <Grid container justifyContent="center"><CircularProgress /></Grid>;

  const refresh = () => {
    setAddingNew(false);
    setNewLookup({ name: '' } as ILookup);
    if (refetchAction) {
      refetchAction(parentUuidName && parentUuid ? { [parentUuidName]: parentUuid } : undefined);
    }
  }

  const handleSubmit = useCallback(() => {
    if (newLookup.name) {
      if (invalidValues.includes(newLookup.name)) {
        setSnack({ snackType: 'warning', snackOn: 'That value cannot be used here.' })
        return;
      }

      setLookupUpdater(newLookup.name);
      if (createAction) {
        let actionBody: ActionBody = { name: newLookup.name };
        createAction(actionBody).then(res => {
          const { id: lookupId } = res;
          if (attachAction && lookupId && attachName) {
            const attachPayload = { [attachName]: lookupId };
            if (parentUuid && parentUuidName) {
              attachPayload[parentUuidName] = parentUuid;
            }
            attachAction(attachPayload).then(() => refresh()).catch(console.error);
          } else {
            refresh();
          }
        }).catch(console.error);
      }
    } else {
      console.error("no lookup name")
    }
  }, [newLookup, createAction, attachAction, attachName, parentUuid, parentUuidName]);

  useEffect(() => {
    if (lookupValue && lookups?.length && isStringArray(lookupValue) && lookupUpdater) {
      const updater = lookups.find(l => l?.name === lookupUpdater);
      if (updater?.id) {
        lookupChange([...lookupValue, updater.id]);
        setLookupUpdater('');
      }
    }
  }, [lookups, lookupValue, lookupUpdater]);

  useEffect(() => {
    if (lookupValue && lookups && lookups?.length && noEmptyValue && !lookupValue?.length) {
      const firstLookup = lookups.at(0) as Required<ILookup>;
      lookupChange(isStringArray(lookupValue) ? [firstLookup.id] : firstLookup.id);
    }
  }, [lookups, lookupValue, noEmptyValue]);

  // useEffect(() => {
  //   if (defaultValue && !lookupValue) {
  //     lookupChange(defaultValue);
  //   }
  // }, [defaultValue, lookupValue]);

  return <>
    {
      addingNew ?
        <TextField
          autoFocus
          fullWidth
          label={`New ${lookupName}`}
          value={newLookup.name}
          onChange={e => {
            setNewLookup({ name: e.target.value } as ILookup)
          }}
          onKeyDown={(e) => {
            ('Enter' === e.key && newLookup.name) && handleSubmit();
          }}
          slotProps={{
            input: {
              endAdornment: <InputAdornment position="end">
                <IconButton aria-label="close new record" onClick={() => {
                  setAddingNew(false);
                  setNewLookup({ ...newLookup, name: '' });
                }}>
                  <ClearIcon style={{ color: 'red' }} />
                </IconButton>
                <IconButton aria-label="create new record" onClick={() => {
                  newLookup.name ? handleSubmit() : void setSnack({ snackOn: 'Provide a name for the record.', snackType: 'info' });
                }}>
                  <CheckIcon style={{ color: 'green' }} />
                </IconButton>
              </InputAdornment>
            }
          }}
        /> : <TextField
          select
          fullWidth
          required={required}
          disabled={disabled}
          autoFocus={!!lookupUpdater}
          id={`${lookupName}-lookup-selection`}
          helperText={helperText || ''}
          label={`${lookupName}s`}
          onChange={e => {
            const { value } = e.target as { value: string | string[] };
            if (isStringArray(value) && value.indexOf('new') > -1) {
            } else {
              lookupChange(value);
            }
          }}
          value={lookupValue}
          slotProps={{
            select: {
              multiple,
              renderValue: selected => {
                const lookupText = isStringArray(selected as string[]) ?
                  (selected as string[]).map(v => lookups?.find(r => r?.id === v)?.name).join(', ') :
                  lookups?.find(r => r?.id === selected)?.name as string;

                return <Box sx={{ whiteSpace: 'pre-wrap' }}>{lookupText}</Box>;
              }
            }
          }}
        >
          {!multiple && !noEmptyValue && <MenuItem value="">No selection</MenuItem>}
          {createAction && <MenuItem value="new" onClick={e => {
            e.preventDefault();
            setAddingNew(true);
          }}>Add a {lookupName} to this list</MenuItem>}
          {!lookups?.length && <MenuItem disabled>
            You have no pre-defined {lookupName}s.<br />
            Click the button above to add some.
          </MenuItem>}
          {lookups?.length ? lookups.map((lookup, i) => (
            <MenuItem key={i} style={{ display: 'flex' }} value={lookup?.id}>
              <span style={{ flex: '1' }}>{lookup?.name}</span>
              {refetchAction && deleteAction && <ClearIcon style={{ color: 'red', marginRight: '8px' }} onClick={e => {
                e.stopPropagation();
                if (isStringArray(lookupValue) && lookup?.id) {
                  lookupChange([...lookupValue.filter(l => l !== lookup?.id)]);
                } else if (lookupValue === lookup?.id) {
                  lookupChange('');
                }

                if (lookup?.id && ((parentUuidName && parentUuid && attachName) || deleteActionIdentifier)) {
                  deleteAction(
                    parentUuidName && parentUuid && attachName ? { [parentUuidName]: parentUuid, [attachName]: lookup?.id } :
                      deleteActionIdentifier ? { [deleteActionIdentifier]: lookup?.id } :
                        {}
                  ).then(() => {
                    if (lookup?.id) {
                      deleteComplete && deleteComplete(lookup?.id);
                      refresh()
                    }
                  }).catch(console.error);
                }

              }} />}
            </MenuItem>
          )) : []}
        </TextField>
    }
  </>;
}

export default SelectLookup;
