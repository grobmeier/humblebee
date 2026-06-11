/*
 * Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import type { guiapp } from "../../wailsjs/go/models";

export type ReportSlug = "worktime-by-month" | "worktime-grouped-by-project" | "worktime-project-details" | "worktime-task-details" | "timesheet";

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
  | guiapp.WorktimeProjectDetailsReport
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
  selectProject: string;
  dateRange: string;
  titles: Record<ReportSlug, string>;
};

export const reportDefinitions: Array<{ slug: ReportSlug; needsProject: boolean; requiresExplicitProject: boolean; decimalToggle: boolean }> = [
  { slug: "worktime-by-month", needsProject: false, requiresExplicitProject: false, decimalToggle: false },
  { slug: "worktime-grouped-by-project", needsProject: false, requiresExplicitProject: false, decimalToggle: false },
  { slug: "worktime-project-details", needsProject: true, requiresExplicitProject: true, decimalToggle: false },
  { slug: "worktime-task-details", needsProject: true, requiresExplicitProject: false, decimalToggle: false },
  { slug: "timesheet", needsProject: false, requiresExplicitProject: false, decimalToggle: true }
];
