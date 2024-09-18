import { createSlice } from "@reduxjs/toolkit";

export type ITheme = {
  variant: string;
}

type ThemePayload = [
  state: ITheme,
  action: { payload: Partial<ITheme> }
]

export const themeConfig = {
  name: 'theme',
  initialState: {
    variant: 'dark'
  } as ITheme,
  reducers: {
    setTheme: (...[state, { payload: { variant } }]: ThemePayload) => {
      if (variant) {
        state.variant = variant;
      }
    },

  },
};

export const themeSlice = createSlice(themeConfig);

