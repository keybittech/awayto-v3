import { useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import { ChartTypeRegistry } from 'chart.js/auto';

import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { IField, IFormTemplate, IProtoFormDataPoint, siteApi, targets } from 'awayto/hooks';

import FormChart from './FormChart';

const fieldCharts: Partial<Record<IField['t'], string[]>> = {
  boolean: ['bar', 'pie'],
  'single-select': ['bar', 'pie', 'line'],
  'multi-select': ['bar'],
  number: ['bar', 'pie', 'line'],
};

export function FormReport(_: IComponent): React.JSX.Element {

  const { formId } = useParams();
  if (!formId) return <></>;

  const [chartType, setChartType] = useState<keyof ChartTypeRegistry>('bar');
  const [field, setField] = useState<IField | null>();
  const [fieldData, setFieldData] = useState<IProtoFormDataPoint[]>([]);

  const { data: formRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId });
  const [getFieldReport] = siteApi.useLazyFormServiceGetFormReportQuery();

  const reportFields = useMemo(() => {
    const versionForm: IFormTemplate = formRequest?.groupForm?.form?.version?.form;
    if (!versionForm) return [];

    const fields: IField[] = [];
    const allowedTypes = ['multi-select', 'single-select', 'boolean', 'number'];

    Object.values(versionForm).forEach(row => {
      row.forEach(field => {
        if (allowedTypes.includes(field.t)) {
          fields.push(field);
        }
      });
    });

    return fields;
  }, [formRequest]);

  const handleSubmit = () => {
    if (field) {
      getFieldReport({ formId, fieldId: field.i }).unwrap().then(res => {
        if (res.dataPoints) {
          setFieldData(res.dataPoints);
        }
      });
    }
  }

  return <>
    <Card variant="outlined" sx={{ mb: 2 }}>
      <CardHeader
        title={`Form Report: ${formRequest?.groupForm.form?.name}`}
        subheader="Select a field to view a graph of its collected data."
      />

      <CardContent>
        <Grid container alignItems="end" spacing={2}>
          <Grid>
            <TextField
              {...targets(`form report field selection`, `Select Field`, `select a field to view its visual report`)}
              sx={{ width: 300 }}
              select
              value={field?.i || ''}
              onChange={e => {
                setFieldData([]);
                const field = reportFields.find(f => f.i == e.target.value);
                if (field) {
                  setField(field);
                  const fc = fieldCharts[field.t];
                  if (fc?.length) {
                    setChartType(fc[0] as keyof ChartTypeRegistry);
                  }
                }
              }}
            >
              {!!field && <MenuItem key="unset-selection" value=""><Typography variant="caption">Remove selection</Typography></MenuItem>}
              {reportFields.map((f, i) => <MenuItem key={`report_field_sel${i}`} value={f.i}>{f.l}</MenuItem>)}
            </TextField>
          </Grid>
          {field && <Grid>
            <TextField
              {...targets(`form report field selection`, `Select Field`, `select a field to view its visual report`)}
              sx={{ width: 300 }}
              select
              value={chartType}
              onChange={e => setChartType(e.target.value as keyof ChartTypeRegistry)}
            >
              {fieldCharts[field.t]!.map((f, i) => <MenuItem key={`chart_type_sel${i}`} value={f}>{f}</MenuItem>)}
            </TextField>
          </Grid>}
        </Grid>
      </CardContent>

      {!!field && <CardActionArea
        {...targets(`form report generate`, `request a generated form using the selected field`)}
        onClick={() => handleSubmit()}
      >
        <Typography sx={{ m: 2, display: 'flex' }} color="secondary" variant="button">Generate</Typography>
      </CardActionArea>}
    </Card>

    {field && !!fieldData.length && <FormChart field={field} chartType={chartType} data={fieldData} />}
  </>
}

export default FormReport;
