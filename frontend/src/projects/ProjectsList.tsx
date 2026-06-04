import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectsListProps = {
  projects: WorkItem[];
  selectedProjectId?: number;
  t: ProjectsPageText;
  onCreateProject: () => void;
  onSelectProject: (projectId: number) => void;
};

export function ProjectsList({ projects, selectedProjectId, t, onCreateProject, onSelectProject }: ProjectsListProps) {
  return (
    <aside className="projects-list-panel" aria-label={t.projectList}>
      <div className="projects-list-header">
        <h2>{t.projectList}</h2>
        <button className="primary-button" type="button" onClick={onCreateProject}>
          {t.addProject}
        </button>
      </div>

      {projects.length ? (
        <div className="projects-list">
          {projects.map((project) => (
            <button
              className={`project-list-item ${project.id === selectedProjectId ? "selected" : ""}`}
              key={project.id}
              type="button"
              onClick={() => onSelectProject(project.id)}
            >
              {project.name}
            </button>
          ))}
        </div>
      ) : (
        <p className="projects-empty">{t.emptyProjects}</p>
      )}
    </aside>
  );
}
