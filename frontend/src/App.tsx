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

import { FormEvent, useEffect, useState } from "react";
import "./app.css";
import {
  CreateDatabase,
  CreateTimeEntry,
  CreateProject,
  CreateProjectWithTasks,
  CreateTask,
  DeleteProject,
  DeleteTimeEntry,
  DiscardStopwatch,
  GetDashboard,
  GetDatabaseInfo,
  GetTimeDay,
  ImportTimeAndBill,
  Init,
  ListProjectWorkItems,
  ListStopwatches,
  ListWorkItems,
  PreviewTimeAndBillImport,
  SelectDatabaseFile,
  SelectImportFile,
  SelectNewDatabaseFile,
  SetProjectActive,
  SetTaskActive,
  Start,
  Stop,
  SwitchDatabase,
  UpdateProject,
  UpdateTimeEntry,
  UseDefaultDatabase
} from "../wailsjs/go/guiapp/App";
import { Quit } from "../wailsjs/runtime/runtime";
import type { guiapp } from "../wailsjs/go/models";
import { DashboardCalendar } from "./dashboard/DashboardCalendar";
import { DashboardSummary } from "./dashboard/DashboardSummary";
import { StopwatchSidebar } from "./dashboard/StopwatchSidebar";
import { TimeEntriesEmptyState } from "./dashboard/TimeEntriesEmptyState";
import { addDays, atLocalNoon } from "./dashboard/calendarUtils";
import { formatInputDate, formatTime } from "./dashboard/dateFormat";
import { TimeEntryModal } from "./dashboard/TimeEntryModal";
import type { TimeEntryFormState } from "./dashboard/timeEntryTypes";
import { type Language, translations } from "./dashboard/translations";
import { HumbleBeeLogo } from "./components/HumbleBeeLogo";
import { DatabaseSwitchIcon, ImportIcon } from "./components/AppIcons";
import { DatabaseSwitchModal } from "./database/DatabaseSwitchModal";
import { TimeAndBillImportModal } from "./importing/TimeAndBillImportModal";
import { ProjectsPage } from "./projects/ProjectsPage";
import { ReportsPage } from "./reports/ReportsPage";
import { reportSlugFromHash } from "./reports/reportUtils";
import type { ReportSlug } from "./reports/reportTypes";

type Dashboard = {
  initialized: boolean;
  dbPath: string;
  userEmail: string;
  running: null | { workItemName: string; startTimeUTC: number };
  todayTotalSeconds: number;
};

type WorkItem = { id: number; name: string; parentId?: number | null; depth: number; status?: string };

type DatabaseBusyError = {
  dbPath: string;
  details: string;
};

type StopwatchOverlapError = {
  stopwatchId: number;
  workItemId: number;
  startDate: string;
  startTime: string;
  endDate: string;
  endTime: string;
  details: string;
};

type DashboardSummaryTotals = {
  monthSeconds: number;
  weekSeconds: number;
};

type AppPage = "dashboard" | "reports" | "projects";

const localProfileEmail = "local@humblebee.local";

