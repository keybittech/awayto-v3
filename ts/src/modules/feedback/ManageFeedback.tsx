import React from 'react';

import { DataGrid } from '@mui/x-data-grid';

import { useGrid, siteApi, dayjs } from 'awayto/hooks';

export function ManageFeedbacks(): React.JSX.Element {
  const { data: feedbackRequest } = siteApi.useGroupFeedbackServiceGetGroupFeedbackQuery();

  const feedbackGridProps = useGrid({
    rows: feedbackRequest?.feedback || [],
    columns: [
      { flex: 1, headerName: 'Message', field: 'feedbackMessage' },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.createdOn)) }
    ]
  });

  return <DataGrid {...feedbackGridProps} />
}

export default ManageFeedbacks;
