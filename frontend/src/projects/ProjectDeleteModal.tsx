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
