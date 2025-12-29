import React, { Suspense, useState } from 'react';

import Button from '@mui/material/Button';
import InputAdornment from '@mui/material/InputAdornment';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';

import { siteApi, targets } from 'awayto/hooks';
import ManageFormModal from './ManageFormModal';
import Dialog from '@mui/material/Dialog';


interface FormPickerProps extends IComponent {
  label: string;
  helperText: string;
  formIds?: string[];
  onSelectForm: (formIds?: string[]) => void;
}

export function FormPicker({ formIds, label, helperText, onSelectForm, ...props }: FormPickerProps): React.JSX.Element {
  const { data: groupFormsRequest, refetch: getGroupForms, isSuccess: groupFormsLoaded } = siteApi.useGroupFormServiceGetGroupFormsQuery();
  const [dialog, setDialog] = useState('');
  const [value, setValue] = useState<string[]>(formIds || []);

  return <>
    {groupFormsLoaded && onSelectForm && <TextField
      {...targets(`form pick select`, label, `select a form to use`)}
      select
      fullWidth
      value={value}
      helperText={helperText}
      onChange={e => {
        const formValue = e.target.value as unknown as string[];
        if (formValue.filter(f => f === "").length) {
          setValue([]);
          onSelectForm(undefined);
          return;
        }
        setValue(formValue);
        onSelectForm(formValue.length ? formValue : undefined);
      }}
      slotProps={{
        select: {
          multiple: true
        },
        input: {
          endAdornment: <InputAdornment position="end" sx={{ mr: 2 }}>
            <Tooltip title="Create Form">
              <Button
                {...targets(`form pick new`, `create a new form`)}
                color="secondary"
                onClick={() => setDialog('create_form')}
              >
                <Typography variant="button">New</Typography>
              </Button>
            </Tooltip>
          </InputAdornment>
        }
      }}
    >
      {formIds?.length && <MenuItem key="unset-selection" value=""><Typography variant="caption">Remove selection</Typography></MenuItem>}
      {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>)}
      {!groupFormsRequest.groupForms && <MenuItem key={`no-forms`} value="" inert>No forms created</MenuItem>}
    </TextField>}

    <Dialog onClose={setDialog} open={dialog === 'create_form'} fullWidth maxWidth="lg">
      <Suspense>
        <ManageFormModal {...props} closeModal={() => {
          setDialog('')
          void getGroupForms();
        }} />
      </Suspense>
    </Dialog>
  </>;
}

export default FormPicker;
