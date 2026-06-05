import { ProjectTaskRow } from "./ProjectTaskRow";
import type { DateLanguage } from "../dashboard/dateFormat";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectTaskListProps = {
  language: DateLanguage;
  tasks: WorkItem[];
  t: ProjectsPageText;
  onToggleCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectTaskList({ language, tasks, t, onToggleCompleted }: ProjectTaskListProps) {
  if (!tasks.length) {
    return <p className="projects-empty">{t.emptyTasks}</p>;
  }

  return (
    <div className="project-task-list">
      {tasks.map((task) => (
        <ProjectTaskRow key={task.id} language={language} task={task} t={t} onToggleCompleted={onToggleCompleted} />
      ))}
    </div>
  );
}
