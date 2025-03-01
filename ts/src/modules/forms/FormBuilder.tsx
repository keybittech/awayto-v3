import React, { useCallback, useMemo, useState, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import Switch from '@mui/material/Switch';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';

import SettingsIcon from '@mui/icons-material/Settings';

import { IField, IFormVersion, deepClone, targets } from 'awayto/hooks';

import Field from './Field';

// text

// Single line text input.

// textarea

// Multiple line text input.

// textarea

// select

// Common single select input. See description how to configure options below.

// select

// select-radiobuttons

// Single select input through group of radio buttons. See description how to configure options below.

// group of input

// multiselect

// Common multiselect input. See description how to configure options below.

// select

// multiselect-checkboxes

// Multiselect input through group of checkboxes. See description how to configure options below.

// group of input

// html5-email

// Single line text input for email address based on HTML 5 spec.

// html5-tel

// Single line text input for phone number based on HTML 5 spec.

// html5-url

// Single line text input for URL based on HTML 5 spec.

// html5-number

// Single line input for number (integer or float depending on step) based on HTML 5 spec.

// html5-range

// Slider for number entering based on HTML 5 spec.

// html5-datetime-local

// Date Time input based on HTML 5 spec.

// html5-date

// Date input based on HTML 5 spec.

// html5-month

// Month input based on HTML 5 spec.

// html5-week

// Week input based on HTML 5 spec.

// html5-time

// Time input based on HTML 5 spec.

interface FormBuilderProps extends IComponent {
  version: IFormVersion;
  setVersion(value: IFormVersion): void;
  editable: boolean;
}

export default function FormBuilder({ version, setVersion, editable = true }: FormBuilderProps): React.JSX.Element {

  const [rows, setRows] = useState({} as Record<string, IField[]>);
  const [cell, setCell] = useState({} as IField & Object);
  const [position, setPosition] = useState({ row: '', col: 0 });

  const cellSelected = cell.hasOwnProperty('l');
  const inputTypes = ['text', 'date', 'time'];

  useEffect(() => {
    if (Object.keys(version.form).length) {
      setRows({ ...version.form });
    }
  }, [version]);

  const updateData = useCallback((newRows: typeof rows) => {
    const rowSet = { ...rows, ...newRows };
    setRows(rowSet);
    setVersion({ ...version, form: rowSet });

  }, [rows, version]);

  const rowKeys = useMemo(() => Object.keys(rows), [rows]);

  const addRow = useCallback(() => updateData({ ...rows, [(new Date()).getTime().toString()]: [makeField()] }), [rows]);

  const addCol = useCallback((row: string) => updateData({ ...rows, [row]: Array.prototype.concat(rows[row], [makeField()]) }), [rows]);

  const delCol = useCallback((row: string, col: number) => {
    rows[row].splice(col, 1);
    if (!rows[row].length) delete rows[row];
    updateData({ ...rows });
  }, [rows]);

  const makeField = useCallback((): IField => ({ l: '', t: 'text', h: '', r: false, v: '', x: '' }), []);

  const setCellAttr = useCallback((value: string, attr: string) => {
    if (cell && Object.keys(cell).length) {
      const newCell = {
        ...cell,
        [attr]: value
      };

      setCell(newCell);

      const newRows = deepClone(rows);

      if (newRows[position.row] && newRows[position.row][position.col]) {
        newRows[position.row][position.col] = newCell;
        updateData({ ...newRows });
      }
    }
  }, [rows, cell]);

  return <Grid container spacing={2}>

    {Object.keys(rows).length < 3 && <Grid size={12}>
      <Button
        {...targets(`form build add row`, `add a row to the form`)}
        variant="outlined"
        fullWidth
        onClick={addRow}
      >add row</Button>
    </Grid>}
    {Object.keys(rows).length > 0 && <Grid size={12}>
      <Typography variant="caption">Click the <SettingsIcon fontSize="small" sx={{ verticalAlign: 'bottom' }} /> icon to edit fields.</Typography>
    </Grid>}

    <Grid role="rowgroup" id="form-builder-fieldset" size="grow" container spacing={2} sx={{ alignItems: 'start' }}>
      {rowKeys.map((rowId, i) => <Grid key={`form_fields_row_${i}`} size={12}>
        <Grid role="row" container spacing={2}>
          {rows[rowId].length < 3 && <Grid className="add-column-btn" size={{ xs: 12, md: 2 }}>
            <Grid container direction="column" sx={{ placeItems: 'center', height: '100%' }}>
              <Button
                {...targets(`form build add column row ${i + 1}`, `add a column to row ${i + 1}`)}
                fullWidth
                variant="outlined"
                color="warning"
                sx={{ alignItems: 'center', display: 'flex', flex: 1 }}
                onClick={() => addCol(rowId)}
              >add column</Button>
              {/* <ButtonBase sx={{ display: 'flex', padding: '2px', backgroundColor: 'rgba(255, 0, 0, .1)' }} onClick={() => delRow(rowId)}>- row</ButtonBase> */}
            </Grid>
          </Grid>}
          <Grid size={{ xs: 12, md: rows[rowId].length < 3 ? 10 : 12 }}>
            <Grid container spacing={2}>
              {rows[rowId].map((field, j) => {
                return <Grid role="gridcell" size={12 / rows[rowId].length} key={`form_fields_cell_${i + 1}_${j}`}>
                  <Field
                    defaultDisplay
                    editable={false}
                    field={field}
                    settingsBtn={
                      <IconButton
                        {...targets(`form build edit field ${field.l}`, `edit field labeled ${field.l}`)}
                        sx={{ color: position.row == rowId && position.col == j ? 'white' : 'gray' }}
                        onClick={() => {
                          setCell(field);
                          setPosition({ row: rowId, col: j })
                        }}
                      >
                        <SettingsIcon />
                      </IconButton>
                    }
                  />
                </Grid>
              })}
            </Grid>
          </Grid>
        </Grid>
      </Grid>)}
    </Grid>

    {
      editable && cellSelected && <Grid>
        <Divider orientation="vertical" />
      </Grid>
    }

    {
      editable && cellSelected && <Grid size={4}>
        <Grid container spacing={2} direction="column">

          <Grid>
            <Grid container alignItems="center">
              <Grid sx={{ display: 'flex', flex: 1 }}>
                <Typography variant="body2">Field Attributes</Typography>
              </Grid>
              <Grid>
                <Button
                  {...targets(`form build close editing`, `close the field editing panel`)}
                  variant="text"
                  onClick={() => {
                    setPosition({ row: '', col: 0 })
                    setCell({} as IField);
                  }}
                >Close</Button>
              </Grid>
            </Grid>
          </Grid>

          <Grid>
            <TextField
              {...targets(`form build select field type`, `Field Type`, `modify the field's data type`)}
              fullWidth
              select
              value={cell.t}
              onChange={e => setCellAttr(e.target.value, 't')}
            >
              <MenuItem key={`field_type_1`} value={'text'}>Textfield</MenuItem>
              <MenuItem key={`field_type_2`} value={'date'}>Date</MenuItem>
              <MenuItem key={`field_type_3`} value={'time'}>Time</MenuItem>
              <MenuItem key={`field_type_4`} value={'labelntext'}>Label and Text</MenuItem>
            </TextField>
          </Grid>

          <Grid>
            <TextField
              fullWidth
              autoFocus
              {...targets(`form build field label input ${position.row} ${position.col}`, `Label`, `change the label that will appear above the form field`)}
              type="text"
              helperText="Required."
              value={cell.l}
              onChange={e => setCellAttr(e.target.value, 'l')}
            />
          </Grid>

          {'labelntext' === cell.t && <Grid>
            <TextField
              {...targets(`form build field text`, `Text`, `change the text of a label and text style form field`)}
              fullWidth
              type="text"
              value={cell.x}
              onChange={e => setCellAttr(e.target.value, 'x')}
            />
          </Grid>}

          {!inputTypes.includes(cell.t || '') ? <></> : <>
            <Grid>
              <TextField
                fullWidth
                {...targets(`form build field helper text`, `Helper Text`, `change the text that will appear as helper text below the form field`)}
                type="text"
                value={cell.h}
                onChange={e => setCellAttr(e.target.value, 'h')}
              />
            </Grid>

            <Grid>
              <TextField
                fullWidth
                {...targets(`form build field default value`, `Default Value`, `change the default value of the form field`)}
                type={cell.t}
                value={cell.v}
                onChange={e => setCellAttr(e.target.value, 'v')}
                slotProps={{
                  inputLabel: {
                    shrink: true
                  }
                }}
              />
            </Grid>

            <Grid>
              <Typography variant="body1">Required</Typography>
              <Switch
                {...targets(`form build field required`, `set the field to be required or not during form submission`)}
                value={cell.r}
                checked={cell.r}
                onChange={() => {
                  rows[position.row][position.col].r = !cell.r;
                  updateData({ ...rows })
                }} />
            </Grid>
          </>}

          <Grid>
            <Button
              {...targets(`form build delete field`, `delete this field from the form`)}
              fullWidth
              color="error"
              onClick={() => {
                setPosition({ row: '', col: 0 })
                delCol(position.row, position.col);
                setCell({} as IField);
              }}
            >Delete</Button>
          </Grid>
        </Grid>
      </Grid>
    }

  </Grid >;
}
