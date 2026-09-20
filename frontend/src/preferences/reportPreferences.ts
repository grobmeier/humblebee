import type { ReportFilter, ReportSlug, WorkItem } from "../reports/reportTypes";

export type ReportPreferences = ReportFilter & { report: ReportSlug; showDecimal: boolean };

function validDate(value: unknown): value is string {
  if (typeof value !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const date = new Date(`${value}T00:00:00Z`);
  return Number.isFinite(date.getTime()) && date.toISOString().slice(0, 10) === value;
}

export function restoreReportPreferences(value: unknown, projects: WorkItem[]): ReportPreferences | null {
  if (!value || typeof value !== "object") return null;
  const p = value as Record<string, unknown>;
  if (!["worktime-by-month", "worktime-grouped-by-project", "worktime-project-details", "worktime-task-details", "timesheet"].includes(String(p.report))) return null;
  if (p.mode !== "monthly" && p.mode !== "daily") return null;
  for (const key of ["month", "startMonth", "endMonth", "year", "projectId"]) {
    if (typeof p[key] !== "number" || !Number.isSafeInteger(p[key])) return null;
  }
  const candidate = p as unknown as ReportPreferences;
  if (candidate.month < 1 || candidate.month > 12 || candidate.startMonth < 1 || candidate.endMonth > 12 || candidate.startMonth > candidate.endMonth || candidate.year < 1 || candidate.year > 9999 || candidate.projectId < 0) return null;
  if (!validDate(p.startDate) || !validDate(p.endDate) || p.startDate > p.endDate || typeof p.showDecimal !== "boolean") return null;
  return {
    report: candidate.report, mode: candidate.mode, month: candidate.month,
    startMonth: candidate.startMonth, endMonth: candidate.endMonth, year: candidate.year,
    startDate: candidate.startDate, endDate: candidate.endDate, showDecimal: candidate.showDecimal,
    projectId: projects.some((project) => project.parentId == null && project.id === candidate.projectId && project.name.toLowerCase() !== "default") ? candidate.projectId : 0
  };
}
