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

import { ArchiveIcon, EditIcon, EyeIcon, TrashIcon } from "./ProjectIcons";
import type { ProjectsPageText } from "./projectTypes";

type ProjectActionsProps = {
  canToggleHiddenTasks: boolean;
  isArchivedProject: boolean;
  showHiddenTasks: boolean;
  t: ProjectsPageText;
  onAddTask: () => void;
  onArchiveProject: () => void;
  onDeleteProject: () => void;
  onEditProject: () => void;
  onReactivateProject: () => void;
  onToggleHiddenTasks: () => void;
};

export function ProjectActions({
  canToggleHiddenTasks,
  isArchivedProject,
  showHiddenTasks,
  t,
  onAddTask,
  onArchiveProject,
  onDeleteProject,
  onEditProject,
  onReactivateProject,
  onToggleHiddenTasks
}: ProjectActionsProps) {
  if (isArchivedProject) {
    return (
      <div className="project-detail-actions">
        <button className="primary-button" type="button" onClick={onReactivateProject}>
          {t.reactivateProject}
        </button>
        <button
          className="icon-button danger-icon-button"
          type="button"
          onClick={onDeleteProject}
          aria-label={t.deleteProject}
          title={t.deleteProject}
        >
          <TrashIcon />
        </button>
      </div>
    );
  }

  return (
    <div className="project-detail-actions">
      <button
        className={`icon-button eye-button ${showHiddenTasks && canToggleHiddenTasks ? "active" : ""}`}
        type="button"
        disabled={!canToggleHiddenTasks}
        onClick={onToggleHiddenTasks}
        aria-label={t.showHiddenTasks}
        title={t.showHiddenTasks}
      >
        <EyeIcon />
      </button>
      <button className="icon-button" type="button" onClick={onEditProject} aria-label={t.editProject} title={t.editProject}>
        <EditIcon />
      </button>
      <button className="primary-button" type="button" onClick={onAddTask}>
        {t.addTask}
      </button>
      <button className="icon-button" type="button" onClick={onArchiveProject} aria-label={t.archiveProject} title={t.archiveProject}>
        <ArchiveIcon />
      </button>
      <button
        className="icon-button danger-icon-button"
        type="button"
        onClick={onDeleteProject}
        aria-label={t.deleteProject}
        title={t.deleteProject}
      >
        <TrashIcon />
      </button>
    </div>
  );
}
