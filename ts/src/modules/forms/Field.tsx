import React, { useMemo } from 'react';

import TextField, { TextFieldProps } from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import { dayjs, IField } from 'awayto/hooks';

type FieldProps = {
  field?: IField;
  defaultDisplay?: boolean;
  editable?: boolean;
  settingsBtn?: React.JSX.Element;
};

declare global {
  interface IComponent extends FieldProps { }
}

function Field({ settingsBtn, defaultDisplay, field, editable = false }: IComponent): React.JSX.Element {
  if (!field) return <></>;

  const FieldElement: (props: TextFieldProps) => JSX.Element = useMemo(() => {
    switch (field.t) {
      case 'date':
      case 'time':
      case 'text':
        return TextField;
      case 'labelntext':
        return () => <Card sx={{ flex: 1, p: '6px' }}>
          <CardHeader title={field.l || 'Label'} variant="h6" action={settingsBtn} />
          <CardContent>{field.x || 'Text'}</CardContent>
        </Card>
      default:
        return () => <></>;
    }
  }, [field, settingsBtn]);

  const defaultValue = useMemo(() => {
    switch (field.t) {
      case 'date':
        return dayjs().format('YYYY-MM-DD');
      case 'time':
        return dayjs().format('hh:mm:ss');
      default:
        return defaultDisplay ? ' ' : '';
    }
  }, [field]);

  return <FieldElement
    fullWidth
    disabled={!editable}
    label={field.l}
    type={field.t}
    helperText={`${field.r ? 'Required. ' : ''}${field.h || ''}`}
    value={field.v ? field.v : defaultValue}
    InputProps={{
      endAdornment: settingsBtn
    }}
  />;
}

export default Field;
