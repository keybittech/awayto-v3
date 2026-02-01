import React from 'react';

import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import FormLabel from '@mui/material/FormLabel';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Checkbox from '@mui/material/Checkbox';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import { targets, toSnakeCase, type IField } from 'awayto/hooks';

interface FieldProps extends IComponent {
  field: IField;
  error?: boolean;
  disabled?: boolean;
  onChange: (e: any) => void;
  settingsBtn?: React.JSX.Element;
}

export function Field({ settingsBtn, field, error, disabled, onChange }: FieldProps): React.JSX.Element {
  const val = field.v;
  let comp = <></>;

  if ('labelntext' === field.t) {
    comp = <Card variant="outlined" sx={{ flex: 1, p: '6px' }}>
      <CardHeader title={field.l} variant="h6" />
      <CardContent>{field.x}</CardContent>
    </Card>
  } else if ('multi-select' === field.t) {
    const selectedValues = Array.isArray(val) ? val : [];
    comp = <FormControl error={error} component="fieldset" fullWidth>
      <FormLabel component="legend">{field.l}</FormLabel>
      <FormGroup>
        {field.o?.map(opt => <FormControlLabel key={opt.i || opt.v} label={opt.l} value={opt.v} control={
          <Checkbox disabled={disabled} checked={selectedValues.includes(opt.v)} value={opt.v} onChange={onChange} />
        } />)}
      </FormGroup>
      {field.h && <FormHelperText>{field.h}</FormHelperText>}
    </FormControl>
  } else if ('single-select' === field.t) {
    comp = <FormControl error={error} component="fieldset" fullWidth>
      <FormLabel component="legend">{field.l}</FormLabel>
      <RadioGroup
        aria-label={toSnakeCase(field.l)}
        name={field.i}
        value={val ?? ''}
        onChange={onChange}
      >
        {field.o?.map(opt => <FormControlLabel key={opt.i || opt.v} label={opt.l} value={opt.v} control={
          <Radio disabled={disabled} />
        } />)}
      </RadioGroup>
      {field.h && <FormHelperText>{field.h}</FormHelperText>}
    </FormControl>
  } else if ('boolean' === field.t) {
    comp = <FormControl error={error} component="fieldset">
      <FormGroup>
        <FormControlLabel label={field.l + (field.r ? ' *' : '')} control={
          <Checkbox disabled={disabled} checked={!!val} onChange={e => onChange({ target: { value: e.target.checked } })} />
        } />
      </FormGroup>
      {field.h && <FormHelperText>{field.h}</FormHelperText>}
    </FormControl>
  } else {
    const isNumber = field.t === 'number';

    const handleTextChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      if (isNumber) {
        const numericValue = e.target.value.replace(/[^0-9.]/g, '');
        onChange({ ...e, target: { ...e.target, value: numericValue } });
      } else {
        onChange(e);
      }
    };

    comp = <TextField
      fullWidth
      {...targets(`form field ${field.l}`, field.l)}
      error={error}
      disabled={disabled}
      type={isNumber ? 'text' : field.t}
      helperText={`${field.r ? 'Required. ' : ''}${field.h || ''}`}
      value={val ?? ''}
      required={field.r}
      onChange={handleTextChange}
      slotProps={{
        htmlInput: {
          inputMode: isNumber ? 'decimal' : undefined
        },
        inputLabel: {
          shrink: true
        }
      }}
    />
  }

  return <Grid container>
    <Grid>{settingsBtn}</Grid>
    <Grid size="grow">{comp}</Grid>
  </Grid>;
}

export default Field;
