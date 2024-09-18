export const scheduleSchema = {
  id: '',
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
};

export const bracketSchema = {
  duration: 1,
  automatic: false,
  multiplier: '1.00'
};
