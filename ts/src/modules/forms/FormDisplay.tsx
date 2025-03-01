import React, { useMemo, useEffect, useCallback } from 'react';

import Grid from '@mui/material/Grid';

import type { IForm, IFormSubmission } from 'awayto/hooks';

import Field from './Field';

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
