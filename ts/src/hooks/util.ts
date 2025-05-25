import { createSlice } from '@reduxjs/toolkit';

import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import relativeTime from 'dayjs/plugin/relativeTime';
import utc from 'dayjs/plugin/utc';
import isoWeek from 'dayjs/plugin/isoWeek';
import timezone from 'dayjs/plugin/timezone';
import updateLocale from 'dayjs/plugin/updateLocale';

dayjs.extend(duration);
dayjs.extend(relativeTime);
dayjs.extend(utc);
dayjs.extend(isoWeek);
dayjs.extend(timezone);
dayjs.extend(updateLocale);

dayjs.updateLocale('en', {
  weekStart: 1
});

export { dayjs };

export const isExternal = (loc: string) => {
  return loc.startsWith('/app/ext/');
}

export type ConfirmActionProps = [
  approval?: boolean,
]

export type ConfirmAction = (...props: ConfirmActionProps) => void | Promise<void>;

export type IUtil = {
  confirmAction: ConfirmAction;
  confirmActionId: string;
  isConfirming: boolean;
  confirmEffect: string;
  confirmSideEffect?: {
    approvalAction: string;
    approvalEffect: string;
    rejectionAction: string;
    rejectionEffect: string;
  };
  isLoading: boolean;
  loadingMessage: string;
  error: Error;
  canSubmitAssignments: boolean;
  snackType: 'success' | 'info' | 'warning' | 'error';
  snackOn: string;
  snackRequestId: string;
  perPage: number;
}

type UtilPayload = [
  state: IUtil,
  action: { payload: Partial<IUtil> }
]

export const utilConfig = {
  name: 'util',
  initialState: {
    snackOn: '',
    isLoading: false,
    loadingMessage: '',
    isConfirming: false,
  } as IUtil,
  reducers: {
    openConfirm: (...[state, { payload: { confirmAction, confirmEffect, confirmSideEffect } }]: UtilPayload) => {
      if (confirmEffect) {
        state.isConfirming = true;
        state.confirmAction = confirmAction as ConfirmAction;
        state.confirmEffect = confirmEffect;
        state.confirmSideEffect = confirmSideEffect;
        state.confirmActionId = btoa(confirmEffect);
      }
    },
    closeConfirm: (...[state]: UtilPayload) => {
      state.isConfirming = false;
    },
    setLoading: (...[state, { payload: { isLoading, loadingMessage } }]: UtilPayload) => {
      state.isLoading = isLoading || !state.isLoading;
      state.loadingMessage = loadingMessage || '';
    },
    setSnack: (...[state, action]: UtilPayload) => {
      state.snackOn = action.payload.snackOn || '';
      state.snackType = action.payload.snackType || 'info';
      state.snackRequestId = action.payload.snackRequestId || '';
    }
  },
};

export const utilSlice = createSlice(utilConfig);

let arbitraryCounter = 0;

function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
    (+c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> +c / 4).toString(16)
  );
}

export function nid(uuid?: string) {
  if ('random' === uuid) {
    return uuidv4();
  }
  arbitraryCounter++;
  return arbitraryCounter;
}

export function isString(str?: string | unknown): str is string {
  return 'string' == typeof str as string;
}

export function isStringArray(str?: string | string[]): str is string[] {
  return (str as string[]).forEach !== undefined;
}

export const targets = (name: string, label?: string, aria?: string): {
  name: string;
  id: string;
  label: string;
  'aria-label': string;
} => {
  return {
    name,
    id: toSnakeCase(name),
    label: label || '',
    'aria-label': aria || label || ''
  };
}

export const toSnakeCase = (name: string): string => {
  return name.replace(/\W+/g, " ").split(" ").join('_').toLowerCase();
};

export const toTitleCase = (name: string): string => {
  return name.substring(1).replace(/([A-Z])/g, " $1").replace(/^./, (str) => str.toUpperCase());
};

export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => deepClone(item) as T) as unknown as T;
  }

  const result: Record<string, unknown> = {};

  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      result[key] = deepClone((obj as Record<string, unknown>)[key]);
    }
  }

  return result as T;
}

export function getRelativeCoordinates(event: MouseEvent | React.MouseEvent<HTMLCanvasElement> | React.Touch, canvas: HTMLCanvasElement) {
  const rect = canvas.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const y = event.clientY - rect.top;
  return { x, y };
}

