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
import { formatInputDate } from "../dashboard/dateFormat";
import type { ReportFilter, ReportSlug } from "./reportTypes";

export function reportSlugFromHash(hash: string): ReportSlug {
  const value = hash.replace(/^#reports\/?/, "");
  if (value === "worktime-grouped-by-project" || value === "worktime-project-details" || value === "worktime-task-details" || value === "timesheet") {
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
