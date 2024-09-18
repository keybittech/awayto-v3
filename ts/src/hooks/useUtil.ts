import { useMemo } from 'react';
import { useAppDispatch } from './store';
import { IUtil, newUtilSlice } from './util';

export function useUtil(): typeof newUtilSlice.actions {
  const dispatch = useAppDispatch();
  return useMemo(() => new Proxy(newUtilSlice.actions, {
    get: function(target, prop: keyof typeof newUtilSlice.actions) {
      // Forward the arguments passed to the action creators
      return (...args: [IUtil]) => dispatch(target[prop](...args));
    }
  }), []);
}
