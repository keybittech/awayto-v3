import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';

import { siteApi } from './api';
import { utilSlice } from './util';
import { validSlice, initialState as initialValidState } from './valid';
import { authSlice } from './auth';
import { customUtilMiddleware, listenerMiddleware } from './middleware';

// Here we can undo any tags done by the rtk openapi generation to prevent unnecessary refetching
siteApi.enhanceEndpoints({
  endpoints: {
    groupScheduleServiceGetGroupScheduleByDate: {
      providesTags: []
    },
    bookingServicePatchBookingRating: {
      invalidatesTags: []
    }
  }
});

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
