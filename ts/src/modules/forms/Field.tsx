import React, { useMemo } from 'react';

import TextField, { TextFieldProps } from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import { IField, targets } from 'awayto/hooks';

interface FieldProps extends IComponent {
  field: IField;
  editable: boolean;
  defaultDisplay?: boolean;
  settingsBtn?: React.JSX.Element;
}

function Field({ settingsBtn, defaultDisplay, field, editable = false, ...props }: FieldProps & TextFieldProps): React.JSX.Element {
  if (!field) return <></>;

  const FieldElement: (props: TextFieldProps) => React.JSX.Element = useMemo(() => {
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

  // const defaultValue = useMemo(() => {
  //   switch (field.t) {
  //     case 'date':
  //       return dayjs().format('YYYY-MM-DD');
  //     case 'time':
  //       return dayjs().format('hh:mm:ss');
  //     default:
  //       return defaultDisplay ? ' ' : '';
  //   }
  // }, [field]);

  return <FieldElement
    fullWidth
    disabled={!editable}
    {...targets(`form field ${field.l}`, field.l, `modify the form field ${field.l} which is of type ${field.t}`)}
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
