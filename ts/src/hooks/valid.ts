import { createSlice } from "@reduxjs/toolkit";

export type IValidationSchemas = {
  group?: boolean;
  roles?: boolean;
  service?: boolean;
  schedule?: boolean;
  serviceIntake?: boolean;
  tierIntake?: boolean;
};

export type IValidationAreas = {
  onboarding: IValidationSchemas;
  requestQuote: IValidationSchemas;
};

export type IValidActionPayload = {
  area: keyof IValidationAreas;
  schema: keyof IValidationSchemas;
  valid: boolean;
};

export const initialState: IValidationAreas = {
  onboarding: {},
  requestQuote: {},
}

export const validSlice = createSlice({
  name: 'validations',
  initialState,
  reducers: {
    setValid: (state, action: { payload: IValidActionPayload }) => {
      state[action.payload.area][action.payload.schema] = action.payload.valid;
    }
  }
});
