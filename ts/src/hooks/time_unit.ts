import dayjs from 'dayjs';

/**
 * @category Time Unit
 * @purpose lists the available time contexts for use in the site based on calendar unit
 */
export type ITimeUnitNames = string | ('minute' | 'hour' | 'day' | 'week' | 'month' | 'year');

/**
 * @category Time Unit
 * @purpose provides enumerator based definition and selection for Time Unit related objects
 */
export enum TimeUnit {
  MINUTE = 'minute',
  HOUR = 'hour',
  DAY = 'day',
  WEEK = 'week',
  MONTH = 'month',
  YEAR = 'year'
}

/**
 * @category Time Unit
 */
export const timeUnitOrder = [
  'minute',
  'hour',
  'day',
  'week',
  'month',
  'year'
] as ITimeUnitNames[];

/**
 * @category Time Unit
 */
export const millisTimeUnits: Record<ITimeUnitNames, number> = {
  [TimeUnit.YEAR]: 31536000000,
  [TimeUnit.MONTH]: 2629800000,
  [TimeUnit.WEEK]: 604800000,
  [TimeUnit.DAY]: 86400000,
  [TimeUnit.HOUR]: 3600000,
  [TimeUnit.MINUTE]: 60000
}

/**
 * @category Time Unit
 */
export function getRelativeDuration(amount: number, fromUnit: ITimeUnitNames = TimeUnit.HOUR, toUnit: ITimeUnitNames = TimeUnit.MINUTE): number {
  const fromDuration = millisTimeUnits[fromUnit];
  const toDuration = millisTimeUnits[toUnit];
  return (amount * fromDuration) / toDuration;
}

/**
 * @category Time Unit
 */
export function plural(n: number, singular: string, plural: string): string {
  return n.toString() + ' ' + (n === 1 ? singular : plural);
}

/**
 * @category Time Unit
 */
export function staticDT(weekStart: dayjs.Dayjs, startTime: string): dayjs.Dayjs {
  const d = dayjs.duration(startTime);
  return weekStart.day(d.days()).hour(d.hours()).minute(d.minutes());
}

/**
 * @category Time Unit
 */
export function quotedDT(startDate: string, startTime: string): dayjs.Dayjs {
  const d = dayjs.duration(startTime);
  return dayjs(startDate).startOf('day').hour(d.hours()).minute(d.minutes());
}



/**
 * @category Time Unit
 */
export function bookingDT(slotDate: string, startTime: string): dayjs.Dayjs {
  return staticDT(dayjs(slotDate).startOf('day').startOf('week'), startTime);
}

/**
 * @category Time Unit
 */
export function bookingDTHours(slotDate: string, startTime: string): string {
  return bookingDT(slotDate, startTime).format("hh:mm a");
}

/**
 * @category Time Unit
 */
export function shortNSweet(slotDate: string, startTime: string): string {
  return `${dayjs(slotDate).format("ddd, MMM D")} at ${bookingDTHours(slotDate, startTime)}`;
}

/**
 * @category Time Unit
 */
export function utcDTLocal(utc: string): string {
  return dayjs.utc(utc).local().format("YYYY-MM-DD hh:mm a");
}

/**
 * @category Time Unit
 */
export function utcNow(): dayjs.Dayjs {
  return dayjs.utc(new Date());
}

/**
 * @category Time Unit
 */
export function utcNowString(): string {
  return dayjs.utc(new Date()).toISOString();
}
