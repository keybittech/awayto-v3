import React, { useMemo, useState, useEffect } from 'react';

import { siteApi } from './api';
import { IForm } from './form';
import { deepClone } from './util';
import FormDisplay from '../modules/forms/FormDisplay';

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
