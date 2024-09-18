import { useCallback } from 'react';

import { DurationUnitType } from 'dayjs/plugin/duration';
import { getFormattedScheduleContext, getRelativeDuration, TimeUnit, timeUnitOrder } from './time_unit';
import { IScheduleBracketSlot } from './api';
import dayjs from 'dayjs';

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

export function useSchedule(): (schedule: UseScheduleProps) => UseScheduleResult {

  const getScheduleData = useCallback(({ scheduleTimeUnitName, bracketTimeUnitName, slotTimeUnitName, slotDuration, beginningOfMonth, bracketSlots = [] }: UseScheduleProps) => {
    console.time('GENERATING_SCHEDULE');

    let dayDiff = 0;

    if (beginningOfMonth) {
      const startOfMonthWeek = beginningOfMonth.startOf(TimeUnit.WEEK).startOf(TimeUnit.DAY);
      dayDiff = beginningOfMonth.diff(startOfMonthWeek, TimeUnit.DAY);
    }
    
    const xAxisTypeName = timeUnitOrder[timeUnitOrder.indexOf(scheduleTimeUnitName) - 1];
    const yAxisTypeName = slotTimeUnitName == bracketTimeUnitName ? bracketTimeUnitName : slotTimeUnitName;
    const divisions = beginningOfMonth ? beginningOfMonth.daysInMonth() + dayDiff : getRelativeDuration(1, scheduleTimeUnitName, xAxisTypeName);
    const selections = getRelativeDuration(1, xAxisTypeName, yAxisTypeName) / slotDuration;
    const durations = [] as CellDuration[][];

    function freshDuration () {
      return dayjs.duration(0);
    }

    let startDuration = freshDuration();
    
    for (let x = 0; x < divisions; x++) {
      if (beginningOfMonth && !beginningOfMonth.startOf(TimeUnit.WEEK).startOf(TimeUnit.DAY).add(startDuration).add(x, TimeUnit.DAY).day()) {
        startDuration = freshDuration();
      }

      durations[x] = [] as CellDuration[];

      for (let y = 0; y < selections; y++) {
        const startTime = startDuration.toISOString();

        let scheduleBracketSlotIds: string[] = [];
        if (beginningOfMonth) {
          scheduleBracketSlotIds = bracketSlots.filter(b => b.startTime === startTime).map(b => b.id!);
        }
        
        const cell = {
          contextFormat: getFormattedScheduleContext(xAxisTypeName, startDuration.toISOString(), beginningOfMonth),
          scheduleBracketSlotIds,
          startTime,
          active: !!scheduleBracketSlotIds.length,
          x, y
        };

        durations[x][y] = cell;

        startDuration = startDuration.add(slotDuration, yAxisTypeName as DurationUnitType);
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
  }, []);

  return getScheduleData;
}
