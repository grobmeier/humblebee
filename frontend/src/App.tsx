import { FormEvent, useEffect, useState } from "react";
import "./app.css";
import {
  CreateTimeEntry,
  DeleteTimeEntry,
  DiscardStopwatch,
  GetDashboard,
  GetTimeDay,
  Init,
  ListStopwatches,
  ListWorkItems,
  Start,
  Stop,
  UpdateTimeEntry
} from "../wailsjs/go/guiapp/App";
import { Quit } from "../wailsjs/runtime/runtime";
import type { guiapp } from "../wailsjs/go/models";
import { DashboardCalendar } from "./dashboard/DashboardCalendar";
import { DashboardSummary } from "./dashboard/DashboardSummary";
import { StopwatchSidebar } from "./dashboard/StopwatchSidebar";
import { TimeEntriesEmptyState } from "./dashboard/TimeEntriesEmptyState";
import { atLocalNoon } from "./dashboard/calendarUtils";
import { formatInputDate, formatTime } from "./dashboard/dateFormat";
import { TimeEntryModal } from "./dashboard/TimeEntryModal";
import type { TimeEntryFormState } from "./dashboard/timeEntryTypes";
import { type Language, translations } from "./dashboard/translations";

type Dashboard = {
  initialized: boolean;
  dbPath: string;
  userEmail: string;
  running: null | { workItemName: string; startTimeUTC: number };
  todayTotalSeconds: number;
};

