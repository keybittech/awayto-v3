import { MutationDefinition, QueryDefinition, RetryOptions } from '@reduxjs/toolkit/query';
import { FetchArgs, BaseQueryFn, fetchBaseQuery, FetchBaseQueryError, TypedUseQueryHookResult, retry, reactHooksModule, buildCreateApi, coreModule } from '@reduxjs/toolkit/query/react'

import { utilSlice } from './util';
import { RootState } from './store';
import { decryptCacheData, encryptCacheData } from './session_crypto';
import { authSlice } from './auth';

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
  responseHandler: async (response) => {
    if (304 == response.status) {
      return null;
    }
    const contentType = response.headers.get('Content-Type');
    if (contentType === 'application/x-awayto-vault') {
      return response.text();
    }
    return response.json();
  },
});

type CustomExtraOptions = {
  vaultSecret?: string;
  hasRetriedVault?: boolean;
} & RetryOptions;

const getCacheKey = (url: string, sid?: string) => {
  return `RTK_CACHE_${sid || 'NO_SESS'}${url}`;
}

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError, CustomExtraOptions> = async (xargs, api, extraOptions = {}) => {
  // handle originalArgs versus currentArgs when needing to retry, as we need to preserve whatever encryption occurs
  const oArgs = typeof xargs === 'string' ? { url: xargs } : { ...xargs };
  const cArgs = typeof xargs === 'string' ? { url: xargs, headers: new Headers() } : { ...xargs, headers: new Headers(xargs.headers as HeadersInit) };

  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;

  (cArgs.headers as Headers).set('X-Tz', tz);

  const state = api.getState() as RootState;
  const { sessionId, vaultKey } = state.auth;
  const isMutation = ['POST', 'PUT', 'PATCH'].includes(cArgs.method || '');
  const cacheKey = getCacheKey(cArgs.url, sessionId);

  if (vaultKey && sessionId && "function" == typeof window.pqcEncrypt) {

    if (isMutation) {

      if (!cArgs.body) {
        cArgs.body = {};
      }

      const encryptedBody = window.pqcEncrypt(vaultKey, JSON.stringify(cArgs.body), sessionId);

      if (encryptedBody && encryptedBody.blob) {
        const bstring = atob(encryptedBody.blob);
        const len = bstring.length;
        const bbytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bbytes[i] = bstring.charCodeAt(i);
        }

        cArgs.body = bbytes;
        (cArgs.headers as Headers).set('Content-Type', 'application/x-awayto-vault');
        (cArgs.headers as Headers).set('X-Original-Content-Type', 'application/json');

        extraOptions.vaultSecret = encryptedBody.secret;
      }
    } else {
      const encryptedBody = window.pqcEncrypt(vaultKey, " ", sessionId);
      if (encryptedBody && encryptedBody.blob) {
        (cArgs.headers as Headers).set('X-Awayto-Vault', encryptedBody.blob);
        extraOptions.vaultSecret = encryptedBody.secret;
      }
    }
  } else {
    (cArgs.headers as Headers).set('Content-Type', 'application/json');
  }

  let cachedData: { etag: string, data: any, timestamp: number } | null = null;

  if (!isMutation && vaultKey && sessionId) {
    try {
      const cached = sessionStorage.getItem(cacheKey);
      if (cached) {
        cachedData = await decryptCacheData(cached, vaultKey, sessionId);
        if (cachedData?.etag) {
          (cArgs.headers as Headers).set('If-None-Match', cachedData.etag);
        }
      }
    } catch (e) {
      console.warn("failed storage check, err: ", e);
    }
  }

  let result = await baseQuery(cArgs, api, extraOptions);

  if ('PARSING_ERROR' === result.error?.status && extraOptions.vaultSecret) { // key may have changed

    if (!extraOptions.hasRetriedVault) {
      const refreshResult = await baseQuery({ url: '/v1/vault/key', method: 'GET' }, api, {});

      if (refreshResult.data) {
        const newKey = (refreshResult.data as { key: string }).key;
        if (newKey.length) {
          api.dispatch(authSlice.actions.setVault({ vaultKey: newKey }));
          return customBaseQuery(oArgs, api, { ...extraOptions, hasRetriedVault: true });
        }
      }
    }

    api.dispatch(setSnack({ snackOn: result.error.data as string }));
  }

  if (401 === result.error?.status) { // client is no longer authenticated
    localStorage.clear();
    sessionStorage.clear();
    window.location.href = `/auth/login?tz=${tz}`
  }

  if (304 === result.meta?.response?.status) { // request was cached
    if (cachedData && cachedData.data) {
      // Cached on server and locally
      result.data = cachedData.data;
      delete result.error;
    } else {
      // Cached on server but not locally
      (cArgs.headers as Headers).delete('If-None-Match');
      result = await baseQuery(cArgs, api, extraOptions);
    }
  }

  if (extraOptions.vaultSecret && sessionId && result.meta?.response?.headers.get('Content-Type') === 'application/x-awayto-vault') {
    if (typeof result.data === 'string') {
      const decryptedJson = window.pqcDecrypt(result.data.trim(), extraOptions.vaultSecret, sessionId);
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

  if (!isMutation && result.data && 200 === result.meta?.response?.status) {
    const etag = result.meta.response.headers.get('ETag');
    if (etag && vaultKey && sessionId) {
      try {
        const cacheResult = {
          etag,
          data: result.data,
          timestamp: Date.now(),
        };
        const encryptedEntry = await encryptCacheData(cacheResult, vaultKey, sessionId);
        if (encryptedEntry) {
          sessionStorage.setItem(cacheKey, encryptedEntry);
        }
      } catch (e) {
        console.warn("failed to store etagged response, check limits, err: ", e);
      }
    }
  }

  // switch (result.error?.status) {
  //   case "PARSING_ERROR":
  //     api.dispatch(setSnack({ snackOn: result.error.data as string }));
  //     break
  //   case "FETCH_ERROR":
  //     result = await retry(baseQuery)(args, api, {});
  //     break
  //   default:
  // }

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
  baseQuery: retry(customBaseQuery, { maxRetries: 3 }),
  endpoints: () => ({}),
});

export type SiteMutation<TQueryArg, TResultType> = MutationDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;
export type SiteQuery<TQueryArg, TResultType> = QueryDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;

export type UseSiteQuery<T, R> = TypedUseQueryHookResult<R, T, CustomBaseQuery>;
