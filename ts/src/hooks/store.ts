import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import { configureStore, Middleware } from '@reduxjs/toolkit';
import { MutationDefinition, QueryDefinition, setupListeners } from '@reduxjs/toolkit/query';
import { siteApi } from './api';
import { ConfirmActionProps, encodeVal, IUtil, utilSlice } from './util';
import { authSlice } from './auth';
import { CustomBaseQuery } from './api.template';
import { TypedUseQueryHookResult } from '@reduxjs/toolkit/query/react';

export type ConfirmActionType = (...props: ConfirmActionProps) => void | Promise<void>;
export type ActionRegistry = Record<string, ConfirmActionType>;
const actionRegistry: ActionRegistry = {};

function registerAction(id: string, action: ConfirmActionType): void {
  actionRegistry[id] = action;
}

export function getUtilRegisteredAction(id: string): ConfirmActionType {
  return actionRegistry[id];
}
const customUtilMiddleware: Middleware = _ => next => action => {
  const a = action as { type: string, payload: Partial<IUtil> };
  if (a.type.includes('openConfirm')) {
    const { confirmEffect, confirmAction } = a.payload;
    if (confirmEffect && confirmAction) {
      registerAction(encodeVal(confirmEffect), confirmAction)
      a.payload.confirmAction = undefined;
    }
  }

  return next(action);
}

export const store = configureStore({
  reducer: {
    [siteApi.reducerPath]: siteApi.reducer,
    util: utilSlice.reducer,
    auth: authSlice.reducer,
  },
  middleware(getDefaultMiddleware) {
    return getDefaultMiddleware().concat([
      siteApi.middleware,
      customUtilMiddleware,
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
