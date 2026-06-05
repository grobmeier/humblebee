import type { FormEvent } from "react";
import type { guiapp } from "../../wailsjs/go/models";
import { Modal } from "../components/Modal";
import type { DatabasePageText } from "../dashboard/translations";

type DatabaseSwitchModalProps = {
  currentDatabasePath: string;
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
  databaseInfo,
  error,
  isSaving,
  t,
  onCreateNew,
  onOpenExisting,
  onClose,
  onUseDefault
}: DatabaseSwitchModalProps) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <Modal
      title={t.title}
      onClose={onClose}
      onSubmit={submit}
      footer={
        <>
          <button className="secondary-button" type="button" onClick={onUseDefault} disabled={isSaving}>
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
        <button className="secondary-button database-choice-button" type="button" onClick={onOpenExisting} disabled={isSaving}>
          <strong>{t.openExisting}</strong>
          <span>{t.openExistingHint}</span>
        </button>
        <button className="secondary-button database-choice-button" type="button" onClick={onCreateNew} disabled={isSaving}>
          <strong>{t.createNew}</strong>
          <span>{t.createNewHint}</span>
        </button>
      </div>
      {t.switchWarning ? <div className="alert alert-warning">{t.switchWarning}</div> : null}
      {error ? <div className="errors alert alert-error">{error}</div> : null}
    </Modal>
  );
}
