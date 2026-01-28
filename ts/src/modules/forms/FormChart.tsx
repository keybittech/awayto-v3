import { useEffect, useRef } from 'react';

import Chart, { ChartTypeRegistry } from 'chart.js/auto';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';

import { IField, IProtoFormDataPoint } from 'awayto/hooks';

interface FormChartProps extends IComponent {
  field?: IField;
  chartType: keyof ChartTypeRegistry;
  data: IProtoFormDataPoint[];
}

export function FormChart({ field, chartType, data }: FormChartProps): React.JSX.Element {

  const labels = data ? data.map(d => d.label!) : [];
  const values = data ? data.map(d => Number(d.value!)) : [];

  if (!field || values.every(isNaN)) return <>No values submitted yet.</>;

  const chartRef = useRef<Chart | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!canvasRef.current || !data.length) return;

    if (chartRef.current) chartRef.current.destroy();

    chartRef.current = new Chart(canvasRef.current, {
      type: chartType,
      data: {
        labels,
        datasets: [{
          label: field.l,
          data: values,
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: chartType === 'pie' || chartType === 'doughnut' }
        },
        scales: {
          x: { display: chartType !== 'pie' && chartType !== 'doughnut' },
          y: {
            display: chartType !== 'pie' && chartType !== 'doughnut',
            beginAtZero: true
          }
        }
      }
    });

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
      }
    }

  }, [field, data, chartType]);

  return <>
    <Card variant="outlined">
      <CardHeader title={field.l} />

      <CardContent>
        <Box height={300}>
          <canvas ref={canvasRef} />
        </Box>
      </CardContent>

    </Card>
  </>
}

export default FormChart;
