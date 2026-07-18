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

import { ProjectActions } from "./ProjectActions";
import { ProjectTaskList } from "./ProjectTaskList";
import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import { isArchivedWorkItem, type ProjectsPageText, type WorkItem } from "./projectTypes";

type ProjectDetailProps = {
  canToggleHiddenTasks: boolean;
  error: string | null;
  language: DateLanguage;
  selectedProject: WorkItem | null;
  showHiddenTasks: boolean;
  t: ProjectsPageText;
  tasks: WorkItem[];
  onAddTask: (project: WorkItem) => void;
  onArchiveProject: (project: WorkItem) => void;
  onDeleteProject: (project: WorkItem) => void;
  onDeleteTask: (task: WorkItem) => void;
  onEditProject: (project: WorkItem) => void;
  onEditTask: (task: WorkItem) => void;
  onReactivateProject: (project: WorkItem) => void;
  onToggleHiddenTasks: () => void;
  onToggleTaskCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectDetail({
  canToggleHiddenTasks,
  error,
  language,
  selectedProject,
  showHiddenTasks,
  t,
  tasks,
  onAddTask,
  onArchiveProject,
  onDeleteProject,
  onDeleteTask,
  onEditProject,
  onEditTask,
  onReactivateProject,
  onToggleHiddenTasks,
  onToggleTaskCompleted
}: ProjectDetailProps) {
  if (!selectedProject) {
    return (
      <section className="project-detail-panel">
        <p className="projects-empty">{t.selectProject}</p>
      </section>
    );
  }

  return (
    <section className="project-detail-panel">
      <div className="project-detail-header">
        <div>
          <h1>{labelWorkItemName(selectedProject.name, language)}</h1>
        </div>
        <ProjectActions
          canToggleHiddenTasks={canToggleHiddenTasks}
          isArchivedProject={isArchivedWorkItem(selectedProject)}
          showHiddenTasks={showHiddenTasks}
          t={t}
          onAddTask={() => onAddTask(selectedProject)}
          onArchiveProject={() => onArchiveProject(selectedProject)}
          onDeleteProject={() => onDeleteProject(selectedProject)}
          onEditProject={() => onEditProject(selectedProject)}
          onReactivateProject={() => onReactivateProject(selectedProject)}
          onToggleHiddenTasks={onToggleHiddenTasks}
        />
      </div>

      {error ? <div className="errors alert alert-error">{error}</div> : null}

      <ProjectTaskList
        language={language}
        tasks={tasks}
        t={t}
        onDelete={onDeleteTask}
        onEdit={onEditTask}
        onToggleCompleted={onToggleTaskCompleted}
      />
    </section>
  );
}
