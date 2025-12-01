import { MutationDefinition, QueryDefinition } from '@reduxjs/toolkit/query';
import { FetchArgs, BaseQueryFn, fetchBaseQuery, FetchBaseQueryError, TypedUseQueryHookResult, retry, reactHooksModule, buildCreateApi, coreModule } from '@reduxjs/toolkit/query/react'

import { utilSlice } from './util';
import { RootState } from './store';

const {
  VITE_REACT_APP_APP_HOST_URL,
} = import.meta.env;

const setSnack = utilSlice.actions.setSnack;

const baseQuery = fetchBaseQuery({
  timeout: 30000,
  cache: 'no-cache',
  mode: 'same-origin',
  credentials: 'include',
  baseUrl: VITE_REACT_APP_APP_HOST_URL + "/api",
  responseHandler: (response) => {
    const contentType = response.headers.get('Content-Type');
    if (contentType === 'application/x-awayto-vault') {
      return response.text();
    }
    return response.json();
  },
});

type CustomExtraOptions = {
  vaultSecret?: string;
}

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError, CustomExtraOptions> = async (args, api, extraOptions = {}) => {
  if (!args.headers) {
    args.headers = new Headers();
  }

  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;

  (args.headers as Headers).set('X-Tz', tz);

  const state = api.getState() as RootState;
  const vaultKey = state.auth.vaultKey;

  if (vaultKey && "function" == typeof window.pqcEncrypt) {

    const isMutation = ['POST', 'PUT', 'PATCH'].includes(args.method || '');

    if (isMutation) {

      if (!args.body) {
        args.body = {};
      }

      const encryptedBody = window.pqcEncrypt(vaultKey, JSON.stringify(args.body));

      if (encryptedBody && encryptedBody.blob) {
        const bstring = atob(encryptedBody.blob);
        const len = bstring.length;
        const bbytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bbytes[i] = bstring.charCodeAt(i);
        }

        args.body = bbytes;
        (args.headers as Headers).set('Content-Type', 'application/x-awayto-vault');
        (args.headers as Headers).set('X-Original-Content-Type', 'application/json');

        extraOptions.vaultSecret = encryptedBody.secret;
      }
    } else {
      const encryptedBody = window.pqcEncrypt(vaultKey, " ");
      if (encryptedBody && encryptedBody.blob) {
        (args.headers as Headers).set('X-Awayto-Vault', encryptedBody.blob);
        extraOptions.vaultSecret = encryptedBody.secret;
      }
    }
  } else {
    (args.headers as Headers).set('Content-Type', 'application/json');
  }

  let result = await baseQuery(args, api, extraOptions);

  if (401 == result.error?.status) {
    window.location.href = `/auth/login?tz=${tz}`
  }

  if (extraOptions.vaultSecret && result.meta?.response?.headers.get('Content-Type') === 'application/x-awayto-vault') {
    if (typeof result.data === 'string') {
      const decryptedJson = window.pqcDecrypt(result.data.trim(), extraOptions.vaultSecret);
      if (!decryptedJson) {
        console.error("WASM pqcDecrypt returned null. Check WASM console logs.");
        api.dispatch(setSnack({ snackOn: "Vault Decryption Failed" }));
      }

      try {
        result.data = JSON.parse(decryptedJson);
      } catch (e) {
        console.error("JSON Parse error after decryption", decryptedJson);
        api.dispatch(setSnack({ snackOn: "Invalid JSON after decrypt" }));
      }
    }
  }

  switch (result.error?.status) {
    case "PARSING_ERROR":
      api.dispatch(setSnack({ snackOn: result.error.data as string }));
      break
    case "FETCH_ERROR":
      result = await retry(baseQuery)(args, api, {});
      break
    default:
  }

  return result;
}

export type CustomBaseQuery = typeof customBaseQuery;

const createApi = buildCreateApi(
  coreModule(),
  reactHooksModule({ unstable__sideEffectsInRender: false })
);

// Keep data for longer than the server 3 min cache
export const siteApiTemplate = createApi({
  keepUnusedDataFor: 200,
  refetchOnMountOrArgChange: 200,
  baseQuery: customBaseQuery,
  endpoints: () => ({}),
});

export type SiteMutation<TQueryArg, TResultType> = MutationDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;
export type SiteQuery<TQueryArg, TResultType> = QueryDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;

export type UseSiteQuery<T, R> = TypedUseQueryHookResult<R, T, CustomBaseQuery>;
