import React, { useCallback, useMemo, useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import Switch from '@mui/material/Switch';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';

import SettingsIcon from '@mui/icons-material/Settings';

import { IField, IFormVersion, deepClone, nid, targets, toSnakeCase } from 'awayto/hooks';

import Field from './Field';

const fieldTypes = [
  { variant: 'labelntext', name: 'Label with Text' },
  { variant: 'text', name: 'Textfield' },
  { variant: 'boolean', name: 'Checkbox (Yes/No)' },
  { variant: 'multi-select', name: 'Checkbox Group (Select Multiple)' },
  { variant: 'single-select', name: 'Radio Group (Select One)' },
  { variant: 'date', name: 'Date' },
  { variant: 'time', name: 'Time' },
];

interface FormBuilderProps extends IComponent {
  version: IFormVersion;
  setVersion(value: IFormVersion): void;
  editable: boolean;
}

export default function FormBuilder({ version, setVersion, editable = true }: FormBuilderProps): React.JSX.Element {

  const [rows, setRows] = useState({} as Record<string, IField[]>);
  const [cell, setCell] = useState({} as IField & Object);
  const [position, setPosition] = useState({ row: '', col: 0 });
  const [opt, setOpt] = useState<string>('');

  const cellSelected = cell.hasOwnProperty('l');

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

  const makeField = useCallback((): IField => ({ i: nid('random'), l: '', t: 'text', h: '', r: false, v: '', x: '', o: [] }), []);

  const addRow = useCallback(() => updateData({ ...rows, [(new Date()).getTime().toString()]: [makeField()] }), [rows]);

  const addCol = useCallback((row: string) => updateData({ ...rows, [row]: [...rows[row], makeField()] }), [rows]);

  const delCol = useCallback((row: string, col: number) => {
    const newRows = deepClone(rows);
    newRows[row].splice(col, 1);
    if (!newRows[row].length) delete newRows[row];
    updateData(newRows);
  }, [rows]);

  const updateCell = (newCell: IField) => {
    setCell(newCell);

    const newRows = deepClone(rows);

    if (newRows[position.row] && newRows[position.row][position.col]) {
      newRows[position.row][position.col] = newCell;
      updateData({ ...newRows });
    }
  }

  const addOpt = () => {
    const val = toSnakeCase(opt.trim());
    if (!val) return;

    const newOpt = { i: nid('random'), l: opt.trim(), v: val };
    updateCell({
      ...cell,
      o: [...(cell.o || []), newOpt]
    })
    setOpt('');
  }

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
            </Grid>
          </Grid>}
          <Grid size={{ xs: 12, md: rows[rowId].length < 3 ? 10 : 12 }}>
            <Grid container spacing={2}>
              {rows[rowId].map((field, j) => {
                return <Grid role="gridcell" size={12 / rows[rowId].length} key={`form_fields_cell_${i + 1}_${j}`}>
                  <Field
                    disabled={true}
                    field={field}
                    onChange={_ => { }}
                    settingsBtn={
                      <IconButton
                        {...targets(`form build edit field ${field.l}`, `edit field labeled ${field.l}`)}
                        sx={{ color: position.row == rowId && position.col == j ? 'blue' : 'gray' }}
                        onClick={() => {
                          setCell(field);
                          setPosition({ row: rowId, col: j });
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
                    setCell({} as IField);
                    setPosition({ row: '', col: 0 });
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
              onChange={e => {
                const newType = e.target.value as IField['t'];
                updateCell({
                  ...cell,
                  t: newType,
                  v: 'multi-select' === newType ? [] : ('boolean' === newType ? false : '')
                });
              }}
            >
              {fieldTypes.map((ft, i) => (
                <MenuItem key={`field_type_${i}`} value={ft.variant}>{ft.name}</MenuItem>
              ))}
            </TextField>
          </Grid>

          <Grid>
            <TextField
              {...targets(`form build field label input ${position.row} ${position.col}`, `Label`, `change the label that will appear above the form field`)}
              fullWidth
              required
              autoFocus
              type="text"
              value={cell.l}
              onChange={e => {
                updateCell({
                  ...cell,
                  l: e.target.value
                });
              }}
            />
          </Grid>

          {'labelntext' === cell.t && <Grid>
            <TextField
              {...targets(`form build field text`, `Text`, `change the text of a label and text style form field`)}
              fullWidth
              type="text"
              value={cell.x}
              onChange={e => {
                updateCell({
                  ...cell,
                  x: e.target.value
                });
              }}
            />
          </Grid>}

          {['single-select', 'multi-select'].includes(cell.t || '') && <Grid>
            <TextField
              {...targets(`form build field option`, `Options`, `add options for the ${cell.t} form field`)}
              fullWidth
              label="Add Option"
              placeholder="e.g. Rhetorical Analysis"
              value={opt}
              onChange={e => setOpt(e.target.value)}
              onKeyDown={e => {
                if ('Enter' === e.key) {
                  addOpt();
                }
              }}
              slotProps={{
                input: {
                  endAdornment: <InputAdornment position="end">
                    <Button
                      variant="contained"
                      size="small"
                      {...targets(`form build add option`, `add an option to the ${cell.t} field`)}
                      color="secondary"
                      onClick={addOpt}
                    >
                      <Typography variant="button">Add</Typography>
                    </Button>
                  </InputAdornment>
                }
              }}
            />
            <Box sx={{ mt: 1, display: 'flex', flexWrap: 'wrap', gap: 1 }}>
              {cell.o?.map((opt, i) => (
                <Chip
                  key={opt.i}
                  label={opt.l}
                  color="primary"
                  variant="outlined"
                  onDelete={() => {
                    const newOpts = cell.o?.filter((_, idx) => idx !== i);
                    updateCell({ ...cell, o: newOpts });
                  }}
                />
              ))}
            </Box>
          </Grid>}

          {'labelntext' === cell.t ? <></> : <>
            <Grid>
              <TextField
                fullWidth
                {...targets(`form build field helper text`, `Helper Text`, `change the text that will appear as helper text below the form field`)}
                type="text"
                value={cell.h}
                onChange={e => {
                  updateCell({
                    ...cell,
                    h: e.target.value
                  });
                }}
              />
            </Grid>

            {/* <Grid> */}
            {/*   {['radio', 'checkbox'].includes(cell.t || '') ? */}
            {/*     <TextField */}
            {/*       fullWidth */}
            {/*       {...targets(`form build field default value`, `Default Value`, `change the default value of the form field`)} */}
            {/*       select */}
            {/*       value={cell.v} */}
            {/*       onChange={e => setCellAttr(e.target.value, 'v')} */}
            {/*     > */}
            {/*       {cell.o?.map((opt, i) => ( */}
            {/*         <MenuItem key={`form_opt_default_${i}`} value={opt}>{opt}</MenuItem> */}
            {/*       ))} */}
            {/*     </TextField> : */}
            {/*     <TextField */}
            {/*       fullWidth */}
            {/*       {...targets(`form build field default value`, `Default Value`, `change the default value of the form field`)} */}
            {/*       type={cell.t} */}
            {/*       value={cell.v} */}
            {/*       onChange={e => setCellAttr(e.target.value, 'v')} */}
            {/*       slotProps={{ */}
            {/*         inputLabel: { */}
            {/*           shrink: true */}
            {/*         } */}
            {/*       }} */}
            {/*     />} */}
            {/* </Grid> */}

            <Grid>
              <Typography variant="body1">Required</Typography>
              <Switch
                {...targets(`form build field required`, `set the field to be required or not during form submission`)}
                value={cell.r || false}
                checked={cell.r}
                onChange={() => {
                  updateCell({
                    ...cell,
                    r: !cell.r
                  });
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
