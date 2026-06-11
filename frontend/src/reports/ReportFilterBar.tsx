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

import { DashboardDateInput } from "../dashboard/TimeEntryModal";
import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import { monthOptions } from "./reportUtils";
import type { ReportFilter, ReportsPageText, WorkItem } from "./reportTypes";

type ReportFilterBarProps = {
  filter: ReportFilter;
  language: DateLanguage;
  needsProject: boolean;
  projectOptions: WorkItem[];
  projectPlaceholder: string;
  showDecimal: boolean;
  supportsDecimal: boolean;
  t: ReportsPageText;
  onChange: (filter: ReportFilter) => void;
  onExport: () => void;
  onPrint: () => void;
  onToggleDecimal: () => void;
};

export function ReportFilterBar({
  filter,
  language,
  needsProject,
  projectOptions,
  projectPlaceholder,
  showDecimal,
  supportsDecimal,
  t,
  onChange,
  onExport,
  onPrint,
  onToggleDecimal
}: ReportFilterBarProps) {
  return (
    <div className="report-filter-panel hide-print">
      <div className="report-filter-tabs" role="tablist" aria-label={t.filterMode}>
        <button className={filter.mode === "monthly" ? "active" : ""} type="button" onClick={() => onChange({ ...filter, mode: "monthly" })}>
          {t.monthly}
        </button>
        <button className={filter.mode === "daily" ? "active" : ""} type="button" onClick={() => onChange({ ...filter, mode: "daily" })}>
          {t.dateRange}
        </button>
      </div>
      <div className="report-filter-controls">
        {needsProject ? (
          <select className="tab-form-control tab-form-control--small" value={filter.projectId} onChange={(event) => onChange({ ...filter, projectId: Number(event.target.value) })}>
            <option value={0}>{projectPlaceholder}</option>
            {projectOptions.map((project) => (
              <option key={project.id} value={project.id}>
                {labelWorkItemName(project.name, language)}
              </option>
            ))}
          </select>
        ) : null}
        {filter.mode === "monthly" ? (
          <>
            <select className="tab-form-control tab-form-control--small" value={filter.month} onChange={(event) => onChange({ ...filter, month: Number(event.target.value) })}>
              {monthOptions(t.months).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
            <input className="tab-form-control tab-form-control--small" type="number" value={filter.year} onChange={(event) => onChange({ ...filter, year: Number(event.target.value) })} />
          </>
        ) : (
          <>
            <DashboardDateInput
              className="tab-form-control tab-form-control--small"
              language={language}
              value={filter.startDate}
              onChange={(value) => onChange({ ...filter, startDate: value })}
            />
            <span className="report-filter-separator">-</span>
            <DashboardDateInput
              className="tab-form-control tab-form-control--small"
              language={language}
              value={filter.endDate}
              onChange={(value) => onChange({ ...filter, endDate: value })}
            />
          </>
        )}
        <button className="secondary-button" type="button" onClick={onExport}>
          {t.exportExcel}
        </button>
        <button className="secondary-button" type="button" onClick={onPrint}>
          {t.print}
        </button>
        {supportsDecimal ? (
          <button className="secondary-button" type="button" onClick={onToggleDecimal}>
            {showDecimal ? "0:00" : "0.00"}
          </button>
        ) : null}
      </div>
    </div>
  );
}
