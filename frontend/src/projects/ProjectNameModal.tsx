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
import { FormRow, Modal } from "../components/Modal";
import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import type { ProjectFormModalState, ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectNameModalProps = {
  copySourceProjectId: number;
  error: string | null;
  isSaving: boolean;
  language: DateLanguage;
  modal: ProjectFormModalState;
  name: string;
  projects: WorkItem[];
  t: ProjectsPageText;
  onChange: (name: string) => void;
  onClose: () => void;
  onCopySourceChange: (projectId: number) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function ProjectNameModal({
  copySourceProjectId,
  error,
  isSaving,
  language,
  modal,
  name,
  projects,
  t,
  onChange,
  onClose,
  onCopySourceChange,
  onSubmit
}: ProjectNameModalProps) {
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
      {modal.type === "create-project" ? (
        <FormRow label={t.copyTasksFrom}>
          <select className="tab-form-control" value={copySourceProjectId} onChange={(event) => onCopySourceChange(Number(event.target.value))}>
            <option value={0}>{t.noTaskTemplate}</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {labelWorkItemName(project.name, language)}
              </option>
            ))}
          </select>
        </FormRow>
      ) : null}
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
