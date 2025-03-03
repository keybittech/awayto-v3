import React, { useState, useMemo, useContext } from 'react';

import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

import CreateIcon from '@mui/icons-material/Create';
import DeleteIcon from '@mui/icons-material/Delete';

import { DataGrid } from '@mui/x-data-grid';

import { useGrid, useUtil, siteApi, dayjs, plural, ISchedule, targets } from 'awayto/hooks';

import GroupContext, { GroupContextType } from '../groups/GroupContext';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import ManageScheduleBracketsModal from './ManageScheduleBracketsModal';

// This is how group users interact with the schedule

export function ManageScheduleBrackets(_: IComponent): React.JSX.Element {

  const { setSnack, openConfirm } = useUtil();

  const {
    // GroupSelect,
    groupSchedules,
  } = useContext(GroupContext) as GroupContextType;

  const {
    getGroupSchedules: {
      refetch: refetchGroupSchedules
    },
    getGroupUserSchedules: {
      refetch: refetchGroupUserSchedules
    },
    getGroupUserScheduleStubs: {
      refetch: refetchGroupUserScheduleStubs
    },
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;


  const [deleteGroupUserScheduleByUserScheduleId] = siteApi.useGroupUserScheduleServiceDeleteGroupUserScheduleByUserScheduleIdMutation();
  const [deleteSchedule] = siteApi.useScheduleServiceDeleteScheduleMutation();

  const [schedule, setSchedule] = useState<ISchedule>();
  const [selected, setSelected] = useState<string[]>([]);
  const [dialog, setDialog] = useState('');

  const { data: schedulesRequest, refetch: getSchedules } = siteApi.useScheduleServiceGetSchedulesQuery();

  const [getScheduleById, { isFetching }] = siteApi.useLazyScheduleServiceGetScheduleByIdQuery();

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_schedule'} title="Edit">
        <IconButton
          {...targets(`manage schedule brackets edit`, `edit the currently selected schedule`)}
          key={'manage_schedule'}
          onClick={() => {
            const sched = schedulesRequest?.schedules.find(sc => sc.id == selected[0])
            if (sched?.id && !isFetching) {
              getScheduleById({ id: sched.id }).unwrap().then(({ schedule: sbid }) => {
                setSchedule(sbid);
                setDialog('manage_schedule');
                setSelected([]);
              }).catch(console.error);
            }
          }}
        >
          <CreateIcon />
        </IconButton>
      </Tooltip>
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_schedule'} title="Delete">
        <IconButton
          {...targets(`manage schedule brackets delete`, `delete the currently selected schedule or schedules`)}
          onClick={() => {
            openConfirm({
              isConfirming: true,
              confirmEffect: `Remove ${plural(selected.length, 'schedule', 'schedules')}. This cannot be undone.`,
              confirmAction: async () => {
                const ids = selected.join(',');
                await deleteGroupUserScheduleByUserScheduleId({ ids }).unwrap();
                await deleteSchedule({ ids }).unwrap();

                void getSchedules();
                void refetchGroupSchedules().then(() => {
                  void refetchGroupUserSchedules();
                  void refetchGroupUserScheduleStubs();
                });

                setSnack({ snackType: 'success', snackOn: 'Successfully removed schedule records.' });
              }
            });
          }}
        >
          <DeleteIcon />
        </IconButton>
      </Tooltip>
    ]
  }, [selected, schedulesRequest?.schedules, isFetching]);

  const scheduleBracketGridProps = useGrid({
    rows: schedulesRequest?.schedules || [],
    columns: [
      { flex: 1, headerName: 'Name', field: 'name' },
      { flex: 1, headerName: 'Created', field: 'createdOn', renderCell: ({ row }) => dayjs().to(dayjs.utc(row.createdOn)) }
    ],
    selected,
    onSelected: selection => setSelected(selection as string[]),
    toolbar: () => <>
      {/* <GroupSelect /> */}
      <Grid container size="grow" alignItems="center">
        <Typography sx={{ textTransform: 'uppercase' }}>Schedules:</Typography>
        <Grid size="grow">
          <Button
            {...targets(`manage schedule brackets create`, `create a new personal schedule`)}
            color="info"
            onClick={() => {
              if (groupSchedules?.length) {
                setSchedule(undefined);
                setDialog('manage_schedule');
              } else {
                setSnack({ snackType: 'warning', snackOn: 'There are no available master schedules.' })
              }
            }}
          >Create</Button>
        </Grid>
        {!!selected.length && <Grid sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Grid>}
      </Grid>
    </>
  })

  return <>
    <Dialog fullScreen open={dialog === 'manage_schedule'}>
      <ManageScheduleBracketsModal
        editSchedule={schedule}
        closeModal={(shouldReload: boolean) => {
          setDialog('');
          if (shouldReload) {
            if (schedule?.id) {
              void getScheduleById({ id: schedule.id }).catch(console.error);
            }
            void getSchedules().catch(console.error);
          }
        }}
      />
    </Dialog>

    <DataGrid {...scheduleBracketGridProps} />
  </>
}

export default ManageScheduleBrackets;
