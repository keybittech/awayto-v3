import React, { useMemo, useCallback } from 'react';

import Grid from '@mui/material/Grid';

import type { IForm } from 'awayto/hooks';

import Field from './Field';

interface FormDisplayProps extends IComponent {
  form?: Required<IForm>;
  setForm(value: IForm): void;
  didSubmit?: boolean;
}

export function FormDisplay({ form, setForm, didSubmit }: FormDisplayProps): React.JSX.Element {

  const rowKeys = useMemo(() => Object.keys(form?.version?.form || {}), [form]);

  const setCellAttr = useCallback((fieldId: string, value: string, fieldType: string) => {
    if (form && setForm) {
      const updatedForm = { ...form };
      if (!updatedForm.version.submission) {
        updatedForm.version.submission = {};
      }

      const currentSubmission = updatedForm.version.submission;

      if ('multi-select' === fieldType) {
        const currentOpts = (currentSubmission[fieldId] as string[]) || [];
        const valueIdx = currentOpts.indexOf(value);
        if (valueIdx > -1) {
          const newOpts = [...currentOpts];
          newOpts.splice(valueIdx, 1);
          currentSubmission[fieldId] = newOpts;
        } else {
          currentSubmission[fieldId] = [...currentOpts, value];
        }
      } else if ('boolean' === fieldType) {
        currentSubmission[fieldId] = Boolean(value);
      } else {
        currentSubmission[fieldId] = value;
      }

      setForm(updatedForm);
    }
  }, [form, setForm]);

  return <Grid container spacing={2}>
    {rowKeys.map((rowId, i) => (
      <Grid key={`form_fields_row_${i}`} size={12}>
        <Grid container spacing={2}>
          {form?.version.form[rowId].map(field => {
            field.v = form.version.submission?.[field.i] ?? '';
            return <Grid key={`form_fields_cell_${field.i}`} size={12 / form.version.form[rowId].length}>
              <Field
                field={field}
                error={didSubmit && field.r && (field.v === '' || (Array.isArray(field.v) && field.v.length == 0))}
                onChange={e => { setCellAttr(field.i, e.target.value, field.t) }}
              />
            </Grid>
          })}
        </Grid>
      </Grid>
    ))}
  </Grid>;
}

export default FormDisplay;
