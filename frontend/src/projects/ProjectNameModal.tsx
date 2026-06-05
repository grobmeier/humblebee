import type { FormEvent } from "react";
import { FormRow, Modal } from "../components/Modal";
import type { ProjectFormModalState, ProjectsPageText } from "./projectTypes";

type ProjectNameModalProps = {
  error: string | null;
  isSaving: boolean;
  modal: ProjectFormModalState;
  name: string;
  t: ProjectsPageText;
  onChange: (name: string) => void;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function ProjectNameModal({ error, isSaving, modal, name, t, onChange, onClose, onSubmit }: ProjectNameModalProps) {
  return (
    <Modal
      title={modalTitle(modal, t)}
      onClose={onClose}
      onSubmit={onSubmit}
      footer={
        <button className="primary-button modal-submit-button" type="submit" disabled={isSaving}>
          {modalSubmitLabel(modal, t)}
        </button>
      }
    >
      {error ? <div className="errors alert alert-error">{error}</div> : null}
      <FormRow label={t.name}>
        <input className="tab-form-control" autoFocus value={name} onChange={(event) => onChange(event.target.value)} />
      </FormRow>
    </Modal>
  );
}

function modalTitle(modal: ProjectFormModalState, t: ProjectsPageText): string {
  if (modal.type === "create-project") {
    return t.createProject;
  }
  if (modal.type === "edit-project") {
    return t.editProject;
  }
  return t.createTask;
}

function modalSubmitLabel(modal: ProjectFormModalState, t: ProjectsPageText): string {
  if (modal.type === "create-project") {
    return t.createProject;
  }
  if (modal.type === "edit-project") {
    return t.saveProject;
  }
  return t.createTask;
}
