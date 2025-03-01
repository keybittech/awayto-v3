import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Grid from '@mui/material/Grid';
import TextField, { TextFieldProps } from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import { siteApi } from './api';
import { IForm, IFormSubmission, IField } from './form';
import { deepClone, targets } from './util';

interface FieldProps extends IComponent {
  field: IField;
  editable: boolean;
  defaultDisplay?: boolean;
  settingsBtn?: React.JSX.Element;
}

export function Field({ settingsBtn, defaultDisplay, field, editable = false, ...props }: FieldProps & TextFieldProps): React.JSX.Element {
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

interface FormDisplayProps extends IComponent {
  form: Required<IForm>;
  setForm(value: IForm): void;
}

export function FormDisplay({ form, setForm }: FormDisplayProps): React.JSX.Element {

  useEffect(() => {
    if (setForm && form && !form?.version?.submission) {
      const submission = Object.keys(form?.version?.form || {}).reduce((m, rowId) => {
        return {
          ...m,
          [rowId]: form?.version?.form[rowId].map(r => r.v) || []
        }
      }, {}) as IFormSubmission;
      setForm({
        ...form,
        version: {
          ...form.version,
          submission
        }
      });
    }
  }, [form, setForm]);

  const rowKeys = useMemo(() => Object.keys(form?.version?.form || {}), [form]);

  const setCellAttr = useCallback((row: string, col: number, value: string, attr: string) => {
    if (form && setForm) {
      const updatedForm = { ...form };
      updatedForm.version.form[row][col][attr] = value;
      updatedForm.version.submission[row][col] = value;
      setForm(updatedForm);
    }
  }, [form, setForm]);

  return <Grid container spacing={2}>

    {rowKeys.map((rowId, i) => <Grid key={`form_fields_row_${i}`} size={12}>
      <Grid container spacing={2}>
        {form?.version.form[rowId].map((cell, j) => {
          return <Grid key={`form_fields_cell_${i + 1}_${j}`} size={12 / form.version.form[rowId].length}>
            <Field
              field={cell}
              fullWidth
              editable={true}
              helperText={`${cell.r ? 'Required. aaa' : ''}${cell.h || ''}`}
              onBlur={(e: React.FocusEvent<HTMLInputElement>) => { setCellAttr(rowId, j, e.target.value, 'v') }}
              defaultValue={cell.v || ''}
            />
          </Grid>
        })}
      </Grid>
    </Grid>)}

  </Grid>
}

type UseGroupFormResponse = {
  form?: IForm,
  comp: () => React.JSX.Element,
  valid: boolean
};

export function useGroupForm(id = ''): UseGroupFormResponse {

  const [form, setForm] = useState<IForm | undefined>();
  const [forms, setForms] = useState<Map<string, IForm>>(new Map());

  const { data: formTemplateRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId: id }, { skip: !id || forms.has(id) });

  useEffect(() => {
    if (formTemplateRequest?.groupForm.form) {
      setForms(new Map([...forms, [id, deepClone(formTemplateRequest?.groupForm.form as IForm)]]));
    }
  }, [formTemplateRequest?.groupForm]);

  useEffect(() => {
    if (!id || !forms.has(id)) {
      setForm(undefined);
    } else if (id && !form) {
      setForm(forms.get(id));
    }
  }, [id, forms, form]);

  const valid = useMemo(() => {
    let v = true;
    if (form) {
      for (const rowId of Object.keys(form.version?.form || {})) {
        form.version.form[rowId].forEach((field, i, arr) => {
          if (field.r && form.version.submission && [undefined, ''].includes(form.version.submission[rowId][i])) {
            v = false;
            arr.length = i + 1;
          }
        })
      }
    }
    return v;
  }, [form]);

  return {
    form,
    comp: !form ? (() => <></>) : () => <FormDisplay form={form} setForm={setForm} />,
    valid
  }
}
