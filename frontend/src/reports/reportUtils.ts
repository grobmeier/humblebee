import type { guiapp } from "../../wailsjs/go/models";
import { formatInputDate } from "../dashboard/dateFormat";
import type { ReportFilter, ReportSlug } from "./reportTypes";

export function reportSlugFromHash(hash: string): ReportSlug {
  const value = hash.replace(/^#reports\/?/, "");
  if (value === "worktime-grouped-by-project" || value === "worktime-task-details" || value === "timesheet") {
    return value;
  }
  return "worktime-by-month";
}

export function defaultReportFilter(): ReportFilter {
  const now = new Date();
  return {
    mode: "monthly",
    month: now.getMonth() + 1,
    year: now.getFullYear(),
    startDate: formatInputDate(new Date(now.getFullYear(), now.getMonth(), 1)),
    endDate: formatInputDate(now),
    projectId: 0
  };
}

export function toReportRequest(filter: ReportFilter, language: string): guiapp.ReportRequest {
  return {
    mode: filter.mode,
    month: filter.month,
    year: filter.year,
    startDate: filter.startDate,
    endDate: filter.endDate,
    projectId: filter.projectId,
    language
  } as guiapp.ReportRequest;
}

export function formatDecimalDuration(duration: string): string {
  const match = /^(\d+):(\d{2})$/.exec(duration);
  if (!match) {
    return duration;
  }
  const hours = Number(match[1]);
  const minutes = Number(match[2]);
  return (hours + minutes / 60).toFixed(2);
}

export function fileURL(path: string): string {
  return `file://${path}`;
}

export function monthOptions(months: string[]) {
  return months.map((label, index) => [String(index + 1), label] as const);
}
