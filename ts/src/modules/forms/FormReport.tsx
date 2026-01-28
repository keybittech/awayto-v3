import { useEffect, useMemo, useState } from 'react';
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
  'single-select': ['bar', 'pie'],
  'multi-select': ['bar'],
  number: ['bar', 'pie', 'line'],
};

export function FormReport(_: IComponent): React.JSX.Element {

  const { formId } = useParams();
  if (!formId) return <></>;

  const [formVersionId, setFormVersionId] = useState('');
  const [chartType, setChartType] = useState<keyof ChartTypeRegistry>('bar');
  const [field, setField] = useState<IField | null>();
  const [fieldData, setFieldData] = useState<IProtoFormDataPoint[]>([]);

  const { data: formRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId });
  const [getVersionFieldReport] = siteApi.useLazyGroupFormServiceGetGroupFormVersionReportQuery();
  // const [getFormVersionData] = siteApi.useLazyFormServiceGetFormVersionDataQuery();
  {/* <Tooltip key={'download_data'} title="Download Data"> */ }
  {/*   <Button */ }
  {/*     {...targets(`manage forms download`, `download the data for the selected form`)} */ }
  {/*     color="info" */ }
  {/*     onClick={() => getFormVersionData({ formVersionId: selected[0] })} */ }
  {/*   > */ }
  {/*     <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Download</Typography> */ }
  {/*     <DownloadIcon sx={classes.variableButtonIcon} /> */ }
  {/*   </Button> */ }
  {/* </Tooltip>, */ }


  useEffect(() => {
    if (formRequest?.versionIds.length && !formVersionId) {
      setFormVersionId(formRequest.versionIds[0]);
    }
  }, [formRequest?.versionIds, formVersionId]);

  const { data: versionRequest } = siteApi.useGroupFormServiceGetGroupFormVersionByIdQuery({ formVersionId }, { skip: !formVersionId });

  const versionForm: IFormTemplate = versionRequest?.version.form;

  const reportFields = useMemo(() => {
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
  }, [versionForm]);

  useEffect(() => {
    if (reportFields.length) {
      setField(prev => {
        const exists = reportFields.find(f => f.i === prev?.i);
        return exists || reportFields[0];
      });
    }
  }, [reportFields]);

  const handleSubmit = () => {
    if (field) {
      getVersionFieldReport({ formVersionId, fieldId: field.i }).unwrap().then(res => {
        if (res.dataPoints) {
          setFieldData(res.dataPoints);
        }
      });
    }
  }

  if (!formRequest?.versionIds.length) return <></>;

  return <>
    <Card variant="outlined" sx={{ mb: 2 }}>
      <CardHeader
        title={`Form Report: ${formRequest?.groupForm.form?.name}`}
        subheader="Select a field to view a graph of its collected data. Data is refreshed after 3 minutes."
      />

      <CardContent>
        <Grid container alignItems="end" spacing={2}>
          <Grid>
            <TextField
              {...targets(`form report version selection`, `Version`, `select a form version for the report`)}
              sx={{ width: 300 }}
              select
              value={formVersionId}
              onChange={e => {
                setFormVersionId(e.target.value);
                setFieldData([]);
                setChartType('bar');
              }}
            >
              {formRequest.versionIds.map((vid, i) => <MenuItem key={`report_field_sel${i}`} value={vid}>Version {formRequest.versionIds.length - i}</MenuItem>)}
            </TextField>
          </Grid>
          <Grid>
            <TextField
              {...targets(`form report field selection`, `Field`, `select a field for the report`)}
              sx={{ width: 300 }}
              select
              value={field?.i || ''}
              onChange={e => {
                setFieldData([]);
                const field = reportFields.find(f => f.i == e.target.value);
                if (field) {
                  setField(field);
                  setChartType('bar');
                }
              }}
            >
              {reportFields.map((f, i) => <MenuItem key={`report_field_sel${i}`} value={f.i}>{f.l} ({f.t})</MenuItem>)}
            </TextField>
          </Grid>
          {field && <Grid>
            <TextField
              {...targets(`form report chart type selection`, `Chart`, `select a chart type to render the report`)}
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

      <CardActionArea
        {...targets(`form report generate`, `request a generated form using the selected field`)}
        onClick={() => handleSubmit()}
      >
        <Typography sx={{ m: 2, display: 'flex' }} color="secondary" variant="button">Generate</Typography>
      </CardActionArea>
    </Card>

    {field && !!fieldData.length && <FormChart field={field} chartType={chartType} data={fieldData} />}
  </>
}

export default FormReport;
