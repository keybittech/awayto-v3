import React, { useCallback, useEffect, useState } from 'react';

import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import MenuItem from '@mui/material/MenuItem';
import CircularProgress from '@mui/material/CircularProgress';
import Grid from '@mui/material/Grid';

import ClearIcon from '@mui/icons-material/Clear';

import { useUtil, ILookup, isStringArray, toSnakeCase, targets } from 'awayto/hooks';

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

      let existing = lookups?.find(l => l?.name == newLookup.name);
      if (existing?.id && isStringArray(lookupValue) && lookupValue.includes(existing.id)) {
        refresh();
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

  // This triggers after handleSubmit, updating the lookupValue with the newly added lookup
  useEffect(() => {
    if (lookupValue && lookups?.length && isStringArray(lookupValue) && lookupUpdater) {
      const updater = lookups.find(l => l?.name === lookupUpdater);
      if (updater?.id) {
        lookupChange([...lookupValue, updater.id]);
        setLookupUpdater('');
      }
    }
  }, [lookups, lookupValue, lookupUpdater]);

  // If a value is required to be selected upon load, pre populate it with the first existing lookup
  useEffect(() => {
    if (lookupValue && lookups && lookups?.length && noEmptyValue && !lookupValue?.length) {
      const firstLookup = lookups.at(0) as Required<ILookup>;
      lookupChange(isStringArray(lookupValue) ? [firstLookup.id] : firstLookup.id);
    }
  }, [lookups, lookupValue, noEmptyValue]);

  return <>
    {addingNew && <Grid container spacing={2} direction="column">
      <Grid size="grow">
        <TextField
          autoFocus
          fullWidth
          {...targets(`select lookup input ${lookupName}`, `${lookupName} Name`, `${lookupName} name text field`)}
          value={newLookup.name}
          onChange={e => {
            setNewLookup({ name: e.target.value } as ILookup)
          }}
          onKeyDown={(e) => {
            ('Enter' === e.key && newLookup.name) && handleSubmit();
          }}
        />
      </Grid>

      <Grid container>
        <Button
          {...targets(`select lookup input cancel ${lookupName}`, ``, `cancel ${lookupName} creation`)}
          color="error"
          onClick={() => {
            setAddingNew(false);
            setNewLookup({ ...newLookup, name: '' });
          }}
        >
          Cancel
        </Button>
        <Button
          {...targets(`select lookup input submit ${lookupName}`, ``, `submit ${lookupName} creation`)}
          color="success"
          onClick={() => {
            newLookup.name ? handleSubmit() :
              void setSnack({ snackOn: 'Provide a name for the record.', snackType: 'info' });
          }}
        >
          Create
        </Button>
      </Grid>
    </Grid>}

    {!addingNew && <TextField
      select
      fullWidth
      required={required}
      disabled={disabled}
      autoFocus={!!lookupUpdater}
      {...targets(`select lookup selection ${lookupName}`, `${lookupName}s`, `select a ${lookupName}`)}
      value={lookupValue}
      helperText={helperText || ''}
      onChange={e => {
        const { value } = e.target as { value: string | string[] };
        if (isStringArray(value) && value.indexOf('new') > -1) {
        } else {
          lookupChange(value);
        }
      }}
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
      {createAction && <Button value="new" sx={{ mx: 1, mb: 1 }} variant="text" color="info" onClick={e => {
        e.preventDefault();
        setAddingNew(true);
      }}>create {lookupName}</Button>}
      {!lookups?.length && <MenuItem disabled>
        You have no pre-defined {lookupName}s.<br />
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
    </TextField>}
  </>;
}

export default SelectLookup;
