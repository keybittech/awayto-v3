import { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { IField, IFormTemplate, siteApi, targets } from 'awayto/hooks';

// const fieldCharts = {
//   boolean: 'pie',
//   'single-select': 'bar',
//   'multi-select': 'bar',
//   number: 'histogram',
// };

export function FormReport(_: IComponent): React.JSX.Element {

  const { formId } = useParams();
  if (!formId) return <></>;

  const [fieldId, setFieldId] = useState('');

  const { data: formRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId });
  const [getFieldReport] = siteApi.useLazyFormServiceGetFormReportQuery();

  const reportFields = useMemo(() => {
    const versionForm: IFormTemplate = formRequest?.groupForm?.form?.version?.form;
    if (!versionForm) return [];

    const fields: IField[] = [];
    const allowedTypes = ['multi-select', 'single-select', 'boolean', 'number'];

    Object.values(versionForm).forEach(row => {
      row.forEach(field => {
        println({ ft: field.t });
        if (allowedTypes.includes(field.t)) {
          fields.push(field);
        }
      });
    });

    return fields;
  }, [formRequest]);

  const handleSubmit = () => {
    getFieldReport({ formId, fieldId }).unwrap().then(res => {
      println({ res });
    });
  }

  return <>
    <Card variant="outlined">
      <CardHeader
        title={`Form Report: ${formRequest?.groupForm.form?.name}`}
        subheader="Select a field to view a graph of its collected data."
      />

      <CardContent>
        <Grid container alignItems="end">
          <Grid>
            <TextField
              {...targets(`form report field selection`, `Select Field`, `select a field to view its visual report`)}
              sx={{ width: 300 }}
              select
              value={fieldId}
              onChange={e => setFieldId(e.target.value)}
            >
              {fieldId.length && <MenuItem key="unset-selection" value=""><Typography variant="caption">Remove selection</Typography></MenuItem>}
              {reportFields.map((f, i) => <MenuItem key={`report_field_sel${i}`} value={f.i}>{f.l}</MenuItem>)}
            </TextField>
          </Grid>
          <Grid>
          </Grid>
        </Grid>
      </CardContent>

      {!!fieldId.length && <CardActionArea
        {...targets(`form report generate`, `request a generated form using the selected field`)}
        onClick={() => handleSubmit()}
      >
        <Typography sx={{ m: 2, display: 'flex' }} color="secondary" variant="button">Generate</Typography>
      </CardActionArea>}
    </Card>
  </>
}

export default FormReport;
