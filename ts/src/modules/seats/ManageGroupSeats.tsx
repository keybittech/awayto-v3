import React from 'react';

import { DataGrid } from '@mui/x-data-grid';

import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import PrintIcon from '@mui/icons-material/Print';

import { useGrid, siteApi, dayjs, utcDTLocal } from 'awayto/hooks';

export function ManageSeats(_: IComponent): React.JSX.Element {
  const { data: seatRequest } = siteApi.useGroupSeatServiceGetGroupSeatPaymentsQuery();
  const [getPo] = siteApi.useLazyGroupSeatServiceGetGroupSeatPurchaseOrderQuery();
  const [getReceipt] = siteApi.useLazyGroupSeatServiceGetGroupSeatReceiptQuery();

  const handlePrint = async (code: string, kind: string) => {
    const printWindow = window.open('', '_blank', 'width=800,height=800');

    if (!printWindow) {
      alert('Please allow popups to view the document.');
      return;
    }

    if (printWindow.document.body) {
      printWindow.document.body.innerHTML = '<h5>Generating document...</h5>';
    }

    try {
      const { data } = await (kind == 'po' ? getPo : getReceipt)({ code });

      if (data?.html) {
        if (printWindow.document.documentElement) {
          printWindow.document.documentElement.innerHTML = data.html;
        }
        const btn = printWindow.document.getElementById('print-btn');
        if (btn) {
          btn.onclick = () => printWindow.print();
        }
      } else {
        if (printWindow.document.body) {
          printWindow.document.body.innerHTML = '<h3>Error: Could not retrieve document.</h3>';
        }
      }
    } catch (e) {
      console.error(e);
      if (printWindow.document.body) {
        printWindow.document.body.innerHTML = '<h3>Error loading document.</h3>';
      }
    }
  }


  const seatGridProps = useGrid({
    rowId: 'code',
    rows: seatRequest?.seatPayments || [],
    initialSort: { field: 'createdOn', sort: 'desc' },
    columns: [
      { flex: 1, headerName: 'Code', field: 'code' },
      { flex: 1, headerName: 'Seats', field: 'seats' },
      { flex: 1, headerName: 'Amount ($)', field: 'amount' },
      { flex: 1, headerName: 'Status', field: 'status' },
      {
        flex: 1, field: 'po', headerName: 'P.O.', sortable: false,
        renderCell: ({ row }) => (
          <Tooltip placement="left" title="Generate Purchase Order">
            <IconButton onClick={() => row.code && handlePrint(row.code, 'po')} size="small">
              <PrintIcon />
            </IconButton>
          </Tooltip>
        )
      },
      {
        flex: 1, field: 'receipt', headerName: 'Receipt', sortable: false,
        renderCell: ({ row }) => row.status == 'paid' ? (
          <Tooltip placement="left" title="Generate Receipt">
            <IconButton onClick={() => row.code && handlePrint(row.code, 'receipt')} size="small">
              <PrintIcon />
            </IconButton>
          </Tooltip>
        ) : <></>
      },
      { flex: 1, headerName: 'Created', field: 'createdOn', valueFormatter: (value) => dayjs(value).format('YYYY-MM-DD') },
      { flex: 1, headerName: 'Due Date', field: 'dueDate', valueFormatter: (value) => dayjs(value).add(30, 'day').format('YYYY-MM-DD') },
      { flex: 1, headerName: 'Paid On', field: 'paidOn', valueFormatter: (value) => value ? dayjs(value).format('YYYY-MM-DD') : 'unpaid' },
    ]
  });

  return <DataGrid {...seatGridProps} />
}

export default ManageSeats;
