import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MutationDefinition, QueryDefinition, setupListeners } from '@reduxjs/toolkit/query';
import { TypedUseQueryHookResult } from '@reduxjs/toolkit/query/react';

import { CustomBaseQuery } from './api.template';
import { siteApi } from './api';
import { utilSlice } from './util';
import { validSlice, initialState as initialValidState } from './valid';
import { authSlice } from './auth';
import { customUtilMiddleware, listenerMiddleware } from './middleware';


export const store = configureStore({
  preloadedState: {
    valid: JSON.parse(localStorage.getItem("validations") || "null") ?? initialValidState
  },
  reducer: {
    [siteApi.reducerPath]: siteApi.reducer,
    auth: authSlice.reducer,
    util: utilSlice.reducer,
    valid: validSlice.reducer,
  },
  middleware(getDefaultMiddleware) {
    return getDefaultMiddleware().concat([
      siteApi.middleware,
      customUtilMiddleware,
      listenerMiddleware.middleware,
    ])
  },
});

setupListeners(store.dispatch)

export type AppDispatch = typeof store.dispatch;

export const useAppDispatch: () => AppDispatch = useDispatch;
export interface RootState extends ReturnType<typeof store.getState> { }
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

export type SiteMutation<TQueryArg, TResultType> = MutationDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;
export type SiteQuery<TQueryArg, TResultType> = QueryDefinition<TQueryArg, CustomBaseQuery, 'Root', TResultType, 'api'>;

export type UseSiteQuery<T, R> = TypedUseQueryHookResult<R, T, CustomBaseQuery>
