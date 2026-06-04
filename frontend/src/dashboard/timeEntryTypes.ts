export type TimeEntryFormState = {
  description: string;
  endDate: string;
  endTime: string;
  id?: number;
  projectId: number;
  startDate: string;
  startTime: string;
  taskId: number;
  untilMidnight: boolean;
};