type WorkItem = { id: number; name: string; parentId?: number | null; depth: number };

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
  const [email, setEmail] = useState("");
  const [selectedWorkItemId, setSelectedWorkItemId] = useState<number>(0);
  const [error, setError] = useState<string>("");
  const [databaseBusyError, setDatabaseBusyError] = useState<DatabaseBusyError | null>(null);
  const [nowTimestamp, setNowTimestamp] = useState(() => Date.now());
  const [selectedDate, setSelectedDate] = useState(() => atLocalNoon(new Date()));
  const [timeDay, setTimeDay] = useState<guiapp.TimeDay | null>(null);
  const [expandedNoteIds, setExpandedNoteIds] = useState<number[]>([]);
  const [stopwatches, setStopwatches] = useState<guiapp.Stopwatch[]>([]);
  const [timeEntryForm, setTimeEntryForm] = useState<TimeEntryFormState>(() => createTimeEntryForm(atLocalNoon(new Date()), 0));
  const [timeEntryModalError, setTimeEntryModalError] = useState<string | null>(null);
  const [isTimeEntryModalOpen, setIsTimeEntryModalOpen] = useState(false);
  const [isSavingTimeEntry, setIsSavingTimeEntry] = useState(false);
  const [isStopwatchConfirmationModal, setIsStopwatchConfirmationModal] = useState(false);
  const [confirmationStopwatchId, setConfirmationStopwatchId] = useState<number | null>(null);
  const [language, setLanguage] = useState<Language>("de");
  const t = translations[language];

  useEffect(() => {
    document.documentElement.lang = language === "de" ? "de" : "en";
  }, [language]);

  async function refresh() {
    setError("");
    setDatabaseBusyError(null);
    const d = await GetDashboard();
    setDashboard(d);
    if (d.initialized) {
      const items = await ListWorkItems();
      setWorkItems(items);
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
    refreshTimeDay(selectedDate).catch(handleError);
  }, [dashboard?.initialized, selectedDate]);

  async function refreshTimeDay(date: Date) {
    const day = await GetTimeDay(formatInputDate(date));
    setTimeDay(day);
  }

  async function refreshStopwatches() {
    const rows = await ListStopwatches();
    setStopwatches(rows);
  }

  async function onInit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    if (!email.trim()) {
      setError("Enter the email address you want to use for this local workspace.");
      return;
    }
    try {
      await Init(email.trim());
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
      await refreshTimeDay(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      const stopwatchOverlap = parseStopwatchOverlapError(e);
      if (stopwatchOverlap) {
        await refresh();
        await refreshTimeDay(selectedDate);
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
      await refreshTimeDay(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      const stopwatchOverlap = parseStopwatchOverlapError(e);
      if (stopwatchOverlap) {
        await refresh();
        await refreshTimeDay(selectedDate);
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
      await refreshTimeDay(selectedDate);
      await refreshStopwatches();
    } catch (e) {
      setStopwatches(previousStopwatches);
      handleError(e);
    }
  }

  function onAddEntry(date = selectedDate) {
    setError("");
    setTimeEntryModalError(null);
    setIsStopwatchConfirmationModal(false);
    setConfirmationStopwatchId(null);
    setTimeEntryForm(createTimeEntryForm(date, selectedWorkItemId));
    setIsTimeEntryModalOpen(true);
  }

  function onEditEntry(entry: guiapp.TimeEntry) {
    const projectId = entry.workItemId ?? selectedWorkItemId;
    setError("");
    setTimeEntryModalError(null);
    setIsStopwatchConfirmationModal(false);
    setConfirmationStopwatchId(null);
    setTimeEntryForm({
      description: entry.description,
      endDate: entry.endDate,
      endTime: entry.endTime,
      id: entry.id,
      projectId,
      startDate: entry.startDate,
      startTime: entry.startTime,
      taskId: projectId,
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
      await refreshTimeDay(selectedDate);
      await refresh();
    } catch (e) {
      setTimeDay(previousTimeDay);
      handleError(e);
    }
  }

  function onToggleEntryNote(entryId: number) {
    setExpandedNoteIds((ids) => (ids.includes(entryId) ? ids.filter((id) => id !== entryId) : [...ids, entryId]));
  }

  function openStopwatchConfirmationModal(error: StopwatchOverlapError) {
    const projectId = error.workItemId || selectedWorkItemId;
    setError("");
    setTimeEntryModalError(
      "Die Stoppuhr ueberschneidet sich mit bereits gebuchter Zeit. Passe den Zeitraum an und speichere den Eintrag."
    );
    setTimeEntryForm({
      description: "",
      endDate: error.endDate,
      endTime: error.endTime,
      id: undefined,
      projectId,
      startDate: error.startDate,
      startTime: error.startTime,
      taskId: projectId,
      untilMidnight: false
    });
    setSelectedDate(atLocalNoon(new Date(`${error.startDate}T12:00:00`)));
    setConfirmationStopwatchId(error.stopwatchId || null);
    setIsStopwatchConfirmationModal(true);
    setIsTimeEntryModalOpen(true);
  }

  function onBookStopwatch(stopwatch: guiapp.Stopwatch) {
    const projectId = stopwatch.workItemId ?? selectedWorkItemId;
    setError("");
    setTimeEntryModalError(
      "Die Stoppuhr ueberschneidet sich mit bereits gebuchter Zeit. Passe den Zeitraum an und speichere den Eintrag."
    );
    setTimeEntryForm({
      description: "",
      endDate: stopwatch.endDate,
      endTime: stopwatch.endTime,
      id: undefined,
      projectId,
      startDate: stopwatch.startDate,
      startTime: stopwatch.startTime,
      taskId: projectId,
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
      setTimeEntryModalError("Bitte waehle ein Projekt aus.");
      return;
    }

    setIsSavingTimeEntry(true);
    try {
      const savedDate = atLocalNoon(new Date(`${timeEntryForm.startDate}T12:00:00`));
      const payload = {
        id: timeEntryForm.id ?? 0,
        workItemId: timeEntryForm.projectId,
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
      setSelectedWorkItemId(timeEntryForm.projectId);
      setSelectedDate(savedDate);
      await refreshTimeDay(savedDate);
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
          <div className="onboarding-progress" aria-label="Onboarding progress">
            <span className="is-active"></span>
            <span></span>
            <span></span>
          </div>
          <p className="eyebrow">Local-first time tracking</p>
          <form className="onboarding-form" onSubmit={onInit}>
            <h1>Set up your local workspace.</h1>
            <p>
              Humblebee keeps your time entries on this computer. Use the email address that should identify
              this local profile.
            </p>
            <label htmlFor="setup-email">Email</label>
            <input
              id="setup-email"
              autoFocus
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              type="email"
            />
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" type="submit">
              Create workspace
            </button>
          </form>
          <div className="local-note">
            <span>Database</span>
            <code>{dashboard.dbPath}</code>
          </div>
        </div>
      </div>
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand-mark" aria-hidden="true">
          ↻
        </div>
        <nav className="primary-nav" aria-label="Primary">
          <a className="selected" href="#dashboard">{t.nav.dashboard}</a>
          <a href="#reports">{t.nav.reports}</a>
          <a href="#projects">{t.nav.projects}</a>
        </nav>
        <div className="user-meta">
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
        <section className="dashboard-page" id="dashboard">
          <div className="dashboard-grid">
            <section className="main-panel">
              <DashboardCalendar
                language={language}
                selectedDate={selectedDate}
                onAddEntry={onAddEntry}
                onSelectDate={(date) => setSelectedDate(atLocalNoon(date))}
              />

              <DashboardSummary
                projectTime={formatHoursMinutes(timeDay?.projectSeconds ?? 0)}
                workTime={formatHoursMinutes(timeDay?.workSeconds ?? 0)}
              />
              <TimeEntriesEmptyState
                entries={timeDay?.entries ?? []}
                expandedNoteIds={expandedNoteIds}
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
      </div>
      {isTimeEntryModalOpen ? (
        <TimeEntryModal
          error={timeEntryModalError}
          form={timeEntryForm}
          isSaving={isSavingTimeEntry}
          language={language}
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
    </main>
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

function createTimeEntryForm(date: Date, projectId: number): TimeEntryFormState {
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
    taskId: projectId,
    untilMidnight: false
  };
}
