import { useMemo, useState, useEffect } from 'react';

import { siteApi } from './api';
import { IForm } from './form';
import { deepClone } from './util';

import FormDisplay from '../modules/forms/FormDisplay';

type UseGroupFormResponse = {
  form?: IForm;
  comp: React.JSX.Element;
  valid: boolean;
};

export function useGroupForm(id = '', didSubmit = false): UseGroupFormResponse {

  const [form, setForm] = useState<IForm | undefined>();

  const [getGroupForm] = siteApi.useLazyGroupFormServiceGetGroupFormByIdQuery();

  useEffect(() => {
    async function go() {
      if (id) {
        const formRequest = await getGroupForm({ formId: id }).unwrap();
        if (formRequest.groupForm.form) {
          setForm(deepClone(formRequest.groupForm.form as IForm));
        }
      } else {
        setForm(undefined);
      }
    }
    void go();
  }, [id]);

  const valid = useMemo(() => {
    if (!form || !form.version.submission) {
      return false;
    }
    for (const rowId of Object.keys(form.version.form || {})) {
      for (let i = 0; i < form.version.form[rowId].length; i++) {
        const formField = form.version.form[rowId][i];
        const submissionValue = form.version.submission[rowId][i];

        if (formField.r && [undefined, ''].includes(submissionValue)) {
          return false;
        }
      }
    }

    return true;
  }, [form]);

  return {
    form,
    comp: form ? <FormDisplay form={form} setForm={setForm} didSubmit={didSubmit} /> : <></>,
    valid
  }
}
