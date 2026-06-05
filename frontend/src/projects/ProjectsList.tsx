import { SideNavigationList } from "../components/SideNavigationList";
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
    <SideNavigationList
      action={
        <button className="primary-button" type="button" onClick={onCreateProject}>
          {t.addProject}
        </button>
      }
      ariaLabel={t.projectList}
      emptyText={t.emptyProjects}
      items={projects.map((project) => ({ id: project.id, label: project.name }))}
      selectedId={selectedProjectId}
      title={t.projectList}
      onSelect={(id) => onSelectProject(Number(id))}
    />
  );
}
