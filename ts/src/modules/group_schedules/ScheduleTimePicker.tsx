import React, { useContext } from 'react';

import Box from '@mui/material/Box';
import { DigitalClock } from '@mui/x-date-pickers/DigitalClock';

import { dayjs } from 'awayto/hooks';

import GroupScheduleSelectionContext, { GroupScheduleSelectionContextType } from './GroupScheduleSelectionContext';

export function ScheduleTimePicker(_: IComponent): React.JSX.Element {

  const {
    selectedTime,
    setSelectedTime,
    dateSlots,
  } = useContext(GroupScheduleSelectionContext) as GroupScheduleSelectionContextType;

  return <Box component="fieldset" sx={{
    border: '1px solid #666',
    borderRadius: '4px',
    '&:hover': {
      borderColor: '#fff'
    }
  }}>
    <legend style={{ fontSize: '12px', color: 'rgba(255, 255, 255, 0.7)' }}>Select Time</legend>
    {dateSlots?.length && <DigitalClock
      value={selectedTime}
      onChange={time => setSelectedTime(time)}
      skipDisabled
      shouldDisableTime={time => {
        let durationToCheck = dayjs.duration(time.hour(), 'hour').add(time.minute(), 'minute');
        return !dateSlots.some(s => {
          if (!s.startDate || !s.startTime) return false;
          const startTimeDuration = dayjs.duration(s.startTime);
          return durationToCheck.hours() == startTimeDuration.hours() &&
            durationToCheck.minutes() == startTimeDuration.minutes();
        });
      }}
    />}
  </Box>
}

export default ScheduleTimePicker;
