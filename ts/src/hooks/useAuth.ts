import { useMemo } from 'react';
import { useAppDispatch } from './store';
import { IAuth, newAuthSlice } from './auth';

export function useAuth(): typeof newAuthSlice.actions {
  const dispatch = useAppDispatch();
  return useMemo(() => new Proxy(newAuthSlice.actions, {
    get: function(target, prop: keyof typeof newAuthSlice.actions) {
      // Forward the arguments passed to the action creators
      return (args: IAuth) => dispatch(target[prop](args));
    }
  }), []);
}
