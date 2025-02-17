import { FetchArgs, BaseQueryFn, createApi, fetchBaseQuery, FetchBaseQueryError, retry } from '@reduxjs/toolkit/query/react'

import { keycloak, authSlice, refreshToken } from './auth';
import { utilSlice } from './util';

const {
  REACT_APP_APP_HOST_URL,
} = process.env as { [prop: string]: string };

const setAuthenticated = authSlice.actions.setAuthenticated;
const setSnack = utilSlice.actions.setSnack;

const baseQuery = retry(fetchBaseQuery({
  baseUrl: REACT_APP_APP_HOST_URL + "/api",
  prepareHeaders(headers) {
    if (!keycloak.token) {
      throw 'no token for api fetch';
    }

    headers.set('Authorization', 'Bearer ' + keycloak.token);
    return headers
  },
}));

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError> = async (args, api) => {

  await refreshToken();

  let result = await baseQuery(args, api, {});

  if (result.error?.data) {
    api.dispatch(setSnack({ snackOn: result.error.data as string }));
  }

  return result;
}

export type CustomBaseQuery = typeof customBaseQuery;

export const siteApiTemplate = createApi({
  baseQuery: customBaseQuery,
  endpoints: () => ({}),
});


