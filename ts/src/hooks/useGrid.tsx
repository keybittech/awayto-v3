import React, { useMemo } from 'react';
import { GridColDef, GridRowSelectionModel, GridValidRowModel } from '@mui/x-data-grid';

import Grid from '@mui/material/Grid';

type UseScheduleProps<T extends GridValidRowModel> = {
  rows: T[];
  columns: GridColDef<T>[];
  columnHeaderHeight?: number;
  rowId?: string;
  noPagination?: boolean;
  selected?: GridRowSelectionModel;
  disableRowSelectionOnClick?: boolean;
  onSelected?: (value: GridRowSelectionModel) => void;
  toolbar?: () => React.JSX.Element;
};

export function useGrid<T extends GridValidRowModel>({ rows, columns, columnHeaderHeight, rowId, noPagination, selected, onSelected, toolbar, disableRowSelectionOnClick = true }: UseScheduleProps<T>) {
  const defaultHeight = 42;
  return {
    autoHeight: true,
    sx: {
      borderColor: 'primary.light',
      '& [role="row"]': {
        bgcolor: 'inherit !important'
      }
    },
    rows,
    columns,
    columnHeaderHeight: columnHeaderHeight || defaultHeight,
    rowSelectionModel: selected,
    checkboxSelection: !!selected,
    onRowSelectionModelChange: onSelected,
    hideFooterPagination: noPagination,
    pageSizeOptions: noPagination ? [] : [5, 10, 25, 50, 100],
    disableRowSelectionOnClick,
    getRowId: (row: T) => (rowId ? row[rowId] : row.id) as string,
    slots: { toolbar: () => toolbar ? <Grid container p={2} alignItems="center">{toolbar()}</Grid> : <></> }
  }
}
