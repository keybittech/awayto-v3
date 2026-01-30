import { useMemo, useState, useEffect, useCallback, useRef, SetStateAction } from 'react';

import { siteApi } from './api';
import { IForm } from './form';
import { deepClone } from './util';

type UseGroupFormResponse = {
  forms?: IForm[];
  setForm: (index: number, valueOrFn: SetStateAction<IForm | undefined>) => void;
  valid: boolean;
  reset: () => void;
};

export function useGroupForms(ids: string[] = []): UseGroupFormResponse {

  const [forms, setForms] = useState<IForm[]>([]);
  const original = useRef<IForm[]>([]);

  // TODO use the active endpoint
  const [getGroupForm] = siteApi.useLazyGroupFormServiceGetGroupFormByIdQuery();

  const reset = useCallback(() => {
    setForms(ids.length ? deepClone(original.current) : []);
  }, [ids]);

  const setForm = useCallback((index: number, valueOrFn: IForm | undefined | ((prev: IForm | undefined) => IForm | undefined)) => {
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
        const reqForms = res.map(r => r?.groupForm.form as unknown as IForm).filter(Boolean);
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
      const submission = form.version.submission || {};

      for (const rowId of Object.keys(form.version.form || {})) {
        for (const field of form.version.form[rowId]) {
          if (field.r) {
            const val = submission[field.i];

            if (undefined === val || null === val) {
              return false;
            }

            if ('string' === typeof val && val.trim() === '') {
              return false;
            }

            if (Array.isArray(val) && val.length === 0) {
              return false;
            }

            if ('boolean' === field.t && val !== true) {
              return false;
            }
          }
        }
      }

      return true;
    });
  }, [forms]);

  return {
    forms,
    setForm,
    valid,
    reset
  }
}
