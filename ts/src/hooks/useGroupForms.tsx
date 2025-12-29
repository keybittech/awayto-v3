import { useMemo, useState, useEffect, useCallback, useRef } from 'react';

import { siteApi } from './api';
import { IForm } from './form';
import { deepClone } from './util';

import FormDisplay from '../modules/forms/FormDisplay';

type UseGroupFormResponse = {
  forms?: IForm[];
  comp: React.JSX.Element;
  valid: boolean;
  reset: () => void;
};

export function useGroupForms(ids: string[] = [], didSubmit = false): UseGroupFormResponse {

  const [forms, setForms] = useState<IForm[]>([]);
  const original = useRef<IForm[]>([]);

  const [getGroupForm] = siteApi.useLazyGroupFormServiceGetGroupFormByIdQuery();

  const reset = useCallback(() => {
    setForms(ids.length ? deepClone(original.current) : []);
  }, [ids]);

  const updateFormAtIndex = useCallback((index: number, valueOrFn: IForm | undefined | ((prev: IForm | undefined) => IForm | undefined)) => {
    setForms(prevForms => {
      const newForms = [...prevForms];

      const currentVal = newForms[index];

      const newValue = typeof valueOrFn === 'function' ? valueOrFn(currentVal) : valueOrFn;

      if (newValue) {
        newForms[index] = newValue;
      }
      return newForms;
    });
  }, []);

  useEffect(() => {
    if (ids.length) {
      const gets = ids.map(id => getGroupForm({ formId: id }).unwrap());
      Promise.all(gets).then(res => {
        const reqForms = res.map(r => r?.groupForm.form as IForm).filter(Boolean);
        original.current = reqForms;
        setForms(deepClone(reqForms));
      }).catch(console.error);
    } else {
      setForms([]);
    }
  }, [JSON.stringify(ids)]);

  const valid = useMemo(() => {
    if (!forms.length) return true;

    return forms.every(form => {
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
    });
  }, [forms]);

  const comp = useMemo(() => <>
    {forms.map((form, index) => (
      <FormDisplay key={form.id} form={form} setForm={val => updateFormAtIndex(index, val)} didSubmit={didSubmit} />
    ))}
  </>, [forms, didSubmit, updateFormAtIndex]);

  return {
    forms,
    comp,
    valid,
    reset
  }
}
