import React, { Suspense, useState } from 'react';

import Button from '@mui/material/Button';
import InputAdornment from '@mui/material/InputAdornment';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';

import { siteApi } from 'awayto/hooks';
import ManageFormModal from './ManageFormModal';
import Dialog from '@mui/material/Dialog';

declare global {
  interface IComponent {
    label?: string;
    helperText?: string;
    formId?: string;
    onSelectForm?: (formId: string) => void;
  }
}

export function FormPicker({ formId, label, helperText, onSelectForm, ...props }: IComponent) {
  const { data: groupFormsRequest, refetch: getGroupForms, isSuccess: groupFormsLoaded } = siteApi.useGroupFormServiceGetGroupFormsQuery();
  const [dialog, setDialog] = useState('');

  return <>
    {groupFormsLoaded && onSelectForm && <TextField
      select
      fullWidth
      value={formId}
      label={label}
      helperText={helperText}
      onChange={e => onSelectForm(e.target.value)}
      InputProps={{
        endAdornment: (
          <InputAdornment position="end" sx={{ mr: 2 }}>
            <Tooltip key={'create_form'} title="New">
              <Button color="secondary" onClick={() => setDialog('manage_form')}>
                <Typography variant="button">New</Typography>
              </Button>
            </Tooltip>
          </InputAdornment>
        ),
      }}
    >
      {formId && <MenuItem key="unset-selection" value=""><Typography variant="caption">Remove selection</Typography></MenuItem>}
      {groupFormsRequest?.groupForms?.map(gf => <MenuItem key={`form-version-select${gf.form?.id}`} value={gf.form?.id}>{gf.form?.name}</MenuItem>) || <MenuItem key={`no-forms`} value="">No forms created</MenuItem>}
    </TextField>}

    <Dialog open={dialog === 'manage_form'} fullWidth maxWidth="lg">
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
