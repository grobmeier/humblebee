import { displayWorkItem, type WorkItemNode } from "./workItemUtils";
import type { DateLanguage } from "./dateFormat";

type TimeEntryRow = {
  description: string;
  durationSeconds: number;
  endDate: string;
  endTime: string;
  id: number;
  startDate: string;
  startTime: string;
  workItemId?: number;
};

type TimeEntriesEmptyStateProps = {
  expandedNoteIds: number[];
  entries: TimeEntryRow[];
  language: DateLanguage;
  onDeleteEntry: (entry: TimeEntryRow) => void;
  onEditEntry: (entry: TimeEntryRow) => void;
  onToggleNote: (entryId: number) => void;
  workItems: WorkItemNode[];
};

export function TimeEntriesEmptyState({ entries, expandedNoteIds, language, onDeleteEntry, onEditEntry, onToggleNote, workItems }: TimeEntriesEmptyStateProps) {
  return (
    <section className="entries-section">
      <h2>Zeiteintraege</h2>
      {entries.length ? (
        <div className="entries-list">
          {entries.map((entry) => {
            const display = displayWorkItem(entry.workItemId ?? 0, workItems, language);
            const hasNote = entry.description.trim().length > 0;
            const isNoteExpanded = expandedNoteIds.includes(entry.id);
            return (
              <div className="entry-row-wrap" key={entry.id}>
                <button className="entry-row" type="button" onClick={() => onEditEntry(entry)}>
                  <span>{entry.startTime} - {entry.endTime}</span>
                  <strong>
                    {display.projectName}
                    {display.taskName ? <small>{display.taskName}</small> : null}
                  </strong>
                  <em>{formatDuration(entry.durationSeconds)}</em>
                </button>
                <div className="entry-actions">
                  {hasNote ? (
                    <button
                      className="entry-icon-button has-note"
                      type="button"
                      onClick={() => onToggleNote(entry.id)}
                      aria-label={isNoteExpanded ? "Notiz ausblenden" : "Notiz anzeigen"}
                      title="Notiz anzeigen"
                    >
                      !
                    </button>
                  ) : null}
                  <button
                    className="entry-icon-button danger"
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onDeleteEntry(entry);
                    }}
                    aria-label="Zeiteintrag loeschen"
                    title="Zeiteintrag loeschen"
                  >
                    <svg aria-hidden="true" viewBox="0 0 24 24">
                      <path d="M9 3h6l1 2h4v2H4V5h4l1-2Zm1 6h2v9h-2V9Zm4 0h2v9h-2V9ZM7 9h2l1 11h4l1-11h2l-1.2 13H8.2L7 9Z" />
                    </svg>
                  </button>
                </div>
                {hasNote && isNoteExpanded ? <div className="entry-note">{entry.description}</div> : null}
              </div>
            );
          })}
        </div>
      ) : (
        <div className="entries-empty">Keine Zeiteintraege fuer diesen Tag.</div>
      )}
    </section>
  );
}

function formatDuration(totalSeconds: number): string {
  const seconds = Math.max(0, Math.floor(totalSeconds));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours === 0) return `${minutes}m`;
  return `${hours}h ${String(minutes).padStart(2, "0")}m`;
}
