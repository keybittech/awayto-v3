import { useMemo, useState, useEffect, useCallback, useRef } from 'react';

import { siteApi } from './api';
import { IForm } from './form';
import { deepClone } from './util';

import FormDisplay from '../modules/forms/FormDisplay';

type UseGroupFormResponse = {
  form?: IForm;
  comp: React.JSX.Element;
  valid: boolean;
  reset: () => void;
};

export function useGroupForm(id = '', didSubmit = false): UseGroupFormResponse {

  const [form, setForm] = useState<IForm | undefined>();
  const original = useRef<IForm | undefined>(undefined);

  const [getGroupForm] = siteApi.useLazyGroupFormServiceGetGroupFormByIdQuery();

  const reset = useCallback(() => {
    setForm(id.length ? deepClone(original.current) : undefined);
  }, [id]);

  useEffect(() => {
    if (id.length) {
      getGroupForm({ formId: id }).unwrap().then(formRequest => {
        if (formRequest?.groupForm.form) {
          original.current = formRequest.groupForm.form as IForm;
          setForm(deepClone(original.current));
        }
      });
    } else {
      setForm(undefined);
    }
  }, [id]);

  const valid = useMemo(() => {
    if (!form || !form.version.submission) {
      return true;
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
    valid,
    reset
  }
}
