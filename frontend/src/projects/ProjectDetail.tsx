import { ProjectActions } from "./ProjectActions";
import { ProjectTaskList } from "./ProjectTaskList";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectDetailProps = {
  error: string | null;
  selectedProject: WorkItem | null;
  showHiddenTasks: boolean;
  t: ProjectsPageText;
  tasks: WorkItem[];
  onAddTask: (project: WorkItem) => void;
  onDeleteProject: (project: WorkItem) => void;
  onEditProject: (project: WorkItem) => void;
  onToggleHiddenTasks: () => void;
  onToggleTaskCompleted: (task: WorkItem, completed: boolean) => void;
};

export function ProjectDetail({
  error,
  selectedProject,
  showHiddenTasks,
  t,
  tasks,
  onAddTask,
  onDeleteProject,
  onEditProject,
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
          <h1>{selectedProject.name}</h1>
        </div>
        <ProjectActions
          showHiddenTasks={showHiddenTasks}
          t={t}
          onAddTask={() => onAddTask(selectedProject)}
          onDeleteProject={() => onDeleteProject(selectedProject)}
          onEditProject={() => onEditProject(selectedProject)}
          onToggleHiddenTasks={onToggleHiddenTasks}
        />
      </div>

      {error ? <div className="errors alert alert-error">{error}</div> : null}

      <ProjectTaskList tasks={tasks} t={t} onToggleCompleted={onToggleTaskCompleted} />
    </section>
  );
}
