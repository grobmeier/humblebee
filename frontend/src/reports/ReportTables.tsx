import type { ReactNode } from "react";
import type { guiapp } from "../../wailsjs/go/models";
import { formatDecimalDuration } from "./reportUtils";

export function WorktimeByMonthTable({ report }: { report: guiapp.WorktimeByMonthReport }) {
  const rows = report.rows ?? [];
  if (report.empty || !rows.length) return <p className="projects-empty">No report data for this period.</p>;
  return (
    <ReportTable headers={["Project", "Task", "Date", "Start", "End", "Duration", "Description"]}>
      {rows.map((row, index) => (
        <tr key={`${row.date}-${row.startTime}-${index}`}>
          <td>{row.projectName}</td>
          <td>{row.taskName}</td>
          <td>{row.date}</td>
          <td>{row.startTime}</td>
          <td>{row.endTime}</td>
          <td>{row.duration}</td>
          <td>{row.description}</td>
        </tr>
      ))}
      <tr>
        <td colSpan={5}></td>
        <td><strong>{report.totalDuration}</strong></td>
        <td></td>
      </tr>
    </ReportTable>
  );
}

export function GroupedByProjectTable({ report }: { report: guiapp.WorktimeGroupedByProjectReport }) {
  const groups = report.groups ?? [];
  if (report.empty || !groups.length) return <p className="projects-empty">No report data for this period.</p>;
  return (
    <>
      {groups.map((group) => (
        <section className="report-section" key={group.projectId}>
          <h2>{group.projectName}</h2>
          <ReportTable headers={["Task", "Date", "Start", "End", "Duration", "Description"]}>
            {(group.rows ?? []).map((row, index) => (
              <tr key={`${row.date}-${row.startTime}-${index}`}>
                <td>{row.taskName}</td>
                <td>{row.date}</td>
                <td>{row.startTime}</td>
                <td>{row.endTime}</td>
                <td>{row.duration}</td>
                <td>{row.description}</td>
              </tr>
            ))}
            <tr>
              <td colSpan={4}></td>
              <td><strong>{group.totalDuration}</strong></td>
              <td></td>
            </tr>
          </ReportTable>
        </section>
      ))}
    </>
  );
}

export function TaskDetailsTable({ report }: { report: guiapp.WorktimeTaskDetailsReport }) {
  const rows = report.rows ?? [];
  if (report.empty || !rows.length) return <p className="projects-empty">No report data for this period.</p>;
  return (
    <ReportTable headers={["Project", "Task", "Duration"]}>
      {rows.map((row) => (
        <tr key={`${row.projectId}-${row.taskId}`}>
          <td>{row.projectName}</td>
          <td>{row.taskName}</td>
          <td>{row.duration}</td>
        </tr>
      ))}
      <tr>
        <td colSpan={2}></td>
        <td><strong>{report.totalDuration}</strong></td>
      </tr>
    </ReportTable>
  );
}

export function TimesheetTable({ report, showDecimal }: { report: guiapp.TimesheetReport; showDecimal: boolean }) {
  if (report.empty) return <p className="projects-empty">No report data for this period.</p>;
  const dailyRows = report.dailyRows ?? [];
  const projectRows = report.projectRows ?? [];
  if (dailyRows.length) {
    return (
      <ReportTable headers={["Date", "Total", "Project time"]}>
        {dailyRows.map((row) => (
          <tr key={row.date}>
            <td>{row.date}</td>
            <td>{showDecimal ? formatDecimalDuration(row.totalDuration) : row.totalDuration}</td>
            <td>{showDecimal ? formatDecimalDuration(row.projectDuration) : row.projectDuration}</td>
          </tr>
        ))}
        <tr>
          <td></td>
          <td><strong>{showDecimal ? formatDecimalDuration(report.totalDuration) : report.totalDuration}</strong></td>
          <td><strong>{showDecimal ? formatDecimalDuration(report.totalDuration) : report.totalDuration}</strong></td>
        </tr>
      </ReportTable>
    );
  }
  return (
    <section className="report-section">
      <h2>{report.userName}</h2>
      <ReportTable headers={["Project", "Duration"]}>
        {projectRows.map((row) => (
          <tr key={row.projectId}>
            <td>{row.projectName}</td>
            <td>{showDecimal ? formatDecimalDuration(row.duration) : row.duration}</td>
          </tr>
        ))}
        <tr>
          <td></td>
          <td><strong>{showDecimal ? formatDecimalDuration(report.totalDuration) : report.totalDuration}</strong></td>
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
