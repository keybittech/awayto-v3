import React, { useState, useCallback, useMemo, useEffect } from 'react';

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

import { DesktopDatePicker } from '@mui/x-date-pickers/DesktopDatePicker';

import { siteApi, useTimeName, IGroupSchedule, ITimeUnit, TimeUnit, timeUnitOrder, getRelativeDuration, ISchedule, useDebounce, useStyles, dayjs, targets, IValidationAreas, useValid } from 'awayto/hooks';
import SelectLookup from '../common/SelectLookup';
import ScheduleDisplay from '../schedules/ScheduleDisplay';

export const scheduleSchema = {
  name: '',
  startTime: '',
  endTime: '',
  timezone: '',
  slotDuration: 30,
  scheduleTimeUnitId: '',
  scheduleTimeUnitName: '',
  bracketTimeUnitId: '',
  bracketTimeUnitName: '',
  slotTimeUnitId: '',
  slotTimeUnitName: ''
} as ISchedule;

interface ManageSchedulesModalProps extends IComponent {
  showCancel?: boolean;
  editGroupSchedule?: IGroupSchedule;
  validArea?: keyof IValidationAreas;
  saveToggle?: number;
}

export function ManageSchedulesModal({ children, editGroupSchedule, validArea, saveToggle = 0, showCancel = true, closeModal, ...props }: ManageSchedulesModalProps): React.JSX.Element {

  const classes = useStyles();

  const { setValid } = useValid();

  const [patchGroupSchedule] = siteApi.useGroupScheduleServicePatchGroupScheduleMutation();
  const [postGroupSchedule] = siteApi.useGroupScheduleServicePostGroupScheduleMutation();
  const [getGroupScheduleMasterById] = siteApi.useLazyGroupScheduleServiceGetGroupScheduleMasterByIdQuery();

  const { data: lookups, isSuccess: lookupsRetrieved } = siteApi.useLookupServiceGetLookupsQuery();

  const [groupSchedule, setGroupSchedule] = useState({
    schedule: {
      ...scheduleSchema,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
    },
    ...editGroupSchedule,
    ...JSON.parse(localStorage.getItem(`${validArea}_schedule`) || '{}') as IGroupSchedule
  } as IGroupSchedule);

  const schedule = useMemo(() => (groupSchedule.schedule || { ...scheduleSchema }) as Required<ISchedule>, [groupSchedule.schedule]);

  const debouncedSchedule = useDebounce(schedule, 150);

  const scheduleTimeUnitName = useTimeName(schedule.scheduleTimeUnitId);
  const bracketTimeUnitName = useTimeName(schedule.bracketTimeUnitId);
  const slotTimeUnitName = useTimeName(schedule.slotTimeUnitId);

  const setDefault = useCallback((scheduleType: string) => {
    const week = lookups?.timeUnits?.find(s => s.name === TimeUnit.WEEK);
    const hour = lookups?.timeUnits?.find(s => s.name === TimeUnit.HOUR);
    const day = lookups?.timeUnits?.find(s => s.name === TimeUnit.DAY);
    const minute = lookups?.timeUnits?.find(s => s.name === TimeUnit.MINUTE);
    const month = lookups?.timeUnits?.find(s => s.name === TimeUnit.MONTH);
    if (!week || !hour || !day || !minute || !month) return;
    if ('hoursweekly30minsessions' == scheduleType) {
      setGroupSchedule({
        schedule: {
          ...schedule,
          scheduleTimeUnitName: week.name,
          scheduleTimeUnitId: week.id,
          bracketTimeUnitName: hour.name,
          bracketTimeUnitId: hour.id,
          slotTimeUnitName: minute.name,
          slotTimeUnitId: minute.id,
          slotDuration: 30
        }
      });
    } else if ('dailybookingpermonth' == scheduleType) {
      setGroupSchedule({
        schedule: {
          ...schedule,
          scheduleTimeUnitName: month.name,
          scheduleTimeUnitId: month.id,
          bracketTimeUnitName: week.name,
          bracketTimeUnitId: week.id,
          slotTimeUnitName: day.name,
          slotTimeUnitId: day.id,
          slotDuration: 1
        }
      });
    }
  }, [lookups, schedule]);

  const slotDurationMarks = useMemo(() => {
    const factors = [] as { value: number, label: number }[];
    if (!bracketTimeUnitName || !slotTimeUnitName || !scheduleTimeUnitName) return factors;

    const relativeDuration = Math.round(getRelativeDuration(1, bracketTimeUnitName, slotTimeUnitName));

    const minimumSlotDuration = 'minute' == slotTimeUnitName ? 5 : 1;

    for (let value = minimumSlotDuration; value <= Math.ceil(relativeDuration / 2); value++) {
      if (relativeDuration % value === 0) {
        factors.push({ value, label: value });
      }
    }
    return factors;
  }, [bracketTimeUnitName, slotTimeUnitName, scheduleTimeUnitName]);

  const handleSubmit = useCallback(async () => {
    if (groupSchedule.schedule) {
      groupSchedule.schedule.startTime = groupSchedule.schedule.startTime?.length ? groupSchedule.schedule.startTime : undefined;
      groupSchedule.schedule.endTime = groupSchedule.schedule.endTime?.length ? groupSchedule.schedule.endTime : undefined;

      if (validArea != 'onboarding') {
        const s = groupSchedule.schedule;

        const newSchedule = {
          name: s.name,
          startTime: s.startTime,
          endTime: s.endTime
        } as ISchedule;

        if (s.id) {
          newSchedule.id = s.id;
          await patchGroupSchedule({
            patchGroupScheduleRequest: {
              groupSchedule: {
                schedule: newSchedule
              }
            }
          }).unwrap();
        } else {
          newSchedule.scheduleTimeUnitId = s.scheduleTimeUnitId
          newSchedule.bracketTimeUnitId = s.bracketTimeUnitId
          newSchedule.slotTimeUnitId = s.slotTimeUnitId
          newSchedule.slotDuration = s.slotDuration
          await postGroupSchedule({
            postGroupScheduleRequest: {
              groupSchedule: {
                schedule: newSchedule
              }
            }
          }).unwrap();
        }
      }
    }

    closeModal && closeModal(groupSchedule);
  }, [groupSchedule]);

  // Onboarding handling
  useEffect(() => {
    if (validArea) {
      localStorage.setItem(`${validArea}_schedule`, JSON.stringify({ schedule: debouncedSchedule }));
      setValid({ area: validArea, schema: 'schedule', valid: Boolean(debouncedSchedule.name) });
    }
  }, [validArea, debouncedSchedule]);

  // Onboarding handling
  useEffect(() => {
    if (saveToggle > 0) {
      handleSubmit();
    }
  }, [saveToggle]);

  useEffect(() => {
    async function go() {
      if (lookupsRetrieved) {
        if (editGroupSchedule?.schedule?.id) {
          const masterSchedule = await getGroupScheduleMasterById({ groupScheduleId: editGroupSchedule.schedule.id }).unwrap();
          setGroupSchedule({ ...scheduleSchema, ...masterSchedule.groupSchedule });
        } else if (!schedule.scheduleTimeUnitId) {
          setDefault('hoursweekly30minsessions');
        }
      }
    }
    void go();
  }, [lookupsRetrieved, editGroupSchedule]);

  useEffect(() => {
    if (!groupSchedule.schedule?.scheduleTimeUnitName && scheduleTimeUnitName && bracketTimeUnitName && slotTimeUnitName) {
      setGroupSchedule({
        schedule: {
          ...groupSchedule.schedule,
          scheduleTimeUnitName,
          bracketTimeUnitName,
          slotTimeUnitName
        }
      });
    }
  }, [groupSchedule, scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName]);

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
                {...targets(`manage schedule modal schedule name`, `Schedule Name`, `change the name of the schedule`)}
                fullWidth
                disabled={!!schedule.id}
                helperText="Ex: Spring 2022 Campaign, Q1 Offering"
                value={schedule.name || ''}
                required
                onChange={e => setGroupSchedule({ schedule: { ...schedule, name: e.target.value } })}
              />
            </Box>

            <Box mb={4}>
              <Grid container spacing={2} direction="row">
                <Grid size={6}>
                  <DesktopDatePicker
                    {...targets(`manage schedule modal start date`, `Start Date`, `set when the schedule should start`)}
                    value={schedule.startTime ? dayjs(schedule.startTime) : null}
                    format="MM/DD/YYYY"
                    formatDensity="spacious"
                    onChange={date => setGroupSchedule({ schedule: { ...schedule, startTime: date ? date.toISOString() : undefined } })}
                    disableHighlightToday={true}
                    slotProps={{
                      openPickerButton: {
                        color: 'secondary'
                      },
                      clearIcon: { sx: { color: 'red' } },
                      field: {
                        clearable: true,
                        onClear: () => setGroupSchedule({ schedule: { ...schedule, startTime: undefined } })
                      },
                      textField: {
                        fullWidth: true,
                        helperText: 'Services can be scheduled starting this date. Clear to disable the schedule. With monthly schedules, the 28 days start on this date. '
                      }
                    }}
                  />
                </Grid>
                <Grid size={6}>
                  <DesktopDatePicker
                    {...targets(`manage schedule modal end date`, `End Date`, `set when the schedule should end`)}
                    value={schedule.endTime ? dayjs(schedule.endTime) : null}
                    format="MM/DD/YYYY"
                    formatDensity="spacious"
                    onChange={date => setGroupSchedule({ schedule: { ...schedule, endTime: date ? date.toISOString() : undefined } })}
                    disableHighlightToday={true}
                    slotProps={{
                      openPickerButton: {
                        color: 'secondary'
                      },
                      clearIcon: { sx: { color: 'red' } },
                      field: {
                        clearable: true,
                        onClear: () => setGroupSchedule({ schedule: { ...schedule, endTime: undefined } })
                      },
                      textField: {
                        fullWidth: true,
                        helperText: 'Optional. No bookings will be allowed after this date.'
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
              {!schedule.id && <>
                <Typography variant="body2">Use premade selections for this schedule.</Typography>
                <Button
                  {...targets(`manage schedule modal preset hoursweekly`, `use schedule preset that resets weekly and has 30 minute appointments`)}
                  color="secondary"
                  onClick={() => setDefault('hoursweekly30minsessions')}
                >reset weekly, 30 minute appointments</Button><br />
                <Button
                  {...targets(`manage schedule modal preset dailymonth`, `use schedule preset that resets monthly and has full day bookings`)}
                  color="secondary"
                  onClick={() => setDefault('dailybookingpermonth')}
                >reset monthly, full-day booking</Button>
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
                  const { id, name } = lookups?.timeUnits?.find(c => c.id == val) || {};
                  if (!id || !name) return;
                  const validBracket = lookups?.timeUnits?.find(s => s.name == timeUnitOrder[timeUnitOrder.indexOf(name) - 1]);
                  if (!validBracket?.name) return;
                  const validSlot = lookups.timeUnits.find(s => s.name == timeUnitOrder[timeUnitOrder.indexOf(validBracket.name!) - 1]);
                  if (!validSlot?.id) return;
                  setGroupSchedule({
                    schedule: {
                      ...schedule,
                      scheduleTimeUnitName: name,
                      scheduleTimeUnitId: id,
                      bracketTimeUnitName: validBracket.name,
                      bracketTimeUnitId: validBracket.id,
                      slotTimeUnitName: validSlot.name,
                      slotTimeUnitId: validSlot.id,
                      slotDuration: 1,
                    }
                  });
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
                  if (!name || !id) return;
                  const validSlot = lookups.timeUnits.find(s => s.name == timeUnitOrder[timeUnitOrder.indexOf(name!) - 1]);
                  if (!validSlot) return;
                  setGroupSchedule({
                    schedule: {
                      ...schedule,
                      bracketTimeUnitName: name,
                      bracketTimeUnitId: id,
                      slotTimeUnitName: validSlot.name,
                      slotTimeUnitId: validSlot.id,
                      slotDuration: validSlot.name == 'minute' ? 5 : 1,
                    }
                  });
                }}
                {...props}
              />
            </Box>

            <Box mb={4}>
              <TextField
                {...targets(`manage schedule modal slot context name`, `Booking Slot Length`, `an uneditable field showing the currently selected slot context name`)}
                disabled={true}
                value={slotTimeUnitName || ''}
                helperText={slotDurationMarks.length > 1 ? `The # of ${slotTimeUnitName}s to deduct from the bracket upon accepting a booking. Alternatively, if you meet with clients, this is the length of time per session.` : 'The booking slot will be for an entire day.'}
                slotProps={{
                  inputLabel: {
                    shrink: true
                  }
                }}
              />

              {slotDurationMarks.length > 1 && <Box mt={2} sx={{ display: 'flex', alignItems: 'baseline' }}>
                <Box>{schedule.slotDuration} <span>&nbsp;</span> &nbsp;</Box>
                <Slider
                  {...targets(`manage schedule modal slot value`, `adjust the numeric value of the slot duration`)}
                  disabled={!!schedule.id}
                  value={schedule.slotDuration}
                  step={null}
                  marks={slotDurationMarks}
                  max={Math.max(...slotDurationMarks.map(m => m.value))}
                  onChange={(_, val) => {
                    setGroupSchedule({ schedule: { ...schedule, slotDuration: parseFloat(val.toString()) } });
                  }}
                />
              </Box>}
            </Box>
          </Box>
        </Grid>
        <Grid size={12} sx={{}}>
          <Box component="fieldset" p={2} sx={classes.legendBox}>
            <legend>Step 3. Review</legend>
            <Typography pb={2} variant="body1">Preview what your users will see when filling out their own schedules.</Typography>
            <ScheduleDisplay schedule={schedule} />
          </Box>
        </Grid>
        <Grid size={12}>
          <Alert color="info">Note: Schedules become read-only after creation, excluding start and end date. Make sure everything looks good here before saving!</Alert>
        </Grid>
      </Grid>
    </CardContent>
    {validArea != 'onboarding' && <CardActions>
      <Grid size="grow" container justifyContent={showCancel ? "space-between" : "flex-end"}>
        {showCancel && <Button
          {...targets(`manage schedule modal close`, `close the schedule management modal`)}
          color="error"
          onClick={closeModal}
        >Cancel</Button>}
        <Button
          {...targets(`manage schedule modal submit`, `submit the current schedule for editing or creation`)}
          disabled={!schedule.name}
          color="info"
          onClick={handleSubmit}
        >Save Schedule</Button>
      </Grid>
    </CardActions>}
  </Card>
}

export default ManageSchedulesModal;
