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

export type ReportsPageText = {
  columns: {
    date: string;
    description: string;
    duration: string;
    end: string;
    project: string;
    projectTime: string;
    start: string;
    task: string;
    total: string;
  };
  emptyReport: string;
  exportExcel: string;
  filterMode: string;
  firstReportableProject: string;
  loadingReport: string;
  months: string[];
  monthly: string;
  print: string;
  reportList: string;
  savedTo: string;
  dateRange: string;
  titles: Record<ReportSlug, string>;
};

export const reportDefinitions: Array<{ slug: ReportSlug; needsProject: boolean; decimalToggle: boolean }> = [
  { slug: "worktime-by-month", needsProject: false, decimalToggle: false },
  { slug: "worktime-grouped-by-project", needsProject: false, decimalToggle: false },
  { slug: "worktime-task-details", needsProject: true, decimalToggle: false },
  { slug: "timesheet", needsProject: false, decimalToggle: true }
];
