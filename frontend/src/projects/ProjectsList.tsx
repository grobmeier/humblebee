import { SideNavigationList } from "../components/SideNavigationList";
import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import type { ProjectsPageText, WorkItem } from "./projectTypes";

type ProjectsListProps = {
  language: DateLanguage;
  projects: WorkItem[];
  selectedProjectId?: number;
  t: ProjectsPageText;
  onCreateProject: () => void;
  onSelectProject: (projectId: number) => void;
};

export function ProjectsList({ language, projects, selectedProjectId, t, onCreateProject, onSelectProject }: ProjectsListProps) {
  return (
    <SideNavigationList
      action={
        <button className="primary-button" type="button" onClick={onCreateProject}>
          {t.addProject}
        </button>
      }
      ariaLabel={t.projectList}
      emptyText={t.emptyProjects}
      items={projects.map((project) => ({ id: project.id, label: labelWorkItemName(project.name, language) }))}
      selectedId={selectedProjectId}
      title={t.projectList}
      onSelect={(id) => onSelectProject(Number(id))}
    />
  );
}
