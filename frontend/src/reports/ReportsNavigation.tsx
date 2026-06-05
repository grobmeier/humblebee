import { reportDefinitions, type ReportSlug } from "./reportTypes";

export function ReportsNavigation({ activeReport }: { activeReport: ReportSlug }) {
  return (
    <aside className="reports-list-panel" aria-label="Reports">
      <div className="projects-list-header">
        <h2>Reports</h2>
      </div>
      <div className="projects-list">
        {reportDefinitions.map((report) => (
          <a className={`project-list-item ${report.slug === activeReport ? "selected" : ""}`} href={`#reports/${report.slug}`} key={report.slug}>
            {report.title}
          </a>
        ))}
      </div>
    </aside>
  );
}
