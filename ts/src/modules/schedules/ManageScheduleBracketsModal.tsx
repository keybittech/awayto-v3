import React, { useMemo, useState, useRef, useCallback, Suspense, useContext } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Slider from '@mui/material/Slider';
import Switch from '@mui/material/Switch';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import DialogActions from '@mui/material/DialogActions';

import { siteApi, useUtil, getRelativeDuration, ISchedule, IService, IScheduleBracket, timeUnitOrder, useTimeName, targets } from 'awayto/hooks';

import GroupContext, { GroupContextType } from '../groups/GroupContext';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import ScheduleDisplay from './ScheduleDisplay';

const bracketSchema = {
  duration: 1,
  automatic: false,
  multiplier: 100
};

interface ManageScheduleBracketsModalProps extends IComponent {
  editSchedule: ISchedule;
}

export function ManageScheduleBracketsModal({ editSchedule, closeModal }: ManageScheduleBracketsModalProps): React.JSX.Element {

  const { setSnack } = useUtil();

  const {
    groupServices,
    groupSchedules
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
    selectGroupSchedule: {
      item: groupSchedule,
      comp: GroupScheduleSelect
    }
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [schedule, setSchedule] = useState(editSchedule);

  if (groupSchedule?.schedule && !editSchedule && (!schedule || groupSchedule.schedule.id !== schedule.id)) {
    setSchedule({ ...groupSchedule?.schedule, brackets: {} });
  }

  const scheduleTimeUnitName = useTimeName(schedule.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(schedule.bracketTimeUnitId);

  const firstLoad = useRef(true);
  const [viewStep, setViewStep] = useState(1);
  if (firstLoad.current && viewStep === 1 && Object.keys(schedule.brackets || {}).length) {
    setViewStep(2);
  }

  firstLoad.current = false;

  const [postSchedule] = siteApi.useScheduleServicePostScheduleMutation();
  const [postScheduleBrackets] = siteApi.useScheduleServicePostScheduleBracketsMutation();
  const [postGroupUserSchedule] = siteApi.useGroupUserScheduleServicePostGroupUserScheduleMutation();

  // const scheduleParent = useRef<HTMLDivElement>(null);
  const [bracket, setBracket] = useState({ ...bracketSchema, services: {}, slots: {} } as Required<IScheduleBracket>);

  const scheduleBracketsValues = useMemo(() => Object.values(schedule.brackets || {}) as Required<IScheduleBracket>[], [schedule.brackets]);
  const bracketServicesValues = useMemo(() => Object.values(bracket.services || {}) as Required<IService>[], [bracket.services]);

  const remainingBracketTime = useMemo(() => {
    if (scheduleTimeUnitName && bracketTimeUnitName) {
      // Complex time adjustment example - this gets the remaining time that can be scheduled based on the schedule context, which always selects its first child as the subcontext (Week > Day), multiply this by the schedule duration in that context (1 week is 7 days), then convert the result to whatever the bracket type is. So if the schedule is for 40 hours per week, the schedule duration is 1 week, which is 7 days. The bracket, in hours, gives 24 hrs per day * 7 days, resulting in 168 total hours. Finally, subtract the time used by selected slots in the schedule display.
      const scheduleUnitChildUnit = timeUnitOrder[timeUnitOrder.indexOf(scheduleTimeUnitName) - 1];
      const scheduleChildDuration = getRelativeDuration(1, scheduleTimeUnitName, scheduleUnitChildUnit); // 7
      const usedDuration = scheduleBracketsValues.reduce((m, d) => m + (d.duration || 0), 0);
      const totalDuration = getRelativeDuration(Math.floor(scheduleChildDuration), scheduleUnitChildUnit, bracketTimeUnitName);
      return Math.floor(totalDuration - usedDuration);
    }
    return 0;
  }, [scheduleTimeUnitName, bracketTimeUnitName]);

  const handleSubmit = useCallback(() => {
    async function go() {
      if (groupSchedule && schedule.name && scheduleBracketsValues.length) {

        const groupScheduleId = groupSchedule.schedule?.id;
        let userScheduleId = editSchedule?.id;

        if (!userScheduleId) {
          const newSchedule = await postSchedule({ postScheduleRequest: { schedule } }).unwrap().catch(console.error);
          userScheduleId = newSchedule?.id;
        }

        if (!userScheduleId || !groupScheduleId) return;

        const newBrackets = scheduleBracketsValues.reduce<Record<string, IScheduleBracket>>(
          (m, { id, duration, automatic, multiplier, slots, services }) => ({
            ...m,
            [id]: {
              id,
              duration,
              automatic,
              multiplier,
              slots,
              services
            } as IScheduleBracket
          }), {}
        );

        await postScheduleBrackets({
          postScheduleBracketsRequest: {
            scheduleId: userScheduleId,
            brackets: newBrackets
          }
        }).unwrap().catch(console.error);

        await postGroupUserSchedule({
          postGroupUserScheduleRequest: {
            userScheduleId: userScheduleId, // refers to the user's schedule record
            groupScheduleId: groupScheduleId // refers to the master schedule record
          }
        }).unwrap().catch(console.error);

        void refetchGroupSchedules().then(() => {
          void refetchGroupUserSchedules();
          void refetchGroupUserScheduleStubs();
        });

        setSnack({ snackOn: 'Successfully added your schedule to group schedule: ' + schedule.name, snackType: 'info' });
        if (closeModal) {
          firstLoad.current = true;
          closeModal(true);
        }
      } else {
        setSnack({ snackOn: 'A schedule should have a name, a duration, and at least 1 bracket.', snackType: 'info' });
      }
    }
    void go();
  }, [schedule, groupSchedule, scheduleBracketsValues]);

  return <>
    <DialogTitle>{!editSchedule?.id ? 'Create' : 'Manage'} Schedule Bracket</DialogTitle>
    <DialogContent>

      {1 === viewStep ? <>
        <Box mt={2} />

        <Box mb={4}>
          <GroupScheduleSelect>
            {groupSchedules.map(s => {
              return <MenuItem
                key={`schedule-select${s.id}`}
                value={s.id}
                sx={{ alignItems: 'baseline' }}
              >
                {s.schedule?.name}&nbsp;&nbsp;&nbsp;
                <Typography variant="caption" fontSize={10}>Timezone: {s.schedule?.timezone}</Typography>
              </MenuItem>
            })}
          </GroupScheduleSelect>
        </Box>

        <Box mb={4}>
          <Typography variant="body1"></Typography>
          <TextField
            {...targets(`manage schedule brackets modal remaining time`, `# of ${bracketTimeUnitName}s`, `set the number of ${bracketTimeUnitName}s which should available to schedule`)}
            fullWidth
            type="number"
            helperText={`Number of ${bracketTimeUnitName}s for this schedule. (Remaining: ${remainingBracketTime})`}
            value={bracket.duration || ''}
            onChange={e => setBracket({ ...bracket, duration: Math.min(Math.max(0, parseInt(e.target.value || '', 10)), remainingBracketTime) })}
          />
        </Box>

        <Box sx={{ display: 'none' }}>
          <Typography variant="h6">Multiplier</Typography>
          <Typography variant="body2">Affects the cost of all services in this bracket.</Typography>
          <Box sx={{ display: 'flex', alignItems: 'baseline' }}>
            <Box>{(bracket.multiplier || 100) / 100}x <span>&nbsp;</span> &nbsp;</Box>
            <Slider
              {...targets(`manage schedule brackets modal multiplier`, `set a multiplier to be applied to the cost of the selected services`)}
              value={bracket.multiplier || 100}
              onChange={(_, val) => setBracket({ ...bracket, multiplier: val as number })}
              step={1}
              min={0}
              max={500}
            />
          </Box>
        </Box>

        <Box sx={{ display: 'none' }}>
          <Typography variant="h6">Automatic</Typography>
          <Typography variant="body2">Bracket will automatically accept new bookings.</Typography>
          <Switch color="primary" value={bracket.automatic} onChange={e => setBracket({ ...bracket, automatic: e.target.checked })} />
        </Box>

        {groupServices && <Box mb={4}>
          <TextField
            {...targets(`manage schedule brackets modal services selection`, `Services`, `select the services to be available on the schedule`)}
            select
            fullWidth
            helperText="Select the services available to be scheduled."
            value={''}
            onChange={e => {
              const serv = groupServices?.find(gs => gs.service?.id === e.target.value)?.service;
              if (serv && bracket.services) {
                bracket.services[e.target.value] = serv;
                setBracket({ ...bracket, services: { ...bracket.services } });
              }
            }}
          >
            {groupServices.filter(s => s.service?.id && !Object.keys(bracket.services).includes(s.service?.id)).map(gs =>
              <MenuItem key={`service-select${gs.service?.id}`} value={gs.service?.id}>{gs.service?.name}</MenuItem>
            )}
          </TextField>

          <Box sx={{ display: 'flex', alignItems: 'flex-end', flexWrap: 'wrap' }}>
            {bracketServicesValues.map((service, i) => {
              return <Box key={`service-chip${i + 1}new`} m={1}>
                <Chip
                  {...targets(`manage schedule brackets modal delete service ${i}`, `${service.name} ${service.cost ? `Cost: ${service.cost}` : ''}`, `remove ${service} from the schedule's services`)}
                  onDelete={() => {
                    delete bracket.services[service.id];
                    setBracket({ ...bracket, services: { ...bracket.services } });
                  }}
                />
              </Box>
            })}
          </Box>
        </Box>}
      </> : <>
        <Suspense fallback={<CircularProgress />}>
          <ScheduleDisplay schedule={schedule} setSchedule={setSchedule} />
        </Suspense>
      </>}
    </DialogContent>
    <DialogActions>
      <Grid container justifyContent="space-between">
        <Button
          {...targets(`manage personal schedule modal close`, `close the schedule editing modal`)}
          onClick={() => {
            if (closeModal) {
              firstLoad.current = true;
              closeModal();
            }
          }}
        >Cancel</Button>
        {1 === viewStep ? <Grid>
          {!!scheduleBracketsValues.length && <Button
            {...targets(`manage personal schedule modal cancel addition`, `continue to the next page of personal schedule mangement without adding another bracket`)}
            onClick={() => {
              setViewStep(2);
              setBracket({ ...bracketSchema, services: {}, slots: {} } as Required<IScheduleBracket>);
            }}
          >
            Cancel Add
          </Button>}
          <Button
            {...targets(`manage personal schedule modal next`, `continue to the next page of personal schedule management`)}
            disabled={!bracket.duration || !bracketServicesValues.length}
            onClick={() => {
              if (schedule.id && bracket.duration && Object.keys(bracket.services).length) {
                bracket.id = (new Date()).getTime().toString();
                bracket.scheduleId = schedule.id;
                const newBrackets = { ...schedule.brackets };
                newBrackets[bracket.id] = bracket;
                setSchedule({ ...schedule, brackets: newBrackets })
                setBracket({ ...bracketSchema, services: {}, slots: {} } as Required<IScheduleBracket>);
                setViewStep(2);
              } else {
                void setSnack({ snackOn: 'Provide a duration, and at least 1 service.', snackType: 'info' });
              }
            }}
          >
            Continue
          </Button>
        </Grid> : <Grid>
          <Button
            {...targets(`manage personal schedule modal previous`, `go to the first page of personal schedule management`)}
            onClick={() => { setViewStep(1); }}
          >Add bracket</Button>
          <Button
            {...targets(`manage personal schedule modal submit`, `submit the current personal schedule for editing or creation`)}
            onClick={handleSubmit}
          >Submit</Button>
        </Grid>}
      </Grid>
    </DialogActions>
  </>
}

export default ManageScheduleBracketsModal;