function formatHoursMinutes(total: number): string {
  const seconds = Math.max(0, Math.floor(total));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

function parseDatabaseBusyError(error: unknown): DatabaseBusyError | null {
  const message = String(error);
  if (!message.includes("HUMBLEBEE_DATABASE_BUSY")) {
    return null;
  }

  const dbPath = message.match(/Database:\s*(.+)/)?.[1]?.trim() ?? "Unknown";
  const details = message.match(/Details:\s*([\s\S]+)/)?.[1]?.trim() ?? message;
  return { dbPath, details };
}

function parseStopwatchOverlapError(error: unknown): StopwatchOverlapError | null {
  const message = String(error);
  if (!message.includes("HUMBLEBEE_STOPWATCH_OVERLAP")) {
    return null;
  }

  return {
    stopwatchId: Number(message.match(/StopwatchID:\s*(\d+)/)?.[1] ?? 0),
    workItemId: Number(message.match(/WorkItemID:\s*(\d+)/)?.[1] ?? 0),
    startDate: message.match(/StartDate:\s*(.+)/)?.[1]?.trim() ?? formatInputDate(new Date()),
    startTime: message.match(/StartTime:\s*(.+)/)?.[1]?.trim() ?? "09:00",
    endDate: message.match(/EndDate:\s*(.+)/)?.[1]?.trim() ?? formatInputDate(new Date()),
    endTime: message.match(/EndTime:\s*(.+)/)?.[1]?.trim() ?? "10:00",
    details: message.match(/Details:\s*([\s\S]+)/)?.[1]?.trim() ?? message
  };
}

export default function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [workItems, setWorkItems] = useState<WorkItem[]>([]);
  const [projectWorkItems, setProjectWorkItems] = useState<WorkItem[]>([]);
  const [selectedWorkItemId, setSelectedWorkItemId] = useState<number>(0);
  const [error, setError] = useState<string>("");
  const [databaseBusyError, setDatabaseBusyError] = useState<DatabaseBusyError | null>(null);
  const [nowTimestamp, setNowTimestamp] = useState(() => Date.now());
  const [selectedDate, setSelectedDate] = useState(() => atLocalNoon(new Date()));
  const [timeDay, setTimeDay] = useState<guiapp.TimeDay | null>(null);
  const [summaryTotals, setSummaryTotals] = useState<DashboardSummaryTotals>({ monthSeconds: 0, weekSeconds: 0 });
  const [expandedNoteIds, setExpandedNoteIds] = useState<number[]>([]);
  const [stopwatches, setStopwatches] = useState<guiapp.Stopwatch[]>([]);
  const [timeEntryForm, setTimeEntryForm] = useState<TimeEntryFormState>(() => createTimeEntryForm(atLocalNoon(new Date()), 0));
  const [timeEntryModalError, setTimeEntryModalError] = useState<string | null>(null);
  const [isTimeEntryModalOpen, setIsTimeEntryModalOpen] = useState(false);
  const [isSavingTimeEntry, setIsSavingTimeEntry] = useState(false);
  const [isStopwatchConfirmationModal, setIsStopwatchConfirmationModal] = useState(false);
  const [confirmationStopwatchId, setConfirmationStopwatchId] = useState<number | null>(null);
  const [language, setLanguage] = useState<Language>("de");
  const [activePage, setActivePage] = useState<AppPage>(() => pageFromHash(window.location.hash));
  const [activeReport, setActiveReport] = useState<ReportSlug>(() => reportSlugFromHash(window.location.hash));
  const [selectedProjectPageProjectId, setSelectedProjectPageProjectId] = useState<number>(0);
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [importFilePath, setImportFilePath] = useState("");
  const [importPreview, setImportPreview] = useState<guiapp.ImportPreview | null>(null);
  const [importResult, setImportResult] = useState<guiapp.ImportResult | null>(null);
  const [importModalError, setImportModalError] = useState<string | null>(null);
  const [isPreviewingImport, setIsPreviewingImport] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [isDatabaseModalOpen, setIsDatabaseModalOpen] = useState(false);
  const [databaseInfo, setDatabaseInfo] = useState<guiapp.DatabaseInfo | null>(null);
  const [databaseModalError, setDatabaseModalError] = useState<string | null>(null);
  const [isSwitchingDatabase, setIsSwitchingDatabase] = useState(false);
  const t = translations[language];

  useEffect(() => {
    document.documentElement.lang = language === "de" ? "de" : "en";
  }, [language]);

  useEffect(() => {
    function syncPageFromHash() {
      setActivePage(pageFromHash(window.location.hash));
      setActiveReport(reportSlugFromHash(window.location.hash));
    }

    window.addEventListener("hashchange", syncPageFromHash);
    return () => window.removeEventListener("hashchange", syncPageFromHash);
  }, []);

  async function refresh() {
    setError("");
    setDatabaseBusyError(null);
    const d = await GetDashboard();
    setDashboard(d);
    if (d.initialized) {
      const items = await refreshWorkItems();
      await refreshProjectWorkItems();
      if (!selectedWorkItemId && items.length) {
        setSelectedWorkItemId(0);
      }
    }
    if (d.initialized) {
      await refreshStopwatches();
    }
  }

  function handleError(error: unknown) {
    const busy = parseDatabaseBusyError(error);
    if (busy) {
      setDatabaseBusyError(busy);
      setError("");
      return;
    }
    setError(String(error));
  }

  useEffect(() => {
    refresh().catch(handleError);
  }, []);

  useEffect(() => {
    const intervalId = window.setInterval(() => setNowTimestamp(Date.now()), 1000);
    return () => window.clearInterval(intervalId);
  }, []);

  useEffect(() => {
    if (!dashboard?.initialized) {
      return;
    }
    refreshDashboardTime(selectedDate).catch(handleError);
  }, [dashboard?.initialized, selectedDate]);

  async function refreshDashboardTime(date: Date) {
    await refreshTimeDay(date);
    await refreshSummaryTotals(date);
  }

  async function refreshTimeDay(date: Date) {
    const day = await GetTimeDay(formatInputDate(date));
    setTimeDay(day);
  }

  async function refreshSummaryTotals(date: Date) {
    const [weekDays, monthDays] = await Promise.all([
      Promise.all(dateRange(startOfIsoWeek(date), addDays(startOfIsoWeek(date), 6)).map((day) => GetTimeDay(formatInputDate(day)))),
      Promise.all(dateRange(startOfMonth(date), endOfMonth(date)).map((day) => GetTimeDay(formatInputDate(day))))
    ]);
    setSummaryTotals({
      weekSeconds: weekDays.reduce((total, day) => total + day.workSeconds, 0),
      monthSeconds: monthDays.reduce((total, day) => total + day.workSeconds, 0)
    });
  }

  async function refreshStopwatches() {
    const rows = await ListStopwatches();
    setStopwatches(rows);
  }

  async function refreshWorkItems() {
    const items = await ListWorkItems();
    setWorkItems(items);
    return items;
  }

  async function refreshProjectWorkItems() {
    const items = await ListProjectWorkItems();
    setProjectWorkItems(items);
    return items;
  }

  function clearWorkspaceState() {
    setWorkItems([]);
    setProjectWorkItems([]);
    setStopwatches([]);
    setTimeDay(null);
    setSummaryTotals({ monthSeconds: 0, weekSeconds: 0 });
    setExpandedNoteIds([]);
    setSelectedWorkItemId(0);
    setSelectedProjectPageProjectId(0);
  }

  async function refreshAfterDatabaseChange() {
    clearWorkspaceState();
    setIsTimeEntryModalOpen(false);
    setConfirmationStopwatchId(null);
    await refresh();
  }

  async function onInit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      await Init(localProfileEmail);
      await refresh();
    } catch (e) {
      handleError(e);
    }
  }

  async function onStart(workItemId = selectedWorkItemId) {
    setError("");
    try {
      await Start(workItemId);
      setConfirmationStopwatchId(null);
      await refresh();
      await refreshDashboardTime(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      const stopwatchOverlap = parseStopwatchOverlapError(e);
      if (stopwatchOverlap) {
        await refresh();
        await refreshDashboardTime(selectedDate);
        await refreshStopwatches();
        openStopwatchConfirmationModal(stopwatchOverlap);
        return;
      }
      if (parseDatabaseBusyError(e)) {
        handleError(e);
        return;
      }
      setError(formatTimeEntryError(e));
    }
  }

  async function onStop() {
    setError("");
    try {
      await Stop();
      await refresh();
      await refreshDashboardTime(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      const stopwatchOverlap = parseStopwatchOverlapError(e);
      if (stopwatchOverlap) {
        await refresh();
        await refreshDashboardTime(selectedDate);
        await refreshStopwatches();
        openStopwatchConfirmationModal(stopwatchOverlap);
        return;
      }
      handleError(e);
    }
  }

  async function onDiscardStopwatch(stopwatchId: number) {
    setError("");
    const previousStopwatches = stopwatches;
    setStopwatches((rows) => rows.filter((stopwatch) => stopwatch.id !== stopwatchId));
    try {
      try {
        await DiscardStopwatch(stopwatchId);
      } catch {
        await DeleteTimeEntry(stopwatchId);
      }
      setConfirmationStopwatchId(null);
      await refresh();
      await refreshDashboardTime(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      setStopwatches(previousStopwatches);
      handleError(e);
    }
  }

  async function onCreateProject(name: string, sourceProjectId: number) {
    const project = sourceProjectId > 0 ? await CreateProjectWithTasks(name, sourceProjectId) : await CreateProject(name);
    await refreshWorkItems();
    await refreshProjectWorkItems();
    setSelectedProjectPageProjectId(project.id);
  }

  async function onUpdateProject(projectId: number, name: string) {
    const project = await UpdateProject(projectId, name);
    await refreshWorkItems();
    await refreshProjectWorkItems();
    await refreshStopwatches();
    setSelectedProjectPageProjectId(project.id);
  }

  async function onCreateTask(projectId: number, name: string) {
    await CreateTask(projectId, name);
    await refreshWorkItems();
    await refreshProjectWorkItems();
  }

  async function onDeleteProject(projectId: number) {
    setError("");
    await DeleteProject(projectId);
    if (selectedWorkItemId !== 0) {
      const selectedItem = projectWorkItems.find((item) => item.id === selectedWorkItemId);
      if (selectedItem?.id === projectId || selectedItem?.parentId === projectId) {
        setSelectedWorkItemId(0);
      }
    }
    const items = await refreshProjectWorkItems();
    await refreshWorkItems();
    await refreshDashboardTime(selectedDate);
    await refreshStopwatches();
    const nextProject = items.find((item) => item.parentId == null && item.name.toLowerCase() !== "default");
    setSelectedProjectPageProjectId(nextProject?.id ?? 0);
  }

  async function onSetProjectActive(projectId: number, active: boolean) {
    setError("");
    await SetProjectActive(projectId, active);
    if (!active && selectedWorkItemId !== 0) {
      const selectedItem = projectWorkItems.find((item) => item.id === selectedWorkItemId);
      if (selectedItem?.id === projectId || selectedItem?.parentId === projectId) {
        setSelectedWorkItemId(0);
      }
    }
    await refreshWorkItems();
    await refreshProjectWorkItems();
    await refreshStopwatches();
  }

  async function onSetTaskActive(taskId: number, active: boolean) {
    await SetTaskActive(taskId, active);
    await refreshWorkItems();
    await refreshProjectWorkItems();
    if (!active && selectedWorkItemId === taskId) {
      setSelectedWorkItemId(0);
    }
  }

  function openImportModal() {
    setImportFilePath("");
    setImportPreview(null);
    setImportResult(null);
    setImportModalError(null);
    setIsImportModalOpen(true);
  }

  async function chooseImportFile() {
    setImportModalError(null);
    try {
      const path = await SelectImportFile();
      if (path) {
        setImportFilePath(path);
        setImportPreview(null);
        setImportResult(null);
      }
    } catch (e) {
      setImportModalError(String(e));
    }
  }

  async function previewImport() {
    if (!importFilePath) {
      return;
    }
    setImportModalError(null);
    setIsPreviewingImport(true);
    try {
      setImportPreview(await PreviewTimeAndBillImport(importFilePath));
      setImportResult(null);
    } catch (e) {
      setImportModalError(String(e));
    } finally {
      setIsPreviewingImport(false);
    }
  }

  async function importTimeAndBill() {
    if (!importFilePath || !importPreview) {
      return;
    }
    setImportModalError(null);
    setIsImporting(true);
    try {
      setImportResult(await ImportTimeAndBill(importFilePath));
      await refresh();
      await refreshDashboardTime(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      setImportModalError(String(e));
    } finally {
      setIsImporting(false);
    }
  }

  async function openDatabaseModal() {
    setDatabaseModalError(null);
    try {
      const info = await GetDatabaseInfo();
      setDatabaseInfo(info);
      setIsDatabaseModalOpen(true);
    } catch (e) {
      handleError(e);
    }
  }

  async function openExistingDatabase() {
    setDatabaseModalError(null);
    setIsSwitchingDatabase(true);
    try {
      const path = await SelectDatabaseFile();
      if (!path) {
        return;
      }
      const info = await SwitchDatabase(path);
      setDatabaseInfo(info);
      setIsDatabaseModalOpen(false);
      await refreshAfterDatabaseChange();
    } catch (e) {
      setDatabaseModalError(String(e));
    } finally {
      setIsSwitchingDatabase(false);
    }
  }

  async function createNewDatabase() {
    setDatabaseModalError(null);
    setIsSwitchingDatabase(true);
    try {
      const path = await SelectNewDatabaseFile();
      if (!path) {
        return;
      }
      const info = await CreateDatabase(path);
      setDatabaseInfo(info);
      setIsDatabaseModalOpen(false);
      await refreshAfterDatabaseChange();
    } catch (e) {
      setDatabaseModalError(String(e));
    } finally {
      setIsSwitchingDatabase(false);
    }
  }

  async function useDefaultDatabase() {
    setDatabaseModalError(null);
    setIsSwitchingDatabase(true);
    try {
      const info = await UseDefaultDatabase();
      setDatabaseInfo(info);
      setIsDatabaseModalOpen(false);
      await refreshAfterDatabaseChange();
    } catch (e) {
      setDatabaseModalError(String(e));
    } finally {
      setIsSwitchingDatabase(false);
    }
  }

  function onAddEntry(date = selectedDate) {
    setError("");
    setTimeEntryModalError(null);
    setIsStopwatchConfirmationModal(false);
    setConfirmationStopwatchId(null);
    const selection = timeEntrySelectionForWorkItem(selectedWorkItemId);
    setTimeEntryForm(createTimeEntryForm(date, selection.projectId, selection.taskId));
    setIsTimeEntryModalOpen(true);
  }

  function onEditEntry(entry: guiapp.TimeEntry) {
    const selection = timeEntrySelectionForWorkItem(entry.workItemId ?? selectedWorkItemId);
    setError("");
    setTimeEntryModalError(null);
    setIsStopwatchConfirmationModal(false);
    setConfirmationStopwatchId(null);
    setTimeEntryForm({
      description: entry.description,
      endDate: entry.endDate,
      endTime: entry.endTime,
      id: entry.id,
      projectId: selection.projectId,
      startDate: entry.startDate,
      startTime: entry.startTime,
      taskId: selection.taskId,
      untilMidnight: false
    });
    setIsTimeEntryModalOpen(true);
  }

  async function onDeleteEntry(entry: guiapp.TimeEntry) {
    setError("");
    const previousTimeDay = timeDay;
    setTimeDay((day) => (day ? { ...day, entries: day.entries.filter((row) => row.id !== entry.id) } : day));
    try {
      await DeleteTimeEntry(entry.id);
      setExpandedNoteIds((ids) => ids.filter((id) => id !== entry.id));
      await refreshDashboardTime(selectedDate);
      await refresh();
    } catch (e) {
      setTimeDay(previousTimeDay);
      handleError(e);
    }
  }

  function onToggleEntryNote(entryId: number) {
    setExpandedNoteIds((ids) => (ids.includes(entryId) ? ids.filter((id) => id !== entryId) : [...ids, entryId]));
  }

  function timeEntrySelectionForWorkItem(workItemId: number): { projectId: number; taskId: number } {
    const workItem = workItems.find((item) => item.id === workItemId);
    if (!workItem) {
      return { projectId: 0, taskId: 0 };
    }
    if (workItem.parentId != null) {
      return { projectId: workItem.parentId, taskId: workItem.id };
    }
    const firstTask = workItems.find((item) => item.parentId === workItem.id);
    return { projectId: workItem.id, taskId: firstTask?.id ?? 0 };
  }

  function openStopwatchConfirmationModal(error: StopwatchOverlapError) {
    const selection = timeEntrySelectionForWorkItem(error.workItemId || selectedWorkItemId);
    setError("");
    setTimeEntryModalError(t.timeEntryModal.conflictMessage);
    setTimeEntryForm({
      description: "",
      endDate: error.endDate,
      endTime: error.endTime,
      id: undefined,
      projectId: selection.projectId,
      startDate: error.startDate,
      startTime: error.startTime,
      taskId: selection.taskId,
      untilMidnight: false
    });
    setSelectedDate(atLocalNoon(new Date(`${error.startDate}T12:00:00`)));
    setConfirmationStopwatchId(error.stopwatchId || null);
    setIsStopwatchConfirmationModal(true);
    setIsTimeEntryModalOpen(true);
  }

  function onBookStopwatch(stopwatch: guiapp.Stopwatch) {
    const selection = timeEntrySelectionForWorkItem(stopwatch.workItemId ?? selectedWorkItemId);
    setError("");
    setTimeEntryModalError(t.timeEntryModal.conflictMessage);
    setTimeEntryForm({
      description: "",
      endDate: stopwatch.endDate,
      endTime: stopwatch.endTime,
      id: undefined,
      projectId: selection.projectId,
      startDate: stopwatch.startDate,
      startTime: stopwatch.startTime,
      taskId: selection.taskId,
      untilMidnight: false
    });
    setSelectedDate(atLocalNoon(new Date(`${stopwatch.startDate}T12:00:00`)));
    setConfirmationStopwatchId(stopwatch.id);
    setIsStopwatchConfirmationModal(true);
    setIsTimeEntryModalOpen(true);
  }

  async function onSubmitTimeEntry(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setTimeEntryModalError(null);

    if (!timeEntryForm.projectId) {
      setTimeEntryModalError(t.timeEntryModal.selectProjectRequired);
      return;
    }
    if (!timeEntryForm.taskId) {
      setTimeEntryModalError(t.timeEntryModal.selectTaskRequired);
      return;
    }

    setIsSavingTimeEntry(true);
    try {
      const savedDate = atLocalNoon(new Date(`${timeEntryForm.startDate}T12:00:00`));
      const payload = {
        id: timeEntryForm.id ?? 0,
        workItemId: timeEntryForm.taskId,
        description: timeEntryForm.description,
        startDate: timeEntryForm.startDate,
        startTime: timeEntryForm.startTime,
        endDate: timeEntryForm.endDate,
        endTime: timeEntryForm.endTime,
        untilMidnight: timeEntryForm.untilMidnight
      };
      if (timeEntryForm.id) {
        await UpdateTimeEntry(payload);
      } else {
        await CreateTimeEntry(payload);
      }
      if (isStopwatchConfirmationModal && confirmationStopwatchId !== null) {
        await DiscardStopwatch(confirmationStopwatchId);
      }
      setSelectedWorkItemId(timeEntryForm.taskId);
      setSelectedDate(savedDate);
      await refreshDashboardTime(savedDate);
      await refresh();
      await refreshStopwatches();
      setConfirmationStopwatchId(null);
      setIsStopwatchConfirmationModal(false);
      setIsTimeEntryModalOpen(false);
    } catch (e) {
      if (parseDatabaseBusyError(e)) {
        handleError(e);
        return;
      }
      setTimeEntryModalError(formatTimeEntryError(e));
    } finally {
      setIsSavingTimeEntry(false);
    }
  }

  if (databaseBusyError) {
    return (
      <div className="recovery-screen">
        <section className="recovery-panel" aria-labelledby="database-busy-title">
          <p className="eyebrow">Local database</p>
          <h1 id="database-busy-title">Database is in use</h1>
          <p>
            HumbleBee cannot access the local database right now. Another HumbleBee window, terminal command,
            backup, or sync tool may still be using it.
          </p>
          <div className="path-box">
            <span>Database</span>
            <code>{databaseBusyError.dbPath}</code>
          </div>
          <p className="recovery-note">
            Close other HumbleBee windows or wait for the other process to finish, then retry.
          </p>
          <details className="technical-details">
            <summary>Technical details</summary>
            <pre>{databaseBusyError.details}</pre>
          </details>
          <div className="recovery-actions">
            <button className="primary-button" type="button" onClick={() => refresh().catch(handleError)}>
              Retry
            </button>
            <button className="secondary-button" type="button" onClick={Quit}>
              Quit HumbleBee
            </button>
          </div>
        </section>
      </div>
    );
  }

  if (!dashboard) {
    return <div className="loading-screen">Loading Humblebee...</div>;
  }

  if (!dashboard.initialized) {
    return (
      <div className="onboarding-screen">
        <div className="onboarding-brand">Humblebee</div>
        <div className="onboarding-panel">
          <p className="eyebrow">Local-first time tracking</p>
          <form className="onboarding-form" onSubmit={onInit}>
            <h1>Set up your local workspace.</h1>
            <p>Humblebee keeps your time entries local.</p>
            <div className="local-note local-note--setup">
              <span>You can find your database here:</span>
              <code>{dashboard.dbPath}</code>
            </div>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" type="submit">
              Create workspace
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand-mark" aria-hidden="true">
          <HumbleBeeLogo />
        </div>
        <nav className="primary-nav" aria-label="Primary">
          <a className={activePage === "dashboard" ? "selected" : ""} href="#dashboard">{t.nav.dashboard}</a>
          <a className={activePage === "reports" ? "selected" : ""} href="#reports">{t.nav.reports}</a>
          <a className={activePage === "projects" ? "selected" : ""} href="#projects">{t.nav.projects}</a>
        </nav>
        <div className="user-meta">
          <button className="icon-button" type="button" onClick={openImportModal} aria-label={t.importPage.importButton} title={t.importPage.importButton}>
            <ImportIcon />
          </button>
          <button className="icon-button" type="button" onClick={() => void openDatabaseModal()} aria-label={t.databasePage.switchButton} title={t.databasePage.switchButton}>
            <DatabaseSwitchIcon />
          </button>
          <div className="language-switch" aria-label="Language">
            <button className={language === "de" ? "active" : ""} type="button" onClick={() => setLanguage("de")}>
              DE
            </button>
            <button className={language === "en" ? "active" : ""} type="button" onClick={() => setLanguage("en")}>
              EN
            </button>
          </div>
        </div>
      </header>

      <div className="content">
        {activePage === "dashboard" ? (
          <section className="dashboard-page" id="dashboard">
            <div className="dashboard-grid">
              <section className="main-panel">
                <DashboardCalendar
                  language={language}
                  selectedDate={selectedDate}
                  t={t.dashboardCalendar}
                  onAddEntry={onAddEntry}
                  onSelectDate={(date) => setSelectedDate(atLocalNoon(date))}
                />

                <DashboardSummary
                  monthWorkTime={formatHoursMinutes(summaryTotals.monthSeconds)}
                  weekWorkTime={formatHoursMinutes(summaryTotals.weekSeconds)}
                />
                <TimeEntriesEmptyState
                  entries={timeDay?.entries ?? []}
                  expandedNoteIds={expandedNoteIds}
                  language={language}
                  workItems={workItems}
                  onDeleteEntry={onDeleteEntry}
                  onEditEntry={onEditEntry}
                  onToggleNote={onToggleEntryNote}
                />
              </section>

              <StopwatchSidebar
                selectedWorkItemId={selectedWorkItemId}
                stopwatches={stopwatches}
                language={language}
                nowTimestamp={nowTimestamp}
                workItems={workItems}
                onBookStopwatch={onBookStopwatch}
                onSelectWorkItem={setSelectedWorkItemId}
                onDiscardStopwatch={onDiscardStopwatch}
                onStart={onStart}
                onStop={onStop}
                t={t.stopwatch}
              />
            </div>

            {error ? <pre className="error-box">{error}</pre> : null}
          </section>
        ) : null}
        {activePage === "projects" ? (
          <ProjectsPage
            language={language}
            selectedProjectId={selectedProjectPageProjectId}
            t={t.projectsPage}
            workItems={projectWorkItems}
            onCreateProject={onCreateProject}
            onCreateTask={onCreateTask}
            onDeleteProject={onDeleteProject}
            onSelectProject={setSelectedProjectPageProjectId}
            onSetProjectActive={onSetProjectActive}
            onSetTaskActive={onSetTaskActive}
            onUpdateProject={onUpdateProject}
          />
        ) : null}
        {activePage === "reports" ? <ReportsPage activeReport={activeReport} language={language} workItems={projectWorkItems} /> : null}
      </div>
      {isTimeEntryModalOpen ? (
        <TimeEntryModal
          error={timeEntryModalError}
          form={timeEntryForm}
          isSaving={isSavingTimeEntry}
          language={language}
          t={t.timeEntryModal}
          workItems={workItems.filter((workItem) => workItem.name.toLowerCase() !== "default")}
          onChange={setTimeEntryForm}
          onClose={() => {
            setIsStopwatchConfirmationModal(false);
            setConfirmationStopwatchId(null);
            setIsTimeEntryModalOpen(false);
          }}
          onSubmit={onSubmitTimeEntry}
        />
      ) : null}
      {isImportModalOpen ? (
        <TimeAndBillImportModal
          error={importModalError}
          filePath={importFilePath}
          isImporting={isImporting}
          isPreviewing={isPreviewingImport}
          preview={importPreview}
          result={importResult}
          t={t.importPage}
          onChooseFile={chooseImportFile}
          onClose={() => setIsImportModalOpen(false)}
          onImport={() => void importTimeAndBill()}
          onPreview={() => void previewImport()}
        />
      ) : null}
      {isDatabaseModalOpen ? (
        <DatabaseSwitchModal
          currentDatabasePath={dashboard.dbPath}
          databaseInfo={databaseInfo}
          error={databaseModalError}
          isSaving={isSwitchingDatabase}
          t={t.databasePage}
          onCreateNew={() => void createNewDatabase()}
          onOpenExisting={() => void openExistingDatabase()}
          onClose={() => setIsDatabaseModalOpen(false)}
          onUseDefault={() => void useDefaultDatabase()}
        />
      ) : null}
    </main>
  );
}

function pageFromHash(hash: string): AppPage {
  if (hash === "#reports" || hash.startsWith("#reports/")) {
    return "reports";
  }
  if (hash === "#projects") {
    return "projects";
  }
  return "dashboard";
}

function PlaceholderPage({ page, text }: { page: "reports"; text: { eyebrow: string; title: string; body: string } }) {
  return (
    <section className="placeholder-page" id={page} aria-labelledby={`${page}-title`}>
      <p className="eyebrow">{text.eyebrow}</p>
      <h1 id={`${page}-title`}>{text.title}</h1>
      <p>{text.body}</p>
    </section>
  );
}

function formatTimeEntryError(error: unknown): string {
  const message = String(error);
  if (message.includes("overlaps")) {
    return "Der Zeiteintrag ueberschneidet sich mit einem bestehenden Eintrag.";
  }
  if (message.includes("end time must be after start time")) {
    return "Die Endzeit muss nach der Startzeit liegen.";
  }
  if (message.includes("invalid")) {
    return "Bitte gib eine gueltige Start- und Endzeit ein.";
  }
  return message;
}

function dateRange(start: Date, end: Date): Date[] {
  const days: Date[] = [];
  let current = atLocalNoon(start);
  const last = atLocalNoon(end);
  while (current.getTime() <= last.getTime()) {
    days.push(current);
    current = addDays(current, 1);
  }
  return days;
}

function startOfIsoWeek(date: Date): Date {
  const normalized = atLocalNoon(date);
  const weekday = (normalized.getDay() + 6) % 7;
  return addDays(normalized, -weekday);
}

function startOfMonth(date: Date): Date {
  return atLocalNoon(new Date(date.getFullYear(), date.getMonth(), 1));
}

function endOfMonth(date: Date): Date {
  return atLocalNoon(new Date(date.getFullYear(), date.getMonth() + 1, 0));
}

function createTimeEntryForm(date: Date, projectId: number, taskId = 0): TimeEntryFormState {
  const start = atLocalNoon(date);
  start.setHours(9, 0, 0, 0);
  const end = atLocalNoon(date);
  end.setHours(10, 0, 0, 0);
  const formattedDate = formatInputDate(date);

  return {
    description: "",
    endDate: formattedDate,
    endTime: formatTime(end),
    projectId,
    startDate: formattedDate,
    startTime: formatTime(start),
    taskId,
    untilMidnight: false
  };
}
