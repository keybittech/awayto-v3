import React, { useState, useCallback, useMemo, useEffect, useRef } from 'react';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Slider from '@mui/material/Slider';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

import { useUtil, siteApi, useTimeName, IGroupSchedule, ITimeUnit, TimeUnit, timeUnitOrder, getRelativeDuration, scheduleSchema, ISchedule, useDebounce, useStyles } from 'awayto/hooks';
import SelectLookup from '../common/SelectLookup';
import ScheduleDisplay from '../schedules/ScheduleDisplay';

interface ManageSchedulesModalProps extends IComponent {
  showCancel?: boolean;
  editGroupSchedule?: IGroupSchedule;
  onValidChanged?: (valid: boolean) => void;
  saveToggle?: number;
}

export function ManageSchedulesModal({ children, editGroupSchedule, onValidChanged, saveToggle = 0, showCancel = true, closeModal, ...props }: ManageSchedulesModalProps): React.JSX.Element {

  const classes = useStyles();
  const { setSnack } = useUtil();

  const [patchGroupSchedule] = siteApi.useGroupScheduleServicePatchGroupScheduleMutation();
  const [postGroupSchedule] = siteApi.useGroupScheduleServicePostGroupScheduleMutation();
  const [getGroupScheduleMasterById] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleMasterByIdQuery();

  const { data: lookups, isSuccess: lookupsRetrieved } = siteApi.useLookupServiceGetLookupsQuery();

  const [groupSchedule, setGroupSchedule] = useState({
    schedule: {
      ...scheduleSchema,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
    },
    ...editGroupSchedule
  } as IGroupSchedule);

  const schedule = useMemo(() => (groupSchedule.schedule || { ...scheduleSchema }) as Required<ISchedule>, [groupSchedule.schedule]);

  const debouncedSchedule = useDebounce(schedule, 50);

  const scheduleTimeUnitName = useTimeName(schedule.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(schedule.bracketTimeUnitId);
  const slotTimeUnitName = useTimeName(schedule.slotTimeUnitId);

  const setDefault = useCallback((scheduleType: string) => {
    const weekId = lookups?.timeUnits?.find(s => s.name === TimeUnit.WEEK)?.id;
    const hourId = lookups?.timeUnits?.find(s => s.name === TimeUnit.HOUR)?.id;
    const dayId = lookups?.timeUnits?.find(s => s.name === TimeUnit.DAY)?.id;
    const minuteId = lookups?.timeUnits?.find(s => s.name === TimeUnit.MINUTE)?.id;
    const monthId = lookups?.timeUnits?.find(s => s.name === TimeUnit.MONTH)?.id;
    if ('hoursweekly30minsessions' == scheduleType) {
      setGroupSchedule({
        schedule: {
          ...schedule,
          scheduleTimeUnitId: weekId as string,
          bracketTimeUnitId: hourId as string,
          slotTimeUnitId: minuteId as string,
          slotDuration: 30
        }
      });
    } else if ('dailybookingpermonth' == scheduleType) {
      setGroupSchedule({
        schedule: {
          ...schedule,
          scheduleTimeUnitId: monthId as string,
          bracketTimeUnitId: weekId as string,
          slotTimeUnitId: dayId as string,
          slotDuration: 1
        }
      });
    }
  }, [lookups, schedule]);

  const slotDurationMarks = useMemo(() => {
    const factors = [] as { value: number, label: number }[];
    if (!bracketTimeUnitName || !slotTimeUnitName || !scheduleTimeUnitName) return factors;
    // const subdivided = bracketTimeUnitName !== slotTimeUnitName;
    // const finalDuration = !subdivided ? 
    //   Math.round(getRelativeDuration(duration, scheduleTimeUnitName, bracketTimeUnitName)) : 
    const finalDuration = Math.round(getRelativeDuration(1, bracketTimeUnitName, slotTimeUnitName));
    for (let value = 1; value <= finalDuration; value++) {
      if (finalDuration % value === 0) {
        factors.push({ value, label: value });
      }
    }
    return factors;
  }, [bracketTimeUnitName, slotTimeUnitName, scheduleTimeUnitName]);

  const handleSubmit = useCallback(async () => {
    if (!groupSchedule.schedule || !groupSchedule.schedule.name) {
      setSnack({ snackOn: 'A name must be provided.', snackType: 'warning' });
      return;
    }

    if (!onValidChanged) {
      if (groupSchedule.scheduleId) {
        await patchGroupSchedule({ patchGroupScheduleRequest: { groupSchedule } }).unwrap();
      } else {
        await postGroupSchedule({ postGroupScheduleRequest: { groupSchedule } }).unwrap();
      }
    }

    closeModal && closeModal(groupSchedule);
  }, [groupSchedule]);

  // Onboarding handling
  useEffect(() => {
    if (onValidChanged) {
      onValidChanged(Boolean(debouncedSchedule.name && debouncedSchedule.startTime));
    }
  }, [debouncedSchedule]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  useEffect(() => {
    async function go() {
      if (lookupsRetrieved) {
        if (schedule.id) {
          const { groupSchedule } = await getGroupScheduleMasterById({ groupScheduleId: schedule.id }).unwrap();
          setGroupSchedule(groupSchedule);
        } else if (!editGroupSchedule?.schedule?.name) {
          setDefault('hoursweekly30minsessions');
        }
      }
    }
    void go();
  }, [lookupsRetrieved, schedule.id]);

  if (!lookups?.timeUnits) return <></>;

  return <Card>
    <CardHeader title={`${schedule.id ? 'Edit' : 'Create'} Schedule`}></CardHeader>
    <CardContent>
      {!!children && children}
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Box component="fieldset" p={2} sx={classes.legendBox}>
            <legend>Step 1. Provide basic details</legend>

            <Box mb={4}>
              <TextField
                fullWidth
                disabled={!!schedule.id}
                label="Name"
                helperText="Ex: Spring 2022 Campaign, Q1 Offering"
                value={schedule.name || ''}
                required
                onChange={e => setGroupSchedule({ schedule: { ...schedule, name: e.target.value } })}
              />
            </Box>

            <Box mb={4}>
              <Grid container spacing={2} direction="row">
                <Grid size={6}>

                  <TextField
                    fullWidth
                    label="Start Date"
                    type="date"
                    value={schedule.startTime || ''}
                    required
                    helperText="Schedule is active after this date. Clear this date to deactivate the schedule. Deactivated schedules do not allow new bookings to be made."
                    onChange={e => setGroupSchedule({ schedule: { ...schedule, startTime: e.target.value } })}
                    slotProps={{
                      inputLabel: {
                        shrink: true
                      }
                    }}
                  />
                  {/* <DesktopDatePicker
                label="Start Date"
                value={schedule.startTime ? schedule.startTime : null}
                onChange={e => {
                  // if (e) {
                  //   console.log({ schedule, e, str: e ? nativeJs(new Date(e.toString())) : null })
                  //   setGroupSchedule({ ...schedule, startTime: nativeJs(new Date(e.toString())).toString() })
                  //   // setGroupSchedule({ ...schedule, startTime: e ? e.format(DateTimeFormatter.ofPattern("yyyy-mm-dd")) : '' });
                  // }
                }}
                renderInput={(params) => <TextField helperText="Bookings can be scheduled any time after this date. Removing this value and saving the schedule will deactivate it, preventing it from being seen during booking." {...params} />}
              /> */}
                </Grid>
                <Grid size={6}>
                  <TextField
                    fullWidth
                    label="End Date"
                    type="date"
                    value={schedule.endTime || ''}
                    helperText="Optional. No bookings will be allowed after this date."
                    onChange={e => setGroupSchedule({ schedule: { ...schedule, endTime: e.target.value } })}
                    slotProps={{
                      inputLabel: {
                        shrink: true
                      }
                    }}
                  />
                </Grid>
              </Grid>
            </Box>
          </Box>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Box component="fieldset" p={2} sx={classes.legendBox}>
            <legend>Step 2. Configure durations</legend>
            <Box mb={4}>
              {!schedule.id ? <>
                <Typography variant="body2">Use premade selections for this schedule.</Typography>
                <Button color="secondary" onClick={() => setDefault('hoursweekly30minsessions')}>reset weekly, 30 minute appointments</Button><br />
                <Button color="secondary" onClick={() => setDefault('dailybookingpermonth')}>reset monthly, full-day booking</Button>
              </> : <>
                <Alert color="info">Schedule template durations are read-only after creation.</Alert>
              </>}
            </Box>

            <Box mb={4}>
              <SelectLookup
                noEmptyValue
                disabled={!!schedule.id}
                lookupName="Schedule Duration"
                helperText="The length of time the schedule will run over. This determines the overall context of the schedule and how time will be divided and managed within. For example, a 40 hour per week schedule would require configuring a 1 week Schedule Duration."
                lookupValue={schedule.scheduleTimeUnitId}
                lookups={lookups?.timeUnits?.filter(sc => [TimeUnit.WEEK, TimeUnit.MONTH].includes(sc.name as TimeUnit))}
                lookupChange={(val: string) => {
                  const { id, name } = lookups?.timeUnits?.find(c => c.id === val) || {};
                  if (!id || !name) return;
                  setGroupSchedule({ schedule: { ...schedule, scheduleTimeUnitName: name, scheduleTimeUnitId: id, bracketTimeUnitId: lookups?.timeUnits?.find(s => s.name === timeUnitOrder[timeUnitOrder.indexOf(name) - 1])?.id as string } })
                }}
                {...props}
              />
            </Box>

            {/* <Box mb={4}>
          <TextField
            fullWidth
            type="number"
            disabled={!!schedule.id}
            label={`# of ${schedule.scheduleTimeUnitName}s`}
            helperText="Provide a duration. After this duration, the schedule will reset, and all bookings will be available again."
            value={schedule.duration}
            onChange={e => {
              setGroupSchedule({ ...schedule, duration: Math.min(Math.max(0, parseInt(e.target.value || '', 10)), 999) })
            }}
          />
        </Box> */}
            {/* {lookups?.timeUnits?.filter(sc => sc.name && sc.name !== scheduleTimeUnitName && timeUnitOrder.indexOf(sc.name) <= timeUnitOrder.indexOf(scheduleTimeUnitName))} */}
            <Box mb={4}>
              <SelectLookup
                noEmptyValue
                disabled={!!schedule.id}
                lookupName="Bracket Duration Type"
                helperText="How to measure blocks of time within the Schedule Duration. For example, in a 40 hour per week situation, blocks of time are divided in hours. Multiple brackets can be used on a single schedule, and all of them share the same Bracket Duration Type."
                lookupValue={schedule.bracketTimeUnitId}
                lookups={
                  'month' == scheduleTimeUnitName ? lookups.timeUnits.filter(sc => sc.name == 'week' || sc.name == 'day') :
                    'week' == scheduleTimeUnitName ? lookups.timeUnits.filter(sc => sc.name == 'day' || sc.name == 'hour') :
                      []
                }
                lookupChange={(val: string) => {
                  const { name, id } = lookups?.timeUnits?.find(c => c.id === val) as ITimeUnit;
                  setGroupSchedule({ schedule: { ...schedule, bracketTimeUnitName: name, bracketTimeUnitId: id, slotTimeUnitName: name, slotTimeUnitId: id, slotDuration: 1 } });
                }}
                {...props}
              />
            </Box>

            <Box mb={4}>
              <SelectLookup
                noEmptyValue
                disabled={!!schedule.id}
                lookupName="Booking Slot Length"
                helperText={`The # of ${slotTimeUnitName}s to deduct from the bracket upon accepting a booking. Alternatively, if you meet with clients, this is the length of time per session.`}
                lookupValue={schedule.slotTimeUnitId}
                lookups={
                  'hour' == bracketTimeUnitName ? lookups.timeUnits.filter(sc => sc.name == 'minute' || sc.name == 'hour') :
                    'day' == bracketTimeUnitName ? lookups.timeUnits.filter(sc => sc.name == 'day' || sc.name == 'hour') :
                      'week' == bracketTimeUnitName ? lookups.timeUnits.filter(sc => sc.name == 'day' || sc.name == 'hour') :
                        []
                }
                lookupChange={(val: string) => {
                  const { name, id } = lookups?.timeUnits?.find(c => c.id === val) || {};
                  if (!name || !id) return;
                  setGroupSchedule({ schedule: { ...schedule, slotTimeUnitName: name, slotTimeUnitId: id, slotDuration: 1 } })
                }}
                {...props}
              />

              <Box mt={2} sx={{ display: 'flex', alignItems: 'baseline' }}>
                <Box>{schedule.slotDuration} <span>&nbsp;</span> &nbsp;</Box>
                <Slider
                  disabled={!!schedule.id}
                  value={schedule.slotDuration}
                  step={null}
                  marks={slotDurationMarks}
                  max={Math.max(...slotDurationMarks.map(m => m.value))}
                  onChange={(_, val) => {
                    setGroupSchedule({ schedule: { ...schedule, slotDuration: parseFloat(val.toString()) } });
                  }}
                />
              </Box>
            </Box>
          </Box>
        </Grid>
        <Grid size={12} sx={{}}>
          <Box component="fieldset" p={2} sx={classes.legendBox}>
            <legend>Step 3. Review</legend>
            <Typography pb={2} variant="body1">This is what your users will see when filling out their own schedules.</Typography>
            <ScheduleDisplay schedule={schedule} />
          </Box>
        </Grid>
      </Grid>
    </CardContent>
    {
      !onValidChanged && <CardActions>
        <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
          {showCancel && <Button onClick={closeModal}>Cancel</Button>}
          <Button disabled={!schedule.name || !schedule.startTime} onClick={handleSubmit}>Save Schedule</Button>
        </Grid>
      </CardActions>
    }
  </Card >
}

export default ManageSchedulesModal;
