import { EditIcon, EyeIcon, TrashIcon } from "./ProjectIcons";
import type { ProjectsPageText } from "./projectTypes";

type ProjectActionsProps = {
  canToggleHiddenTasks: boolean;
  showHiddenTasks: boolean;
  t: ProjectsPageText;
  onAddTask: () => void;
  onDeleteProject: () => void;
  onEditProject: () => void;
  onToggleHiddenTasks: () => void;
};

export function ProjectActions({
  canToggleHiddenTasks,
  showHiddenTasks,
  t,
  onAddTask,
  onDeleteProject,
  onEditProject,
  onToggleHiddenTasks
}: ProjectActionsProps) {
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
