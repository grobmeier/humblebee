import type { FormEvent } from "react";
import { Modal } from "../components/Modal";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectDeleteModalProps = {
  error: string | null;
  isSaving: boolean;
  project: WorkItem;
  t: ProjectsPageText;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function ProjectDeleteModal({ error, isSaving, project, t, onClose, onSubmit }: ProjectDeleteModalProps) {
  return (
    <Modal
      title={t.deleteProjectTitle}
      onClose={onClose}
      onSubmit={onSubmit}
      footer={
        <>
          <button className="secondary-button" type="button" onClick={onClose}>
            {t.cancel}
          </button>
          <button className="secondary-button danger-button" type="submit" disabled={isSaving}>
            {t.deleteProjectConfirm}
          </button>
        </>
      }
    >
      {error ? <div className="errors alert alert-error">{error}</div> : null}
      <p className="project-delete-warning">{t.deleteProjectWarning.replace("{name}", project.name)}</p>
    </Modal>
  );
}
