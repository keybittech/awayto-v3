import React, { useMemo, useState, useRef, useCallback, useContext, useEffect } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Slider from '@mui/material/Slider';
import Switch from '@mui/material/Switch';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import DialogActions from '@mui/material/DialogActions';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';

import { siteApi, useUtil, getRelativeDuration, ISchedule, IService, IScheduleBracket, timeUnitOrder, useTimeName, targets, plural, generateLightBgColor, IGroupSchedule } from 'awayto/hooks';

import GroupContext, { GroupContextType } from '../groups/GroupContext';
import GroupScheduleContext, { GroupScheduleContextType } from '../group_schedules/GroupScheduleContext';
import ScheduleDisplay from './ScheduleDisplay';

const bracketSchema = {
  duration: 1,
  automatic: false,
  multiplier: 100
};

interface ManageScheduleBracketsModalProps extends IComponent {
  groupSchedules: IGroupSchedule[];
  editSchedule?: ISchedule;
}

export function ManageScheduleBracketsModal({ editSchedule, groupSchedules, closeModal }: ManageScheduleBracketsModalProps): React.JSX.Element {

  const { setSnack } = useUtil();

  const {
    groupServices,
  } = useContext(GroupContext) as GroupContextType;

  const {
    getGroupUserSchedules,
    selectGroupSchedule: {
      item: groupSchedule,
      comp: GroupScheduleSelect
    }
  } = useContext(GroupScheduleContext) as GroupScheduleContextType;

  const [schedule, setSchedule] = useState(editSchedule);

  const scheduleTimeUnitName = useTimeName(schedule?.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(schedule?.bracketTimeUnitId);
  const slotTimeUnitName = useTimeName(schedule?.slotTimeUnitId);

  const firstLoad = useRef(true);
  const [viewStep, setViewStep] = useState(1);

  const [postSchedule] = siteApi.useScheduleServicePostScheduleMutation();
  const [postScheduleBrackets] = siteApi.useScheduleServicePostScheduleBracketsMutation();

  const [bracket, setBracket] = useState({ ...bracketSchema, services: {}, slots: {} } as IScheduleBracket);

  const scheduleBracketsValues = useMemo(() => Object.values(schedule?.brackets || {}) as IScheduleBracket[], [schedule?.brackets]);
  const bracketServicesValues = useMemo(() => Object.values(bracket.services || {}) as IService[], [bracket.services]);

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
  }, [scheduleTimeUnitName, bracketTimeUnitName, scheduleBracketsValues]);

  const handleSubmit = useCallback(() => {
    async function go() {
      if (!groupSchedule?.id || !schedule || !scheduleBracketsValues.length) {
        setSnack({ snackOn: 'A schedule should have a name, a duration, and at least 1 bracket.', snackType: 'info' });
      } else {
        const newBrackets = scheduleBracketsValues.reduce<Record<string, IScheduleBracket>>(
          (m, { id, duration, automatic, multiplier, slots, services }) => !id ? m : ({
            ...m,
            [id]: {
              id,
              duration,
              automatic,
              multiplier,
              slots,
              services,
            }
          }), {}
        );

        let success = false;
        if (!editSchedule?.id && groupSchedule.schedule?.id) {
          await postSchedule({
            postScheduleRequest: {
              groupScheduleId: groupSchedule.schedule.id,
              brackets: newBrackets,
              name: schedule.name,
              startTime: schedule.startTime,
              endTime: schedule.endTime,
              scheduleTimeUnitId: schedule.scheduleTimeUnitId,
              bracketTimeUnitId: schedule.bracketTimeUnitId,
              slotTimeUnitId: schedule.slotTimeUnitId,
              slotDuration: schedule.slotDuration
            }
          }).unwrap().then(() => success = true).catch(console.error);
        } else if (editSchedule?.id) {
          await postScheduleBrackets({
            postScheduleBracketsRequest: {
              groupScheduleId: groupSchedule.id,
              userScheduleId: editSchedule.id,
              brackets: newBrackets
            }
          }).unwrap().then(() => success = true).catch(console.error);
        }

        if (success) {
          await getGroupUserSchedules.refetch().unwrap();
          setSnack({ snackOn: 'Successfully added your schedule to group schedule: ' + schedule?.name, snackType: 'info' });
        }
        if (closeModal) {
          firstLoad.current = true;
          closeModal(true);
        }
      }
    }
    void go();
  }, [schedule, editSchedule, groupSchedule, scheduleBracketsValues]);

  const changeColors = () => {
    if (schedule?.brackets) {
      const brackets: typeof schedule.brackets = {};
      Object.values(schedule.brackets).forEach(b => {
        if (!b.id) return;
        brackets[b.id] = {
          ...b,
          color: generateLightBgColor()
        };
      });
      setSchedule({ ...schedule, brackets });
    }
  }

  // reset on group schedule change or init
  if (groupSchedule?.schedule && !editSchedule?.id && (!schedule?.id || groupSchedule.schedule.id !== schedule?.id)) {
    setSchedule({ ...groupSchedule?.schedule, brackets: {} });
    setBracket({ ...bracketSchema, services: {}, slots: {} });
  }

  if (firstLoad.current && viewStep === 1 && schedule?.brackets) {
    setViewStep(2);
    changeColors();
  }
  firstLoad.current = false;

  useEffect(() => {
    if (!scheduleBracketsValues.length) {
      setViewStep(1);
    }
  }, [scheduleBracketsValues]);

  return <>
    <DialogTitle>{!editSchedule?.id ? 'Create' : 'Manage'} Schedule Bracket</DialogTitle>
    <DialogContent>

      {1 === viewStep ? <>
        <Box mt={2} />

        {!Boolean(editSchedule?.id) && <Box mb={4}>
          <GroupScheduleSelect
            disabled={Boolean(editSchedule?.id)}
            helperText={
              schedule?.slotDuration && slotTimeUnitName && `This schedule represents 1 ${scheduleTimeUnitName} of ${bracketTimeUnitName}s where every ${plural(schedule?.slotDuration, slotTimeUnitName, slotTimeUnitName + 's')} is schedulable.`
            }
          >
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
        </Box>}

        <Box mb={4}>
          <TextField
            {...targets(`manage schedule brackets modal remaining time`, `# of ${bracketTimeUnitName}s`, `set the number of ${bracketTimeUnitName}s which should available to schedule`)}
            fullWidth
            type="number"
            helperText={`Number of ${bracketTimeUnitName}s for this schedule. (Remaining: ${remainingBracketTime})`}
            value={bracket.duration || ''}
            onChange={e => {
              const numVal = parseInt(e.target.value || '0', 10);
              if (numVal > 0) {
                setBracket({ ...bracket, duration: Math.min(Math.max(0, numVal), remainingBracketTime) })
              }
            }}
            slotProps={{
              inputLabel: {
                shrink: true
              }
            }}
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

        {!!groupServices.length && <Box mb={4}>
          <Box mb={2}>
            <Typography>Services</Typography>
            <Typography variant="caption">Users will be able to select from these when making their appointment. At least 1 service is required.</Typography>
          </Box>
          <ToggleButtonGroup
            value={Object.keys(bracket.services || {})}
            onChange={(_, serviceIds: string[]) => {
              bracket.services = serviceIds.reduce((m, serviceId) => {
                const groupService = groupServices.find(gs => gs.service?.id == serviceId);
                if (groupService) {
                  return {
                    ...m,
                    [serviceId]: groupService.service
                  }
                }
                return { ...m }
              }, {})
              setBracket({ ...bracket });
            }}
          >
            {groupServices.map((gs, i) => {
              return <ToggleButton
                {...targets(`manage personal schedule modal toggle service ${gs.service?.name}`, `toggle service offered by the schedule`)}
                key={`bracket_service_toggle_${i}`}
                value={gs.service?.id || ''}
              >
                {gs.service?.name}
              </ToggleButton>;
            })}
          </ToggleButtonGroup>
        </Box>}
      </> :
        <ScheduleDisplay schedule={schedule || {}} setSchedule={setSchedule} />
      }
    </DialogContent>
    <DialogActions>
      <Grid container size="grow" justifyContent="space-between">
        <Grid size="grow">
          <Button
            {...targets(`manage personal schedule modal close`, `close the schedule editing modal`)}
            color="error"
            onClick={() => {
              if (closeModal) {
                firstLoad.current = true;
                closeModal();
              }
            }}
          >Cancel</Button>
        </Grid>
        {1 === viewStep ? <>
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
            color="info"
            onClick={() => {
              if (schedule?.id && bracket.duration && Object.keys(bracket.services || {}).length) {
                bracket.id = (new Date()).getTime().toString();
                bracket.scheduleId = schedule.id;
                bracket.color = generateLightBgColor();
                const newBrackets = { ...schedule.brackets };
                newBrackets[bracket.id] = bracket;
                setSchedule({ ...schedule, brackets: newBrackets })
                setBracket({ ...bracketSchema, services: {}, slots: {} });
                setViewStep(2);
              } else {
                void setSnack({ snackOn: 'Provide a duration, and at least 1 service.', snackType: 'info' });
              }
            }}
          >
            Continue
          </Button>
        </> : <>
          <Button
            {...targets(`manage personal schedule modal change colors`, `change the colors used to highlight the brackets`)}
            onClick={changeColors}
          >Change Colors</Button>
          <Button
            {...targets(`manage personal schedule modal add bracket`, `add a new bracket to the personal schedule`)}
            onClick={() => { setViewStep(1); }}
          >Add bracket</Button>
          <Button
            {...targets(`manage personal schedule modal submit`, `submit the current personal schedule for editing or creation`)}
            color="info"
            onClick={handleSubmit}
          >Submit</Button>
        </>}
      </Grid>
    </DialogActions>
  </>
}

export default ManageScheduleBracketsModal;
