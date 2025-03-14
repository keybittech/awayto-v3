import React from 'react';

import TextField, { TextFieldProps } from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import type { IField } from 'awayto/hooks';

interface FieldProps extends IComponent {
  field: IField;
  settingsBtn?: React.JSX.Element;
}

export function Field({ settingsBtn, field = { l: 'Label', x: 'Text' }, error, disabled, onChange }: FieldProps & TextFieldProps): React.JSX.Element {
  if (!field) return <></>;

  if ('labelntext' == field.t) {
    return <Card variant="outlined" sx={{ flex: 1, p: '6px' }}>
      <CardHeader title={field.l} variant="h6" action={settingsBtn} />
      <CardContent>{field.x}</CardContent>
    </Card>
  }

  return <TextField
    fullWidth
    error={error}
    disabled={disabled}
    id={`form field ${field.l}`}
    label={field.l}
    aria-label={`modify the form field ${field.l} which is of type ${field.t}`}
    type={field.t}
    helperText={`${field.r ? 'Required. ' : ''}${field.h || ''}`}
    value={field.v ?? ''}
    required={field.r}
    onChange={onChange}
    slotProps={{
      input: {
        endAdornment: settingsBtn
      },
      inputLabel: {
        shrink: true
      }
    }}
  />;
}

export default Field;
