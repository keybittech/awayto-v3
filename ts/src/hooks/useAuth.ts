import { useMemo } from 'react';
import { useAppDispatch } from './store';
import { IAuth, authSlice } from './auth';

export function useAuth(): typeof authSlice.actions {
  const dispatch = useAppDispatch();
  return useMemo(() => new Proxy(authSlice.actions, {
    get: function(target, prop: keyof typeof authSlice.actions) {
      // Forward the arguments passed to the action creators
      return (args: IAuth) => dispatch(target[prop](args));
    }
  }), []);
}
