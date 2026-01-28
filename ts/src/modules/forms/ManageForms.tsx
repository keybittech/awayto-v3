import React, { useState, useMemo, Suspense } from 'react';
import { useNavigate } from 'react-router';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';

import CreateIcon from '@mui/icons-material/Create';
import NoteAddIcon from '@mui/icons-material/NoteAdd';
import DeleteIcon from '@mui/icons-material/Delete';
import AssessmentIcon from '@mui/icons-material/Assessment';

import { DataGrid } from '@mui/x-data-grid';

import { siteApi, useGrid, useStyles, dayjs, IGroupForm, IForm, targets } from 'awayto/hooks';

import ManageFormModal from './ManageFormModal';

export function ManageForms(props: IComponent): React.JSX.Element {
  const navigate = useNavigate();
  const classes = useStyles();

  const [deleteGroupForm] = siteApi.useGroupFormServiceDeleteGroupFormMutation();
  const { data: groupFormsRequest, refetch: getGroupForms } = siteApi.useGroupFormServiceGetGroupFormsQuery();


  const [groupForm, setGroupForm] = useState<IGroupForm>();
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_form'} title="Edit">
        <Button
          {...targets(`manage forms edit`, `edit the selected form`)}
          color="info"
          onClick={() => {
            setGroupForm(groupFormsRequest?.groupForms?.find(gf => gf.formId === selected[0]));
            setDialog('manage_form');
            setSelected([]);
          }}
        >
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>,
      <Tooltip key={'view_reports'} title="Report">
        <Button
          {...targets(`manage forms report`, `view reports for the selected form`)}
          color="info"
          onClick={() => navigate(`/group/manage/forms/${selected[0]}/report`)}
        >
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Report</Typography>
          <AssessmentIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>,
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_group'} title="Delete">
        <Button
          {...targets(`manage forms delete`, `delete the selected form or forms`)}
          color="error"
          onClick={() => {
            if (selected.length) {
              deleteGroupForm({ ids: selected.join(',') }).unwrap().then(() => {
                setSelected([]);
                void getGroupForms();
              }).catch(console.error);
            }
          }}>
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ]
  }, [selected]);

  const formGridProps = useGrid<IGroupForm>({
    rowId: 'formId',
    rows: groupFormsRequest?.groupForms || [],
    columns: [
      { flex: 1, headerName: 'Name', field: 'name', renderCell: ({ row }) => row.form?.name },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.form?.createdOn)) },
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      <Typography variant="button">Forms:</Typography>
      <Tooltip key={'manage_form'} title="Create">
        <Button
          {...targets(`manage forms create`, `create a new form`)}
          color="info"
          onClick={() => {
            setGroupForm(undefined);
            setDialog('manage_form')
          }}
        >
          <Typography variant="button" sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
          <NoteAddIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
      {!!selected.length && <Box sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Box>}
    </>
  });

  return <>
    <Dialog fullScreen onClose={setDialog} open={dialog === 'manage_form'} fullWidth maxWidth="sm">
      <Suspense>
        <ManageFormModal {...props} editForm={groupForm?.form as IForm} closeModal={() => {
          setDialog('')
          void getGroupForms();
        }} />
      </Suspense>
    </Dialog>

    <DataGrid {...formGridProps} />
  </>
}

export default ManageForms;
