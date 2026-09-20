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

import { useState, type FormEvent } from "react";
import { BackupDatabase } from "../../wailsjs/go/guiapp/App";
import type { guiapp } from "../../wailsjs/go/models";
import { Modal } from "../components/Modal";
import type { DatabasePageText } from "../dashboard/translations";

type DatabaseSwitchModalProps = {
  currentDatabasePath: string;
  language: "de" | "en";
  databaseInfo: guiapp.DatabaseInfo | null;
  error: string | null;
  isSaving: boolean;
  t: DatabasePageText;
  onCreateNew: () => void;
  onOpenExisting: () => void;
  onClose: () => void;
  onUseDefault: () => void;
};

export function DatabaseSwitchModal({
  currentDatabasePath,
  language,
  databaseInfo,
  error,
  isSaving,
  t,
  onCreateNew,
  onOpenExisting,
  onClose,
  onUseDefault
}: DatabaseSwitchModalProps) {
  const [isBackingUp, setIsBackingUp] = useState(false);
  const [backupPath, setBackupPath] = useState("");
  const [backupError, setBackupError] = useState("");
  const busy = isSaving || isBackingUp;
  const de = language === "de";

  async function backup() {
    setIsBackingUp(true);
    setBackupError("");
    setBackupPath("");
    try { setBackupPath(await BackupDatabase(currentDatabasePath)); }
    catch (error) { setBackupError(String(error)); }
    finally { setIsBackingUp(false); }
  }
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <Modal
      title={t.title}
      onClose={() => { if (!busy) onClose(); }}
      onSubmit={submit}
      footer={
        <>
          <button className="secondary-button" type="button" onClick={onUseDefault} disabled={busy}>
            {t.useDefault}
          </button>
        </>
      }
    >
      <div className="modal-meta-grid">
        <span>{t.current}</span>
        <code>{currentDatabasePath}</code>
        {databaseInfo ? (
          <>
            <span>{t.defaultPath}</span>
            <code>{databaseInfo.defaultPath}</code>
          </>
        ) : null}
      </div>
      <div className="database-choice-actions">
        <button className="secondary-button database-choice-button" type="button" onClick={onOpenExisting} disabled={busy}>
          <strong>{t.openExisting}</strong>
          <span>{t.openExistingHint}</span>
        </button>
        <button className="secondary-button database-choice-button" type="button" onClick={onCreateNew} disabled={busy}>
          <strong>{t.createNew}</strong>
          <span>{t.createNewHint}</span>
        </button>
      </div>
      <button className="secondary-button" type="button" disabled={busy} onClick={() => void backup()}>
        {isBackingUp ? (de ? "Sicherung läuft…" : "Backing up…") : (de ? "Datenbank sichern" : "Back up database")}
      </button>
      {backupPath ? <p role="status">{de ? "Datenbank gesichert:" : "Database backed up:"} <code>{backupPath}</code></p> : null}
      {backupError ? <div role="alert" className="errors alert alert-error">{backupError}</div> : null}
      {t.switchWarning ? <div className="alert alert-warning">{t.switchWarning}</div> : null}
      {error ? <div className="errors alert alert-error">{error}</div> : null}
    </Modal>
  );
}
