import type { guiapp } from "../../wailsjs/go/models";

export type ReportSlug = "worktime-by-month" | "worktime-grouped-by-project" | "worktime-task-details" | "timesheet";

export type ReportMode = "monthly" | "daily";

export type ReportFilter = {
  mode: ReportMode;
  month: number;
  year: number;
  startDate: string;
  endDate: string;
  projectId: number;
};

export type WorkItem = { id: number; name: string; parentId?: number | null; depth: number; status?: string };

export type ReportData =
  | guiapp.WorktimeByMonthReport
  | guiapp.WorktimeGroupedByProjectReport
  | guiapp.WorktimeTaskDetailsReport
  | guiapp.TimesheetReport;

export const reportDefinitions: Array<{ slug: ReportSlug; title: string; needsProject: boolean; decimalToggle: boolean }> = [
  { slug: "worktime-by-month", title: "Worktime by month", needsProject: false, decimalToggle: false },
  { slug: "worktime-grouped-by-project", title: "Worktime grouped by project", needsProject: false, decimalToggle: false },
  { slug: "worktime-task-details", title: "Worktime task details", needsProject: true, decimalToggle: false },
  { slug: "timesheet", title: "Timesheet", needsProject: false, decimalToggle: true }
];
