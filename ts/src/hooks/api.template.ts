import { FetchArgs, BaseQueryFn, createApi, fetchBaseQuery, FetchBaseQueryError, retry } from '@reduxjs/toolkit/query/react'

import { keycloak, refreshToken } from './auth';
import { utilSlice } from './util';

const {
  VITE_REACT_APP_APP_HOST_URL,
} = import.meta.env;

const setSnack = utilSlice.actions.setSnack;

const baseQuery = fetchBaseQuery({
  baseUrl: VITE_REACT_APP_APP_HOST_URL + "/api",
  prepareHeaders(headers) {
    if (!keycloak.token) {
      throw 'no token for api fetch';
    }

    headers.set('Authorization', 'Bearer ' + keycloak.token);
    return headers
  },
});

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError> = async (args, api) => {

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

  return result;
}

export type CustomBaseQuery = typeof customBaseQuery;

export const siteApiTemplate = createApi({
  baseQuery: customBaseQuery,
  endpoints: () => ({}),
});


