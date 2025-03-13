export type IField = Record<string, string | boolean> & {
  i?: string; // id
  l: string; // label
  v?: string; // value
  h?: string; // helperText
  x?: string; // text
  t?: string; // input type
  d?: string; // defaultValue
  r?: boolean; // required
};

/**
 * @category Form
 * @purpose contains all Fields in all rows of a Form
 */
export type IFormTemplate = Record<string, IField[]>;

/**
 * @category Form
 * @purpose used during Quote submission to record the actual values users typed into the Form
 */
export type IFormSubmission = Record<string, string[]>;

/**
 * @category Form
 * @purpose container for specific Form Versions that are submitting during a Quote request
 */
export type IFormVersionSubmission = {
  id?: string;
  formVersionId: string;
  submission: IFormSubmission;
};

/**
 * @category Form
 * @purpose tracks the different versions of Forms throughout their history
 */
export type IFormVersion = {
  id: string;
  formId: string;
  form: IFormTemplate;
  submission?: IFormSubmission;
  createdOn: string;
};

/**
 * @category Form
 * @purpose models the base container of a form that Group users create for Services
 */
export type IForm = {
  id: string;
  name: string;
  version: IFormVersion;
  createdOn: string;
};
