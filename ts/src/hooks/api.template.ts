import { FetchArgs, BaseQueryFn, createApi, fetchBaseQuery, FetchBaseQueryError } from '@reduxjs/toolkit/query/react'

import { newAuthSlice, refreshToken } from './auth';
import { newUtilSlice } from './util';

const {
  REACT_APP_APP_HOST_URL,
} = process.env as { [prop: string]: string };

const setAuthenticated = newAuthSlice.actions.setAuthenticated;
const setSnack = newUtilSlice.actions.setSnack;

const baseQuery = fetchBaseQuery({
  baseUrl: REACT_APP_APP_HOST_URL + "/api",
  prepareHeaders(headers) {
    const token = localStorage.getItem('kc_token') as string;

    if (!token) {
      throw "no token for api fetch"
    }

    headers.set('Authorization', 'Bearer ' + token);
    return headers
  },
})

const customBaseQuery: BaseQueryFn<FetchArgs, unknown, FetchBaseQueryError> = async (args, api) => {
  let result = await baseQuery(args, api, {});

  if (result.error) {
    if (result.error.status === "FETCH_ERROR") {
      await refreshToken();
      result = await baseQuery(args, api, {});
      if (result.error) {
        api.dispatch(setAuthenticated({ authenticated: false }))
      }
    } else if (result.error.data) {
      api.dispatch(setSnack({ snackOn: result.error.data as string }));
    }
  }

  return result;
}

export type CustomBaseQuery = typeof customBaseQuery;

export const siteApiTemplate = createApi({
  baseQuery: customBaseQuery,
  endpoints: () => ({}),
})
