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
});

type CustomExtraOptions = {
  vaultSecret?: string;
}

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError, CustomExtraOptions> = async (args, api, extraOptions = {}) => {
  if (!args.headers) {
    args.headers = new Headers();
  }

  (args.headers as Headers).set('X-Tz', Intl.DateTimeFormat().resolvedOptions().timeZone);
  (args.headers as Headers).set('Content-Type', 'application/json');

  const state = api.getState() as RootState;
  const vaultKey = state.auth.vaultKey;


  // TODO This doesn't go off when getting the very first profile details
  // need to move the key fetch into its own seperate call and call that BEFORE profile details
  // otherwise the profile details itself would not be encrypted.
  if (vaultKey && "function" == typeof window.pqcEncrypt) {

    const jsonBody = JSON.stringify(args.body) || "{}";
    const encryptedBody = window.pqcEncrypt(vaultKey, jsonBody);


    println({ jsonBody });

    if (encryptedBody && encryptedBody.blob) {
      const isMutation = ['POST', 'PUT', 'PATCH'].includes(args.method || '');

      if (isMutation) {
        // === CASE 1: Mutation (Send in Body) ===
        const bstring = atob(encryptedBody.blob);
        const len = bstring.length;
        const bbytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bbytes[i] = bstring.charCodeAt(i);
        }

        args.body = bbytes;
        (args.headers as Headers).set('Content-Type', 'application/x-awayto-vault');
      } else {
        // === CASE 2: Query (Send in Header) ===
        // We cannot set args.body for GET.
        // We pass the Base64 blob directly in a custom header.
        (args.headers as Headers).set('X-Awayto-Vault', encryptedBody.blob);
        println("hello");
      }

      extraOptions.vaultSecret = encryptedBody.secret;
    }
  }

  let result = await baseQuery(args, api, extraOptions);

  if (extraOptions.vaultSecret && result.meta?.response?.headers.get('Content-Type') === 'application/x-awayto-vault') {
    println("Printed", result)
    const encryptedBlob = result.data as string;
    const decryptedJson = window.pqcDecrypt(encryptedBlob, extraOptions.vaultSecret);
    result.data = JSON.parse(decryptedJson);
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
