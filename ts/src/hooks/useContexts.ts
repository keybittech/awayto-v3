import { Context, createContext, useMemo, useState } from 'react';

import buildOutput from '../build.json';

const { views } = buildOutput as Record<string, Record<string, string>>;

export type IBaseContexts = Record<string, Context<unknown>>;
const contexts = {} as IBaseContexts;

export function useContexts(): IBaseContexts {
  // @ts-ignore
  // const nullContext = createContext<unknown>(new Proxy({}, (t: {}, p: string) => null));
  const nullContext = createContext<unknown>(null);
  const [loading, setLoading] = useState(false);

  return useMemo(() => new Proxy(contexts, {
    get: function(target: IBaseContexts, prop: string): boolean | Context<unknown> {
      if (!target[prop]) {
        import(`../${views[prop]}`).then((mod: { default: Context<unknown> }) => {
          target[prop] = mod.default;
          setLoading(l => !l);
        }).catch(console.error);
      }
      return Reflect.get(target, prop);
    },
  }), [loading]);
}
