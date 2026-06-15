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

import type { ReactNode } from "react";
import type { guiapp } from "../../wailsjs/go/models";
import { formatDecimalDuration } from "./reportUtils";
import type { ReportsPageText } from "./reportTypes";

type ReportTableProps<T> = {
  language: string;
  report: T;
  showDecimal: boolean;
  t: ReportsPageText;
};

function displayDuration(duration: string, showDecimal: boolean, language: string): string {
  return showDecimal ? formatDecimalDuration(duration, language) : duration;
}

export function WorktimeByMonthTable({ language, report, showDecimal, t }: ReportTableProps<guiapp.WorktimeByMonthReport>) {
  const rows = report.rows ?? [];
  if (report.empty || !rows.length) return <p className="projects-empty">{t.emptyReport}</p>;
  return (
    <ReportTable headers={[t.columns.project, t.columns.task, t.columns.date, t.columns.start, t.columns.end, t.columns.duration, t.columns.description]}>
      {rows.map((row, index) => (
        <tr key={`${row.date}-${row.startTime}-${index}`}>
          <td>{row.projectName}</td>
          <td>{row.taskName}</td>
          <td>{row.date}</td>
          <td>{row.startTime}</td>
          <td>{row.endTime}</td>
          <td>{displayDuration(row.duration, showDecimal, language)}</td>
          <td className="report-note-cell">{row.description}</td>
        </tr>
      ))}
      <tr>
        <td colSpan={5}></td>
        <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
        <td></td>
      </tr>
    </ReportTable>
  );
}

export function GroupedByProjectTable({ language, report, showDecimal, t }: ReportTableProps<guiapp.WorktimeGroupedByProjectReport>) {
  const groups = report.groups ?? [];
  if (report.empty || !groups.length) return <p className="projects-empty">{t.emptyReport}</p>;
  return (
    <>
      {groups.map((group) => (
        <section className="report-section" key={group.projectId}>
          <h2>{group.projectName}</h2>
          <ReportTable headers={[t.columns.task, t.columns.date, t.columns.start, t.columns.end, t.columns.duration, t.columns.description]}>
            {(group.rows ?? []).map((row, index) => (
              <tr key={`${row.date}-${row.startTime}-${index}`}>
                <td>{row.taskName}</td>
                <td>{row.date}</td>
                <td>{row.startTime}</td>
                <td>{row.endTime}</td>
                <td>{displayDuration(row.duration, showDecimal, language)}</td>
                <td className="report-note-cell">{row.description}</td>
              </tr>
            ))}
            <tr>
              <td colSpan={4}></td>
              <td><strong>{displayDuration(group.totalDuration, showDecimal, language)}</strong></td>
              <td></td>
            </tr>
          </ReportTable>
        </section>
      ))}
    </>
  );
}

export function ProjectDetailsTable({ language, report, showDecimal, t }: ReportTableProps<guiapp.WorktimeProjectDetailsReport>) {
  const rows = report.rows ?? [];
  if (report.empty || !rows.length) return <p className="projects-empty">{t.emptyReport}</p>;
  return (
    <ReportTable headers={[t.columns.task, t.columns.date, t.columns.start, t.columns.end, t.columns.duration, t.columns.description]}>
      {rows.map((row, index) => (
        <tr key={`${row.date}-${row.startTime}-${index}`}>
          <td>{row.taskName}</td>
          <td>{row.date}</td>
          <td>{row.startTime}</td>
          <td>{row.endTime}</td>
          <td>{displayDuration(row.duration, showDecimal, language)}</td>
          <td className="report-note-cell">{row.description}</td>
        </tr>
      ))}
      <tr>
        <td colSpan={4}></td>
        <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
        <td></td>
      </tr>
    </ReportTable>
  );
}

export function TaskDetailsTable({ language, report, showDecimal, t }: ReportTableProps<guiapp.WorktimeTaskDetailsReport>) {
  const rows = report.rows ?? [];
  if (report.empty || !rows.length) return <p className="projects-empty">{t.emptyReport}</p>;
  return (
    <ReportTable headers={[t.columns.project, t.columns.task, t.columns.duration]}>
      {rows.map((row) => (
        <tr key={`${row.projectId}-${row.taskId}`}>
          <td>{row.projectName}</td>
          <td>{row.taskName}</td>
          <td>{displayDuration(row.duration, showDecimal, language)}</td>
        </tr>
      ))}
      <tr>
        <td colSpan={2}></td>
        <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
      </tr>
    </ReportTable>
  );
}

export function TimesheetTable({ language, report, showDecimal, t }: ReportTableProps<guiapp.TimesheetReport>) {
  if (report.empty) return <p className="projects-empty">{t.emptyReport}</p>;
  const dailyRows = report.dailyRows ?? [];
  const projectRows = report.projectRows ?? [];
  if (dailyRows.length) {
    return (
      <ReportTable headers={[t.columns.date, t.columns.total, t.columns.projectTime]}>
        {dailyRows.map((row) => (
          <tr key={row.date}>
            <td>{row.date}</td>
            <td>{displayDuration(row.totalDuration, showDecimal, language)}</td>
            <td>{displayDuration(row.projectDuration, showDecimal, language)}</td>
          </tr>
        ))}
        <tr>
          <td></td>
          <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
          <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
        </tr>
      </ReportTable>
    );
  }
  return (
    <section className="report-section">
      <h2>{report.userName}</h2>
      <ReportTable headers={[t.columns.project, t.columns.duration]}>
        {projectRows.map((row) => (
          <tr key={row.projectId}>
            <td>{row.projectName}</td>
            <td>{displayDuration(row.duration, showDecimal, language)}</td>
          </tr>
        ))}
        <tr>
          <td></td>
          <td><strong>{displayDuration(report.totalDuration, showDecimal, language)}</strong></td>
        </tr>
      </ReportTable>
    </section>
  );
}

function ReportTable({ headers, children }: { headers: string[]; children: ReactNode }) {
  return (
    <table className="report-table">
      <thead>
        <tr>
          {headers.map((header) => (
            <th key={header}>{header}</th>
          ))}
        </tr>
      </thead>
      <tbody>{children}</tbody>
    </table>
  );
}
