import React, { useState, useMemo, useContext, useEffect } from 'react';

import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';

import CreateIcon from '@mui/icons-material/Create';
import DeleteIcon from '@mui/icons-material/Delete';
import MoreTimeIcon from '@mui/icons-material/MoreTime';

import { DataGrid } from '@mui/x-data-grid';

import { useGrid, useUtil, siteApi, dayjs, plural, ISchedule, targets, useStyles } from 'awayto/hooks';

import GroupContext, { GroupContextType } from '../groups/GroupContext';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import ManageScheduleBracketsModal from './ManageScheduleBracketsModal';

// This is how group users interact with the schedule

export function ManageScheduleBrackets(_: IComponent): React.JSX.Element {

  const classes = useStyles();
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

  const { data: schedulesRequest, refetch: getSchedules } = siteApi.useScheduleServiceGetSchedulesQuery(undefined, { refetchOnMountOrArgChange: 30 });

  const [getScheduleById, { isFetching }] = siteApi.useLazyScheduleServiceGetScheduleByIdQuery();

  const userScheduleNames = useMemo(() => schedulesRequest?.schedules?.map(s => s.name), [schedulesRequest?.schedules]);
  const unusedGroupSchedules = useMemo(() => groupSchedules.filter(x => !userScheduleNames?.includes(x.name)), [groupSchedules, userScheduleNames]);

  const actions = useMemo(() => {
    const { length } = selected;
    const acts = length == 1 ? [
      <Tooltip key={'manage_schedule'} title="Edit">
        <Button
          {...targets(`manage schedule brackets edit`, `edit the currently selected schedule`)}
          key={'manage_schedule'}
          color="info"
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
          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Edit</Typography>
          <CreateIcon sx={classes.variableButtonIcon} />
        </Button>
      </Tooltip>
    ] : [];

    return [
      ...acts,
      <Tooltip key={'delete_schedule'} title="Delete">
        <Button
          {...targets(`manage schedule brackets delete`, `delete the currently selected schedule or schedules`)}
          color="error"
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
          <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Delete</Typography>
          <DeleteIcon sx={classes.variableButtonIcon} />
        </Button>
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
          <Tooltip key={'create_schedule'} title="Create">
            <Button
              {...targets(`manage schedule brackets create`, `create a new personal schedule`)}
              color="info"
              onClick={() => {
                if (unusedGroupSchedules?.length) {
                  setSchedule(undefined);
                  setDialog('manage_schedule');
                } else {
                  setSnack({ snackType: 'info', snackOn: 'There are no available group schedules to create a new personal schedule. To edit an existing personal schedule, click the checkbox below, then click the pencil icon.' })
                }
              }}
            >
              <Typography sx={{ display: { xs: 'none', md: 'flex' } }}>Create</Typography>
              <MoreTimeIcon sx={classes.variableButtonIcon} />
            </Button>
          </Tooltip>
        </Grid>
        {!!selected.length && <Grid sx={{ flexGrow: 1, textAlign: 'right' }}>{actions}</Grid>}
      </Grid>
    </>
  })

  return <>
    <Dialog fullScreen onClose={setDialog} open={dialog === 'manage_schedule'}>
      <ManageScheduleBracketsModal
        groupSchedules={unusedGroupSchedules}
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
