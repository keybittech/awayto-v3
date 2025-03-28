import { createListenerMiddleware, Middleware } from '@reduxjs/toolkit';

import { validSlice } from './valid';
import { RootState } from './store';
import { ConfirmActionProps, IUtil } from './util';

export const listenerMiddleware = createListenerMiddleware();
listenerMiddleware.startListening({
  actionCreator: validSlice.actions.setValid,
  effect: (_, listenerApi) =>
    localStorage.setItem(
      "validations",
      JSON.stringify((listenerApi.getState() as RootState).valid)
    )
});

export type ConfirmActionType = (...props: ConfirmActionProps) => void | Promise<void>;
export type ActionRegistry = Record<string, ConfirmActionType>;
const actionRegistry: ActionRegistry = {};

function registerAction(id: string, action: ConfirmActionType): void {
  actionRegistry[id] = action;
}

export function getUtilRegisteredAction(id: string): ConfirmActionType {
  return actionRegistry[id];
}

export const customUtilMiddleware: Middleware = _ => next => action => {
  const a = action as { type: string, payload: Partial<IUtil> };
  if (a.type.includes('openConfirm')) {
    const { confirmEffect, confirmAction } = a.payload;
    if (confirmEffect && confirmAction) {
      registerAction(btoa(confirmEffect), confirmAction)
      a.payload.confirmAction = undefined;
    }
  }

  return next(action);
}
