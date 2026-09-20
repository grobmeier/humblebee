import { useEffect, useState } from "react";
import { GetRecentNotes } from "../../wailsjs/go/guiapp/App";
import type { DateLanguage } from "./dateFormat";

type Props = {
  databasePath: string;
  taskId: number;
  value: string;
  language: DateLanguage;
  onSelect: (value: string) => void;
};

export function RecentNotes({ databasePath, taskId, value, language, onSelect }: Props) {
  const [notes, setNotes] = useState<string[]>([]);
  const [pending, setPending] = useState<string | null>(null);
  const de = language === "de";
  useEffect(() => {
    let cancelled = false;
    setNotes([]);
    setPending(null);
    if (taskId) void GetRecentNotes(databasePath, taskId)
      .then((rows) => { if (!cancelled) setNotes(rows ?? []); })
      .catch(() => { /* Suggestions must never interrupt booking. */ });
    return () => { cancelled = true; };
  }, [databasePath, taskId]);
  if (!notes.length) return null;
  return (
    <details className="recent-notes" key={`${databasePath}:${taskId}`}>
      <summary>{de ? "Letzte Notizen" : "Recent notes"}</summary>
      <ul>{notes.map((note) => (
        <li key={note}>
          <button type="button" className="secondary-button" title={note} aria-label={note}
            onClick={() => {
              if (value && value !== note) setPending(note);
              else { onSelect(note); setPending(null); }
            }}>
            {note.length > 120 ? `${note.slice(0, 120)}…` : note}
          </button>
        </li>
      ))}</ul>
      {pending !== null ? <div className="recent-notes-confirm" role="group" aria-label={de ? "Notiz ersetzen?" : "Replace note?"}>
        <p>{de ? "Vorhandene Notiz ersetzen?" : "Replace the current note?"}</p>
        <button className="secondary-button" type="button" onClick={() => { onSelect(pending); setPending(null); }}>{de ? "Ersetzen" : "Replace"}</button>
        <button className="secondary-button" type="button" onClick={() => setPending(null)}>{de ? "Behalten" : "Keep current"}</button>
      </div> : null}
    </details>
  );
}
