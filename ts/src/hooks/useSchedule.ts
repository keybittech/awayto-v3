import { useMemo } from 'react';
import dayjs from 'dayjs';
import { DurationUnitType } from 'dayjs/plugin/duration';

import { getRelativeDuration, timeUnitOrder } from './time_unit';
import { IScheduleBracketSlot } from './api';

type UseScheduleProps = {
  scheduleTimeUnitName?: string;
  bracketTimeUnitName?: string;
  slotTimeUnitName?: string;
  slotDuration?: number;
  bracketSlots?: IScheduleBracketSlot[];
  beginningOfMonth?: dayjs.Dayjs;
};

type CellDuration = {
  x: number;
  y: number;
  startTime: string;
  contextFormat: string;
  completeContextFormat: string;
  active: boolean;
  scheduleBracketSlotIds: string[];
}

type UseScheduleResult = {
  xAxisTypeName?: string;
  yAxisTypeName?: string;
  columns?: number;
  rows?: number;
  durations?: CellDuration[][]
}

export function useSchedule({ scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration }: UseScheduleProps): UseScheduleResult {

  return useMemo(() => {
    if (!scheduleTimeUnitName || !bracketTimeUnitName || !slotTimeUnitName || !slotDuration) return {};

    console.time('GENERATING_SCHEDULE');

    const xAxisTypeName = timeUnitOrder[timeUnitOrder.indexOf(scheduleTimeUnitName) - 1];
    const yAxisTypeName = slotTimeUnitName == bracketTimeUnitName ? bracketTimeUnitName : slotTimeUnitName;
    const columns = Math.floor(getRelativeDuration(1, scheduleTimeUnitName, xAxisTypeName));
    const rows = getRelativeDuration(1, xAxisTypeName, yAxisTypeName) / slotDuration;
    const durations = [] as CellDuration[][];
    const dayColumns = 'day' == xAxisTypeName;
    const daySlots = 'day' == slotTimeUnitName;

    for (let x = 0; x < columns + 1; x++) {
      durations[x] = [] as CellDuration[];
    }

    let baseTime = dayjs().startOf(dayColumns ? 'week' : 'year');

    let headerDuration = dayjs.duration(0);

    for (let x = 1; x < columns + 1; x++) {

      headerDuration = headerDuration.add(1, xAxisTypeName as DurationUnitType);

      let headerLabel = '';
      if (dayColumns) {
        headerLabel = baseTime.day(headerDuration.days()).format('ddd');
      } else {
        headerLabel = `Week ${baseTime.add(headerDuration.weeks() - 1, 'w').format('W')}`;
      }

      durations[x][0] = {
        contextFormat: headerLabel
      } as CellDuration;
    }

    let rowDuration = dayjs.duration(0, 'second');

    for (let y = 1; y < rows + 1; y++) {

      let djst;

      if (dayColumns) {
        djst = baseTime.day(rowDuration.days());
      } else {
        djst = baseTime.startOf('week').add(rowDuration.days(), 'day');
      }

      djst = djst.hour(rowDuration.hours()).minute(rowDuration.minutes());

      // Left column labels
      durations[0][y] = {
        contextFormat: dayColumns ? djst.format('A') : djst.format('ddd')
      } as CellDuration;

      for (let x = 1; x < columns + 1; x++) {

        const axisCorrectedDjst = djst.add(x, dayColumns ? 'day' : 'week');

        let startTime = rowDuration.add(x - 1, dayColumns ? 'day' : 'week').toISOString();
        if ('P0D' == startTime) { // match db formatting of zero durations
          startTime = 'PT0S';
        }

        durations[x][y] = {
          contextFormat: daySlots ? 'Full Day' : axisCorrectedDjst.format('hh:mm'),
          completeContextFormat: daySlots ? `${axisCorrectedDjst.format('ddd')} Week ${x}, Full Day` : axisCorrectedDjst.format('ddd, hh:mm A'),
          scheduleBracketSlotIds: [],
          active: false,
          startTime,
          x, y
        } as CellDuration;

      }

      rowDuration = rowDuration.add(slotDuration, yAxisTypeName as DurationUnitType);
    }

    console.timeEnd('GENERATING_SCHEDULE');

    return {
      xAxisTypeName,
      yAxisTypeName,
      columns,
      rows,
      durations
    };
  }, [scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration]);
}
