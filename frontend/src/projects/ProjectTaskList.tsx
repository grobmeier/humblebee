import { ProjectTaskRow } from "./ProjectTaskRow";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectTaskListProps = {
  tasks: WorkItem[];
  t: ProjectsPageText;
  onToggleCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectTaskList({ tasks, t, onToggleCompleted }: ProjectTaskListProps) {
  if (!tasks.length) {
    return <p className="projects-empty">{t.emptyTasks}</p>;
  }

  return (
    <div className="project-task-list">
      {tasks.map((task) => (
        <ProjectTaskRow key={task.id} task={task} t={t} onToggleCompleted={onToggleCompleted} />
      ))}
    </div>
  );
}
