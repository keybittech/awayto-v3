import React, { useContext, useMemo } from 'react';

import Box from '@mui/material/Box';
import { DigitalClock } from '@mui/x-date-pickers/DigitalClock';

import { dayjs, staticDT, targets } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function ScheduleTimePicker(_: IComponent): React.JSX.Element {

  const {
    selectedDate,
    selectedTime,
    setSelectedTime,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  const selectedDateFormat = selectedDate?.format('YYYY-MM-DD');
  const todayFormat = dayjs().format('YYYY-MM-DD');
  const hasSlotsToday = useMemo(() => {
    return dateSlots?.find(ds => {
      if (!ds.weekStart || !ds.startTime || ds.startDate != todayFormat) return false;
      return staticDT(dayjs(ds.weekStart), ds.startTime).isAfter(dayjs());
    })
  }, [dateSlots, todayFormat]);

  return <Box component="fieldset" sx={{
    border: '1px solid #666',
    borderRadius: '4px',
    '&:hover': {
      borderColor: '#fff'
    }
  }}>
    <legend style={{ fontSize: '12px', color: 'rgba(255, 255, 255, 0.7)' }}>Select Time <span style={{ display: 'contents', color: 'red', fontSize: '24px' }}>*</span></legend>
    {!hasSlotsToday && selectedDateFormat == todayFormat ? <>No available times remain.</> :
      !selectedDateFormat || !dateSlots?.length ? <>Select a date. Times will appear here.</> :
        <DigitalClock
          {...targets(`service time selection`)}
          disablePast={selectedDateFormat == todayFormat}
          skipDisabled
          value={selectedTime || null}
          onChange={time => setSelectedTime(time)}
          shouldDisableTime={time => {
            const durationToCheck = dayjs.duration(time.hour(), 'hour').add(time.minute(), 'minute');
            const checkHours = durationToCheck.hours();
            const checkMinutes = durationToCheck.minutes();
            return !dateSlots.some(dateSlot => {
              if (!dateSlot.startDate || !dateSlot.duration) return false;
              return dateSlot.startDate == selectedDateFormat &&
                checkHours == dateSlot.hour &&
                checkMinutes == dateSlot.minute;
            });
          }}

        />}
  </Box>
}

export default ScheduleTimePicker;
