import { MutationDefinition, QueryDefinition } from '@reduxjs/toolkit/query';
import { FetchArgs, BaseQueryFn, fetchBaseQuery, FetchBaseQueryError, TypedUseQueryHookResult, retry, reactHooksModule, buildCreateApi, coreModule } from '@reduxjs/toolkit/query/react'

import { utilSlice } from './util';

const {
  VITE_REACT_APP_APP_HOST_URL,
} = import.meta.env;

const setSnack = utilSlice.actions.setSnack;

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError> = async (args, api) => {
  const baseQuery = fetchBaseQuery({
    timeout: 30000,
    cache: 'no-cache',
    mode: 'same-origin',
    credentials: 'include',
    baseUrl: VITE_REACT_APP_APP_HOST_URL + "/api",
    prepareHeaders(headers) {
      headers.set('X-Tz', Intl.DateTimeFormat().resolvedOptions().timeZone);
      headers.set('Content-Type', 'application/json');
      return headers;
    },
  });

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
