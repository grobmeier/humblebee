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

import type { FormEvent } from "react";
import type { guiapp } from "../../wailsjs/go/models";
import { FormRow, Modal } from "../components/Modal";
import type { ImportPageText } from "../dashboard/translations";

type TimeAndBillImportModalProps = {
  error: string | null;
  filePath: string;
  isImporting: boolean;
  isPreviewing: boolean;
  preview: guiapp.ImportPreview | null;
  result: guiapp.ImportResult | null;
  t: ImportPageText;
  onChooseFile: () => void;
  onClose: () => void;
  onImport: () => void;
  onPreview: () => void;
};

export function TimeAndBillImportModal({
  error,
  filePath,
  isImporting,
  isPreviewing,
  preview,
  result,
  t,
  onChooseFile,
  onClose,
  onImport,
  onPreview
}: TimeAndBillImportModalProps) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (preview && !result) {
      onImport();
    } else if (!preview) {
      onPreview();
    }
  }

  const activeSummary = result?.summary ?? preview?.summary ?? null;
  const conflicts = result?.conflicts ?? preview?.conflicts ?? [];
  const footer = result ? undefined : (
    <>
      {!preview ? (
        <button className="secondary-button" type="submit" disabled={!filePath || isPreviewing || isImporting}>
          {isPreviewing ? t.previewing : t.preview}
        </button>
      ) : (
        <button className="secondary-button" type="submit" disabled={isImporting || isPreviewing}>
          {isImporting ? t.importing : t.importAction}
        </button>
      )}
    </>
  );

  return (
    <Modal
      title={t.title}
      onClose={onClose}
      onSubmit={submit}
      footer={footer}
    >
      <FormRow label={t.file}>
        {!result ? (
          <button className="secondary-button" type="button" onClick={onChooseFile} disabled={isPreviewing || isImporting}>
            {t.chooseFile}
          </button>
        ) : null}
        <code className="modal-path-value">{filePath || t.noFileSelected}</code>
      </FormRow>

      {preview && preview.existingTimeEntryCount > 0 ? (
        <div className="alert alert-warning">
          {t.existingTimeWarning.replace("{count}", String(preview.existingTimeEntryCount))}
        </div>
      ) : null}

      {preview ? (
        <div className="modal-meta-grid">
          <span>{t.exportUuid}</span>
          <code>{preview.exportUuid}</code>
          <span>{t.sourceUser}</span>
          <strong>{preview.sourceUserEmail || "-"}</strong>
          <span>{t.exportedAt}</span>
          <strong>{preview.exportedAt || "-"}</strong>
        </div>
      ) : null}

      {result ? (
        <div className="import-complete">
          <span aria-hidden="true">✓</span>
          <strong>{t.completed}</strong>
        </div>
      ) : null}
      {activeSummary ? <ImportSummaryView mode={result ? "import" : "preview"} summary={activeSummary} t={t} /> : null}
      {conflicts.length ? <ImportConflictList conflicts={conflicts} t={t} /> : null}
      {error ? <div className="errors alert alert-error">{error}</div> : null}
    </Modal>
  );
}

function ImportSummaryView({ mode, summary, t }: { mode: "preview" | "import"; summary: guiapp.ImportSummary; t: ImportPageText }) {
  return (
    <div className="summary-grid">
      <SummaryRow label={t.projects} mode={mode} values={[summary.projectsCreated, summary.projectsMapped, summary.projectsSkipped]} t={t} />
      <SummaryRow label={t.tasks} mode={mode} values={[summary.tasksCreated, summary.tasksMapped, summary.tasksSkipped]} t={t} />
      <div className="summary-row">
        <span>{t.timeEntries}</span>
        <small>{mode === "preview" ? t.wouldCreate : t.created}: {summary.timeEntriesCreated}</small>
        <small>{mode === "preview" ? t.wouldSkip : t.skipped}: {summary.timeEntriesSkipped}</small>
        <small>{mode === "preview" ? t.wouldConflict : t.conflicts}: {summary.timeEntryConflicts}</small>
      </div>
      {summary.alreadyImported ? <div className="alert alert-warning">{t.alreadyImported}</div> : null}
    </div>
  );
}

function SummaryRow({ label, mode, values, t }: { label: string; mode: "preview" | "import"; values: number[]; t: ImportPageText }) {
  return (
    <div className="summary-row">
      <span>{label}</span>
      <small>{mode === "preview" ? t.wouldCreate : t.created}: {values[0]}</small>
      <small>{mode === "preview" ? t.wouldMap : t.mapped}: {values[1]}</small>
      <small>{mode === "preview" ? t.wouldSkip : t.skipped}: {values[2]}</small>
    </div>
  );
}

function ImportConflictList({ conflicts, t }: { conflicts: guiapp.ImportConflict[]; t: ImportPageText }) {
  return (
    <details className="technical-details import-conflicts">
      <summary>{t.conflictDetails.replace("{count}", String(conflicts.length))}</summary>
      <ul>
        {conflicts.slice(0, 25).map((conflict) => (
          <li key={conflict.timeEntryUuid}>
            <strong>{conflict.projectName}</strong> / {conflict.taskName}: {formatImportDate(conflict.start)} - {formatImportDate(conflict.end)}
          </li>
        ))}
      </ul>
    </details>
  );
}

function formatImportDate(value: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}
