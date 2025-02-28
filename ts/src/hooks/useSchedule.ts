import { useMemo } from 'react';
import dayjs from 'dayjs';
import { DurationUnitType } from 'dayjs/plugin/duration';

import { getFormattedContextLabel, getFormattedTimeLabel, getFormattedScheduleContext, getRelativeDuration, timeUnitOrder } from './time_unit';
import { IScheduleBracketSlot } from './api';

type UseScheduleProps = {
  scheduleTimeUnitName: string;
  bracketTimeUnitName: string;
  slotTimeUnitName: string;
  slotDuration: number;
  bracketSlots?: IScheduleBracketSlot[];
  beginningOfMonth?: dayjs.Dayjs;
};

type CellDuration = {
  x: number;
  y: number;
  startTime: string;
  contextFormat: string;
  active: boolean;
  scheduleBracketSlotIds: string[];
}

type UseScheduleResult = {
  xAxisTypeName: string;
  yAxisTypeName: string;
  divisions: number;
  selections: number;
  durations: CellDuration[][]
}

export function useSchedule({ scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration }: UseScheduleProps): UseScheduleResult {

  return useMemo(() => {
    console.time('GENERATING_SCHEDULE');

    const xAxisTypeName = timeUnitOrder[timeUnitOrder.indexOf(scheduleTimeUnitName) - 1];
    const yAxisTypeName = slotTimeUnitName == bracketTimeUnitName ? bracketTimeUnitName : slotTimeUnitName;
    const divisions = getRelativeDuration(1, scheduleTimeUnitName, xAxisTypeName);
    const selections = getRelativeDuration(1, xAxisTypeName, yAxisTypeName) / slotDuration;
    const durations = [] as CellDuration[][];

    let startDuration = dayjs.duration(0);

    for (let x = 0; x < divisions + 1; x++) {

      durations[x] = [] as CellDuration[];

      let rowHeaderTime = dayjs.duration(0);
      if (x > 0) {
        rowHeaderTime = dayjs.duration((x - 1) * slotDuration, yAxisTypeName as DurationUnitType);
      }
      for (let y = 0; y < selections + 1; y++) {
        if (x != 0 && y != 0) {
          const startTime = startDuration.toISOString();

          durations[x][y] = {
            contextFormat: getFormattedScheduleContext(xAxisTypeName, startDuration.toISOString()), // beginningOfMonth),
            scheduleBracketSlotIds: [] as string[],
            startTime
          } as CellDuration;

          startDuration = startDuration.add(slotDuration, yAxisTypeName as DurationUnitType);
        }

        if (x == 0) { // Left Column Labels
          durations[x][y] = {
            contextFormat: getFormattedTimeLabel(xAxisTypeName, dayjs.duration((y - 1) * slotDuration, yAxisTypeName as DurationUnitType).toISOString())
          } as CellDuration;
        } else if (y == 0) { // Top Row Labels
          durations[x][y] = {
            contextFormat: getFormattedContextLabel(xAxisTypeName, startDuration.toISOString())
          } as CellDuration;
        }

      }
    }
    console.timeEnd('GENERATING_SCHEDULE');

    return {
      xAxisTypeName,
      yAxisTypeName,
      divisions,
      selections,
      durations
    };
  }, [scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration]);
}
