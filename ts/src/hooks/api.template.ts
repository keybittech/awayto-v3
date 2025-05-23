import { MutationDefinition, QueryDefinition } from '@reduxjs/toolkit/query';
import { FetchArgs, BaseQueryFn, fetchBaseQuery, FetchBaseQueryError, TypedUseQueryHookResult, retry, reactHooksModule, buildCreateApi, coreModule } from '@reduxjs/toolkit/query/react'

import { keycloak, refreshToken, setAuthHeaders } from './keycloak';
import { RootState } from './store';
import { utilSlice } from './util';

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

      headers = setAuthHeaders(headers);

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
    if (lastModified != modifiedResources[args.url]) {
      return result;
    } else {
      modifiedResources[args.url] = lastModified;
    }
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
