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

import { useEffect, useMemo, useState } from "react";
import {
  ExportTimesheetReport,
  ExportWorktimeByMonthReport,
  ExportWorktimeGroupedByProjectReport,
  ExportWorktimeTaskDetailsReport,
  GetTimesheetReport,
  GetWorktimeByMonthReport,
  GetWorktimeGroupedByProjectReport,
  GetWorktimeTaskDetailsReport
} from "../../wailsjs/go/guiapp/App";
import type { guiapp } from "../../wailsjs/go/models";
import { translations, type Language } from "../dashboard/translations";
import { ReportFilterBar } from "./ReportFilterBar";
import { GroupedByProjectTable, TaskDetailsTable, TimesheetTable, WorktimeByMonthTable } from "./ReportTables";
import { ReportsNavigation } from "./ReportsNavigation";
import { defaultReportFilter, fileURL, toReportRequest } from "./reportUtils";
import { reportDefinitions, type ReportData, type ReportFilter, type ReportSlug, type ReportsPageText, type WorkItem } from "./reportTypes";

type ReportsPageProps = {
  activeReport: ReportSlug;
  language: Language;
  workItems: WorkItem[];
};

export function ReportsPage({ activeReport, language, workItems }: ReportsPageProps) {
  const [filter, setFilter] = useState<ReportFilter>(() => defaultReportFilter());
  const [data, setData] = useState<{ report: ReportSlug; value: ReportData } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [exportPath, setExportPath] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [showDecimal, setShowDecimal] = useState(false);
  const t = translations[language].reportsPage;
  const definition = reportDefinitions.find((report) => report.slug === activeReport) ?? reportDefinitions[0];
  const projectOptions = useMemo(
    () => workItems.filter((item) => item.parentId == null && item.name.toLowerCase() !== "default"),
    [workItems]
  );

  useEffect(() => {
    setExportPath(null);
    setError(null);
    setIsLoading(true);
    loadReport(activeReport, filter, language)
      .then((value) => setData({ report: activeReport, value }))
      .catch((err) => setError(String(err)))
      .finally(() => setIsLoading(false));
  }, [activeReport, filter, language]);

  async function exportReport() {
    setError(null);
    try {
      setExportPath(await exportActiveReport(activeReport, filter, language));
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <section className="reports-page" id="reports">
      <ReportsNavigation activeReport={activeReport} t={t} />
      <section className="report-detail-panel">
        <div className="project-detail-header">
          <div>
            <h1>{t.titles[definition.slug]}</h1>
          </div>
        </div>
        <ReportFilterBar
          filter={filter}
          language={language}
          needsProject={definition.needsProject}
          projectOptions={projectOptions}
          showDecimal={showDecimal}
          supportsDecimal={definition.decimalToggle}
          t={t}
          onChange={setFilter}
          onExport={() => void exportReport()}
          onPrint={() => window.print()}
          onToggleDecimal={() => setShowDecimal((value) => !value)}
        />
        {exportPath ? (
          <p className="report-export-path hide-print">
            {t.savedTo} <a href={fileURL(exportPath)}>{exportPath}</a>
          </p>
        ) : null}
        {error ? <div className="errors alert alert-error">{error}</div> : null}
        {isLoading ? <p className="projects-empty">{t.loadingReport}</p> : null}
        {!isLoading && data?.report === activeReport ? renderReport(activeReport, data.value, showDecimal, t) : null}
      </section>
    </section>
  );
}

async function loadReport(activeReport: ReportSlug, filter: ReportFilter, language: Language): Promise<ReportData> {
  const request = toReportRequest(filter, language);
  if (activeReport === "worktime-grouped-by-project") {
    return GetWorktimeGroupedByProjectReport(request);
  }
  if (activeReport === "worktime-task-details") {
    return GetWorktimeTaskDetailsReport(request);
  }
  if (activeReport === "timesheet") {
    return GetTimesheetReport(request);
  }
  return GetWorktimeByMonthReport(request);
}

async function exportActiveReport(activeReport: ReportSlug, filter: ReportFilter, language: Language): Promise<string> {
  const request = toReportRequest(filter, language);
  if (activeReport === "worktime-grouped-by-project") {
    return ExportWorktimeGroupedByProjectReport(request);
  }
  if (activeReport === "worktime-task-details") {
    return ExportWorktimeTaskDetailsReport(request);
  }
  if (activeReport === "timesheet") {
    return ExportTimesheetReport(request);
  }
  return ExportWorktimeByMonthReport(request);
}

function renderReport(activeReport: ReportSlug, data: ReportData, showDecimal: boolean, t: ReportsPageText) {
  if (activeReport === "worktime-grouped-by-project") {
    return <GroupedByProjectTable report={data as guiapp.WorktimeGroupedByProjectReport} t={t} />;
  }
  if (activeReport === "worktime-task-details") {
    return <TaskDetailsTable report={data as guiapp.WorktimeTaskDetailsReport} t={t} />;
  }
  if (activeReport === "timesheet") {
    return <TimesheetTable report={data as guiapp.TimesheetReport} showDecimal={showDecimal} t={t} />;
  }
  return <WorktimeByMonthTable report={data as guiapp.WorktimeByMonthReport} t={t} />;
}
