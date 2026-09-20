import { useEffect, useRef, useState } from "react";
import { GetWorkspacePreferences, SaveReportPreferences } from "../../wailsjs/go/guiapp/App";
import { defaultReportFilter } from "../reports/reportUtils";
import type { ReportFilter, ReportSlug, WorkItem } from "../reports/reportTypes";
import { restoreReportPreferences } from "./reportPreferences";

export function useReportPreferences(workspaceKey: string, activeReport: ReportSlug, projects: WorkItem[]) {
  const [filter, setFilter] = useState<ReportFilter>(defaultReportFilter);
  const [showDecimal, setShowDecimal] = useState(false);
  const [readyKey, setReadyKey] = useState<string | null>(null);
  const projectRef = useRef(projects);
  projectRef.current = projects;
  const pendingReport = useRef<ReportSlug | null>(null);

  useEffect(() => {
    let cancelled = false;
    const initialHash = window.location.hash;
    setReadyKey(null);
    setFilter(defaultReportFilter());
    setShowDecimal(false);
    pendingReport.current = null;
    if (!workspaceKey) return;
    GetWorkspacePreferences().then((preferences) => {
      if (cancelled || preferences.workspaceKey !== workspaceKey) return;
      const restored = restoreReportPreferences(preferences.reports, projectRef.current);
      if (restored) {
        const { report, showDecimal, ...restoredFilter } = restored;
        setFilter(restoredFilter);
        setShowDecimal(showDecimal);
        // Explicit report links and navigation during loading take precedence.
        if (window.location.hash === initialHash && (initialHash === "#reports" || initialHash === "#reports/")) {
          pendingReport.current = report;
          window.location.hash = `reports/${report}`;
        }
      }
      setReadyKey(workspaceKey);
    }).catch(() => {
      if (!cancelled) setReadyKey(workspaceKey);
    });
    return () => { cancelled = true; };
  }, [workspaceKey]);

  useEffect(() => {
    if (readyKey !== workspaceKey || !workspaceKey) return;
    if (pendingReport.current && pendingReport.current !== activeReport) return;
    pendingReport.current = null;
    const value = restoreReportPreferences({ ...filter, report: activeReport, showDecimal }, projects);
    if (!value) return;
    if (value.projectId !== filter.projectId) {
      setFilter((current) => ({ ...current, projectId: value.projectId }));
      return;
    }
    const timer = window.setTimeout(() => {
      void SaveReportPreferences(workspaceKey, value).catch(() => { /* Preferences never block booking. */ });
    }, 350);
    return () => window.clearTimeout(timer);
  }, [activeReport, filter, showDecimal, readyKey, workspaceKey, projects]);

  return { filter, setFilter, showDecimal, setShowDecimal, ready: !workspaceKey || readyKey === workspaceKey };
}
