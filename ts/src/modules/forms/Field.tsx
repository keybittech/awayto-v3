import React, { useMemo } from 'react';

import TextField, { TextFieldProps } from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import type { IField } from 'awayto/hooks';

interface FieldProps extends IComponent {
  field: IField;
  editable: boolean;
  defaultDisplay?: boolean;
  settingsBtn?: React.JSX.Element;
}

export function Field({ settingsBtn, defaultDisplay, field = { l: 'Label', x: 'Text' }, editable = false, ...props }: FieldProps & TextFieldProps): React.JSX.Element {
  if (!field) return <></>;

  const FieldElement: (props: TextFieldProps) => React.JSX.Element = useMemo(() => {
    switch (field.t) {
      case 'date':
      case 'time':
      case 'text':
        return TextField;
      case 'labelntext':
        return () => <Card sx={{ flex: 1, p: '6px' }}>
          <CardHeader title={field.l} variant="h6" action={settingsBtn} />
          <CardContent>{field.x}</CardContent>
        </Card>
      default:
        return () => <></>;
    }
  }, [field, settingsBtn]);

  return <FieldElement
    fullWidth
    disabled={!editable}
    id={`form field ${field.l}`}
    label={field.l}
    aria-label={`modify the form field ${field.l} which is of type ${field.t}`}
    type={field.t}
    helperText={`${field.r ? 'Required. ' : ''}${field.h || ''}`}
    value={field.v ?? ''}
    slotProps={{
      input: {
        endAdornment: settingsBtn
      },
      inputLabel: {
        shrink: true
      }
    }}
    {...props}
  />;
}

export default Field;
