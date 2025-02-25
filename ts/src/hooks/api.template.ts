import { FetchArgs, BaseQueryFn, createApi, fetchBaseQuery, FetchBaseQueryError, retry } from '@reduxjs/toolkit/query/react'

import { keycloak, refreshToken } from './auth';
import { utilSlice } from './util';
import { RootState } from './store';

const {
  VITE_REACT_APP_APP_HOST_URL,
} = import.meta.env;

const setSnack = utilSlice.actions.setSnack;

const modifiedResources: Record<string, string> = {};

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError> = async (args, api) => {

  const baseQuery = fetchBaseQuery({
    baseUrl: VITE_REACT_APP_APP_HOST_URL + "/api",
    prepareHeaders(headers) {
      if (!keycloak.token) {
        throw 'no token for api fetch';
      }

      headers.set('X-TZ', Intl.DateTimeFormat().resolvedOptions().timeZone);
      headers.set('Authorization', 'Bearer ' + keycloak.token);

      const lastModified = modifiedResources[args.url];
      if (lastModified) {
        headers.set('If-Modified-Since', lastModified);
      }

      return headers
    },
  });

  await refreshToken();

  let result = await baseQuery(args, api, {});

  switch (result.error?.status) {
    case "PARSING_ERROR":
      api.dispatch(setSnack({ snackOn: result.error.data as string }));
      break
    case "FETCH_ERROR":
      result = await retry(baseQuery)(args, api, {});
      break
    default:
  }

  const lastModified = result.meta?.response?.headers.get('last-modified');
  if (lastModified) {
    modifiedResources[args.url] = lastModified;
  }

  if (result.meta?.response?.status === 304 && api.queryCacheKey) {
    const state = api.getState() as RootState;
    const cachedData = state.api.queries[api.queryCacheKey]?.data;

    if (cachedData) {
      return { data: cachedData };
    }
  }
  return result;
}

export type CustomBaseQuery = typeof customBaseQuery;

export const siteApiTemplate = createApi({
  refetchOnMountOrArgChange: 180,
  baseQuery: customBaseQuery,
  endpoints: () => ({}),
});


