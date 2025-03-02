import { useMemo } from 'react';
import { validSlice, IValidActionPayload } from './valid';
import { useAppDispatch } from './store';

export function useValid(): typeof validSlice.actions {
  const dispatch = useAppDispatch();
  return useMemo(() => new Proxy(validSlice.actions, {
    get: function(target, prop: keyof typeof validSlice.actions) {
      // Forward the arguments passed to the action creators
      return (args: IValidActionPayload) => dispatch(target[prop](args));
    }
  }), []);
}
