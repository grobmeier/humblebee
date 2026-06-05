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
