import React, { useMemo } from 'react';

import Avatar from '@mui/material/Avatar';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';

import CheckIcon from '@mui/icons-material/Check';

import { DataGrid, GridColDef } from '@mui/x-data-grid';

import { useGrid, useStyles, IService } from 'awayto/hooks';
import Box from '@mui/material/Box';

declare global {
  interface IComponent {
    service?: IService;
    showFormChips?: boolean;
  }
}

export function ServiceTierAddons({ service, showFormChips }: IComponent): React.JSX.Element {

  const serviceTiers = useMemo(() => Object.values(service?.tiers || {}), [service?.tiers]);

  const classes = useStyles();

  const serviceTierData = useMemo(() => {
    const rows: { name: string, tiers: string[] }[] = [];
    if (serviceTiers) {
      serviceTiers.forEach(st => {
        for (const addon of Object.values(st.addons || {})) {
          const recordId = rows.findIndex(r => r.name === addon.name);
          const existing = recordId > -1 ? rows[recordId] : { id: `sta_cell_${st.id}_${addon.id}`, name: addon.name, tiers: [] } as { name: string, tiers: string[] };
          if (st.name) {
            existing.tiers.push(st.name);
          }
          if (recordId > -1) {
            rows[recordId] = existing;
          } else {
            rows.push(existing);
          }
        }
      });
    }
    return rows;
  }, [serviceTiers]);

  const tierGridProps = useGrid({
    noPagination: true,
    rows: serviceTierData,
    columnHeaderHeight: showFormChips ? 70 : undefined,
    columns: [
      {
        type: 'string',
        field: 'name',
        headerName: 'Features',
        flex: 1,
        sortable: false
      },
      ...serviceTiers.map<GridColDef<{ tiers: string[] }>>(st => {
        const hasFormOrSurvey = !!st.formId || !!st.surveyId;
        return ({
          type: 'string',
          field: `sta_col_${st.id}`,
          headerName: st.name,
          cellClassName: 'vertical-parent',
          renderHeader: col => {
            return !showFormChips ? col.colDef.headerName : <Box mt={-2}>
              <Typography mt={2}>{col.colDef.headerName}</Typography>
              {st.formId && <Chip color="info" size="small" label="Intake Form" />} &nbsp;
              {st.surveyId && <Chip color="warning" size="small" label="Survey Form" />}
              {!hasFormOrSurvey && <Chip size="small" label="No Forms" />}
            </Box>;
          },
          renderCell: params => {
            return st.name && params.row.tiers.includes(st.name) ?
              <Avatar sx={{ width: 24, height: 24, backgroundColor: 'white' }}>
                <CheckIcon sx={classes.green} />
              </Avatar> :
              '--'
          },
          flex: 1
        })
      })
    ]
  });

  return <DataGrid {...tierGridProps} />;
}

export default ServiceTierAddons;
