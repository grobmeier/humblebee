import { useEffect, useMemo, useState } from "react";
import "./app.css";
import { GetDashboard, Init, ListWorkItems, Start, Stop } from "../wailsjs/go/guiapp/App";

type Dashboard = {
  initialized: boolean;
  dbPath: string;
  userEmail: string;
  running: null | { workItemName: string; startTimeUTC: number };
  todayTotalSeconds: number;
};

type WorkItem = { id: number; name: string; parentId?: number | null; depth: number };

function formatSeconds(total: number): string {
  const seconds = Math.max(0, Math.floor(total));
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}

export default function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [workItems, setWorkItems] = useState<WorkItem[]>([]);
  const [email, setEmail] = useState("");
  const [selectedWorkItemId, setSelectedWorkItemId] = useState<number>(0);
  const [error, setError] = useState<string>("");

  const runningLabel = useMemo(() => {
    if (!dashboard?.running) return "Not running";
    const started = new Date(dashboard.running.startTimeUTC * 1000);
    return `${dashboard.running.workItemName} (started ${started.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false })})`;
  }, [dashboard]);

  async function refresh() {
    setError("");
    const d = await GetDashboard();
    setDashboard(d);
    if (d.initialized) {
      const items = await ListWorkItems();
      setWorkItems(items);
    }
  }

  useEffect(() => {
    refresh().catch((e) => setError(String(e)));
  }, []);

  async function onInit() {
    setError("");
    try {
      await Init(email);
      await refresh();
    } catch (e) {
      setError(String(e));
    }
  }

  async function onStart() {
    setError("");
    try {
      await Start(selectedWorkItemId);
      await refresh();
    } catch (e) {
      setError(String(e));
    }
  }

  async function onStop() {
    setError("");
    try {
      await Stop();
      await refresh();
    } catch (e) {
      setError(String(e));
    }
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
            <button className="secondary" onClick={() => refresh().catch((e) => setError(String(e)))}>
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

