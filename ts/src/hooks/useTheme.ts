import { useMemo } from 'react';
import { useAppDispatch } from './store';
import { ITheme, themeSlice } from './theme';

export function useTheme(): typeof themeSlice.actions {
  const dispatch = useAppDispatch();
  return useMemo(() => new Proxy(themeSlice.actions, {
    get: function(target, prop: keyof typeof themeSlice.actions) {
      // Forward the arguments passed to the action creators
      return (...args: [ITheme]) => dispatch(target[prop](...args));
    }
  }), []);
}
