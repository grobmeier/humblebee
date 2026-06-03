import { useEffect, useMemo, useState } from "react";
import "./app.css";
import { GetDashboard, Init, ListWorkItems, Start, Stop } from "../wailsjs/go/guiapp/App";
import { Quit } from "../wailsjs/runtime/runtime";

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

function formatSeconds(total: number): string {
  const seconds = Math.max(0, Math.floor(total));
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
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

export default function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [workItems, setWorkItems] = useState<WorkItem[]>([]);
  const [email, setEmail] = useState("");
  const [selectedWorkItemId, setSelectedWorkItemId] = useState<number>(0);
  const [error, setError] = useState<string>("");
  const [databaseBusyError, setDatabaseBusyError] = useState<DatabaseBusyError | null>(null);

  const runningLabel = useMemo(() => {
    if (!dashboard?.running) return "Not running";
    const started = new Date(dashboard.running.startTimeUTC * 1000);
    return `${dashboard.running.workItemName} (started ${started.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })})`;
  }, [dashboard]);

  async function refresh() {
    setError("");
    setDatabaseBusyError(null);
    const d = await GetDashboard();
    setDashboard(d);
    if (d.initialized) {
      const items = await ListWorkItems();
      setWorkItems(items);
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

  async function onInit() {
    setError("");
    try {
      await Init(email);
      await refresh();
    } catch (e) {
      handleError(e);
    }
  }

  async function onStart() {
    setError("");
    try {
      await Start(selectedWorkItemId);
      await refresh();
    } catch (e) {
      handleError(e);
    }
  }

  async function onStop() {
    setError("");
    try {
      await Stop();
      await refresh();
    } catch (e) {
      handleError(e);
    }
  }

  if (databaseBusyError) {
    return (
      <div className="container recovery-screen">
        <div className="card recovery-card">
          <p className="eyebrow">Local database</p>
          <h1>Database is in use</h1>
          <p>
            HumbleBee cannot access the local database right now. Another HumbleBee window, terminal command,
            backup, or sync tool may still be using it.
          </p>
          <div className="path-box">
            <span>Database</span>
            <code>{databaseBusyError.dbPath}</code>
          </div>
          <p className="muted">
            Close other HumbleBee windows or wait for the other process to finish, then retry.
          </p>
          <details className="technical-details">
            <summary>Technical details</summary>
            <pre>{databaseBusyError.details}</pre>
          </details>
          <div className="row">
            <button onClick={() => refresh().catch(handleError)}>Retry</button>
            <button className="secondary" onClick={Quit}>
              Quit HumbleBee
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!dashboard) {
    return <div className="container">Loading…</div>;
  }

  if (!dashboard.initialized) {
    return (
      <div className="container">
        <h1>HumbleBee</h1>
        <p>First-time setup</p>
        <div className="card">
          <label>
            Email
            <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
          </label>
          <button onClick={onInit}>Initialize</button>
        </div>
        <p className="muted">Database: {dashboard.dbPath}</p>
        {error ? <pre className="error">{error}</pre> : null}
      </div>
    );
  }

  return (
    <div className="container">
      <header className="header">
        <div>
          <h1>HumbleBee</h1>
          <div className="muted">User: {dashboard.userEmail}</div>
        </div>
        <div className="stats">
          <div>
            <div className="statLabel">Today</div>
            <div className="statValue">{formatSeconds(dashboard.todayTotalSeconds)}</div>
          </div>
        </div>
      </header>

      <div className="grid">
        <div className="card">
          <h2>Timer</h2>
          <div className="muted">{runningLabel}</div>

          <label>
            Work item
            <select value={selectedWorkItemId} onChange={(e) => setSelectedWorkItemId(Number(e.target.value))}>
              <option value={0}>Default</option>
              {workItems
                .filter((w) => w.name.toLowerCase() !== "default")
                .map((w) => (
                  <option key={w.id} value={w.id}>
                    {" ".repeat(Math.max(0, w.depth) * 2)}
                    {w.name}
                  </option>
                ))}
            </select>
          </label>

          <div className="row">
            <button disabled={!!dashboard.running} onClick={onStart}>
              Start
            </button>
            <button disabled={!dashboard.running} onClick={onStop}>
              Stop
            </button>
            <button className="secondary" onClick={() => refresh().catch(handleError)}>
              Refresh
            </button>
          </div>
        </div>

        <div className="card">
          <h2>Next</h2>
          <ul className="list">
            <li>Day view + delete entries</li>
            <li>Work items management</li>
            <li>Monthly report</li>
            <li>Export CSV</li>
          </ul>
        </div>
      </div>

      <footer className="footer muted">Database: {dashboard.dbPath}</footer>

      {error ? <pre className="error">{error}</pre> : null}
    </div>
  );
}
