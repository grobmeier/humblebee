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

import { ProjectTaskRow } from "./ProjectTaskRow";
import type { DateLanguage } from "../dashboard/dateFormat";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectTaskListProps = {
  language: DateLanguage;
  tasks: WorkItem[];
  t: ProjectsPageText;
  onDelete: (task: WorkItem) => void;
  onEdit: (task: WorkItem) => void;
  onToggleCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectTaskList({ language, tasks, t, onDelete, onEdit, onToggleCompleted }: ProjectTaskListProps) {
  if (!tasks.length) {
    return <p className="projects-empty">{t.emptyTasks}</p>;
  }

  return (
    <div className="project-task-list">
      {tasks.map((task) => (
        <ProjectTaskRow
          key={task.id}
          language={language}
          task={task}
          t={t}
          onDelete={onDelete}
          onEdit={onEdit}
          onToggleCompleted={onToggleCompleted}
        />
      ))}
    </div>
  );
}
