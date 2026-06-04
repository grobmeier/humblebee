import { formatDisplayDate, type DateLanguage } from "./dateFormat";

type WorkItem = { id: number; name: string; depth: number };

type Stopwatch = {
  durationSeconds: number;
  endDate: string;
  endTime: string;
  id: number;
  conflicting: boolean;
  running: boolean;
  startDate: string;
  startTime: string;
  workItemId?: number;
  workItemName: string;
};

type StopwatchSidebarProps = {
  selectedWorkItemId: number;
  stopwatches: Stopwatch[];
  language: DateLanguage;
  nowTimestamp: number;
  workItems: WorkItem[];
  onBookStopwatch: (stopwatch: Stopwatch) => void;
  onSelectWorkItem: (workItemId: number) => void;
  onDiscardStopwatch: (stopwatchId: number) => void;
  onStart: (workItemId?: number) => void;
  onStop: () => void;
  t: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    start: string;
    stopStopwatch: string;
  };
};

export function StopwatchSidebar({
  selectedWorkItemId,
  stopwatches,
  language,
  nowTimestamp,
  workItems,
  onBookStopwatch,
  onSelectWorkItem,
  onDiscardStopwatch,
  onStart,
  onStop,
  t
}: StopwatchSidebarProps) {
  const openWorkItemIds = new Set(stopwatches.map((stopwatch) => stopwatch.workItemId ?? 0));
  const availableWorkItems = workItems.filter(
    (workItem) => workItem.name.toLowerCase() !== "default" && !openWorkItemIds.has(workItem.id)
  );
  const selectedWorkItemAvailable = selectedWorkItemId === 0 || availableWorkItems.some((workItem) => workItem.id === selectedWorkItemId);
  const selectedValue = selectedWorkItemAvailable ? selectedWorkItemId : 0;

  return (
    <aside className="stopwatch-panel">
      <div className="stopwatch-create">
        <h2>{t.createStopwatch}</h2>
        <select value={selectedValue} onChange={(event) => onSelectWorkItem(Number(event.target.value))} aria-label="Stoppuhr Work item">
          <option value={0}></option>
          {availableWorkItems.map((workItem) => (
            <option key={workItem.id} value={workItem.id}>
              {"- ".repeat(Math.max(0, workItem.depth))}
              {workItem.name}
            </option>
          ))}
        </select>
        <div className="timer-actions">
          <button className="primary-button" disabled={selectedValue === 0} onClick={() => onStart(selectedValue)} type="button">
            {t.start}
          </button>
        </div>
      </div>

      {stopwatches.map((stopwatch) => (
        <div
          className={`timer-card ${stopwatch.running ? "active" : "book"} ${stopwatch.conflicting ? "conflict" : ""}`}
          key={stopwatch.id}
          style={{ borderLeftColor: stopwatch.conflicting ? "#d77" : "#5bb75b" }}
        >
          <div className="timer-card-dismiss-row">
            <button className="discard-button" type="button" onClick={() => onDiscardStopwatch(stopwatch.id)} aria-label={t.discardRunning} title={t.discardRunning}>
              ×
            </button>
          </div>
          <div className="timer-card-main">
            <div>
              <strong>{stopwatch.workItemName}</strong>
              {!stopwatch.running ? <span>{formatDisplayDate(stopwatch.startDate, language)}</span> : null}
            </div>
          </div>
          <div className="timer-card-times">
            <span>{stopwatch.startTime}</span>
            {!stopwatch.running ? <span>{stopwatch.endTime}</span> : null}
            <span>{formatStopwatchDuration(stopwatch, nowTimestamp)}</span>
          </div>
          <div className="timer-card-actions">
            {stopwatch.conflicting ? (
              <button className="book-button" type="button" onClick={() => onBookStopwatch(stopwatch)}>
                {t.book}
              </button>
            ) : null}
            {stopwatch.running ? (
              <button className="stop-button" type="button" onClick={onStop} aria-label={t.stopStopwatch} title={t.stopStopwatch}>
                {t.stopStopwatch}
              </button>
            ) : (
              <button className="play-button" type="button" onClick={() => onStart(stopwatch.workItemId ?? 0)} aria-label={t.start}>
                {t.start}
              </button>
            )}
          </div>
        </div>
      ))}
    </aside>
  );
}

function formatStopwatchDuration(stopwatch: Stopwatch, nowTimestamp: number): string {
  if (stopwatch.running) {
    const start = new Date(`${stopwatch.startDate}T${stopwatch.startTime}:00`).getTime();
    if (!Number.isNaN(start)) {
      return formatSeconds((nowTimestamp - start) / 1000);
    }
  }
  return formatSeconds(stopwatch.durationSeconds);
}

function formatSeconds(total: number): string {
  const seconds = Math.max(0, Math.floor(total));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours === 0) return `${minutes}m`;
  return `${hours}h ${String(minutes).padStart(2, "0")}m`;
}
