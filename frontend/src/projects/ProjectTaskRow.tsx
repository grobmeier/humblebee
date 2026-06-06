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

import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import { isActiveTask, type ProjectsPageText, type WorkItem } from "./projectTypes";

type ProjectTaskRowProps = {
  language: DateLanguage;
  task: WorkItem;
  t: ProjectsPageText;
  onToggleCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectTaskRow({ language, task, t, onToggleCompleted }: ProjectTaskRowProps) {
  const completed = !isActiveTask(task);

  return (
    <div className={`project-task-row ${completed ? "is-hidden-task" : ""}`}>
      <label>
        <input
          type="checkbox"
          checked={completed}
          onChange={(event) => onToggleCompleted(task, event.target.checked)}
          aria-label={t.completedTask}
        />
        <span>
          <strong>{labelWorkItemName(task.name, language)}</strong>
        </span>
      </label>
    </div>
  );
}
