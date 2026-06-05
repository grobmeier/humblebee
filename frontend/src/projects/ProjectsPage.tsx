import { FormEvent, useEffect, useMemo, useState } from "react";
import { ProjectDeleteModal } from "./ProjectDeleteModal";
import { ProjectDetail } from "./ProjectDetail";
import { ProjectNameModal } from "./ProjectNameModal";
import { ProjectsList } from "./ProjectsList";
import { isActiveTask, type ProjectModalState, type ProjectsPageText, type WorkItem } from "./projectTypes";

type ProjectsPageProps = {
  selectedProjectId: number;
  t: ProjectsPageText;
  workItems: WorkItem[];
  onCreateProject: (name: string) => Promise<void>;
  onCreateTask: (projectId: number, name: string) => Promise<void>;
  onDeleteProject: (projectId: number) => Promise<void>;
  onSelectProject: (projectId: number) => void;
  onSetTaskActive: (taskId: number, active: boolean) => Promise<void>;
  onUpdateProject: (projectId: number, name: string) => Promise<void>;
};

export function ProjectsPage({
  selectedProjectId,
  t,
  workItems,
  onCreateProject,
  onCreateTask,
  onDeleteProject,
  onSelectProject,
  onSetTaskActive,
  onUpdateProject
}: ProjectsPageProps) {
  const projects = useMemo(
    () => workItems.filter((item) => item.parentId == null && item.name.toLowerCase() !== "default" && isActiveTask(item)),
    [workItems]
  );
  const selectedProject = projects.find((project) => project.id === selectedProjectId) ?? projects[0] ?? null;
  const tasks = selectedProject ? workItems.filter((item) => item.parentId === selectedProject.id) : [];
  const [modal, setModal] = useState<ProjectModalState>(null);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [showHiddenTasks, setShowHiddenTasks] = useState(false);
  const visibleTasks = showHiddenTasks ? tasks : tasks.filter(isActiveTask);

  useEffect(() => {
    if (!selectedProject && projects.length) {
      onSelectProject(projects[0].id);
    }
  }, [onSelectProject, projects, selectedProject]);

  function openCreateProjectModal() {
    setModal({ type: "create-project" });
    setName("");
    setError(null);
  }

  function openEditProjectModal(project: WorkItem) {
    setModal({ type: "edit-project", project });
    setName(project.name);
    setError(null);
  }

  function openCreateTaskModal(project: WorkItem) {
    setModal({ type: "create-task", project });
    setName("");
    setError(null);
  }

  function openDeleteProjectModal(project: WorkItem) {
    setModal({ type: "delete-project", project });
    setName("");
    setError(null);
  }

  function closeModal() {
    setModal(null);
    setName("");
    setError(null);
  }

  async function submitModal(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!modal) {
      return;
    }

    if (modal.type === "delete-project") {
      await submitDeleteProject(modal.project);
      return;
    }

    await submitNameModal(modal);
  }

  async function submitDeleteProject(project: WorkItem) {
    setIsSaving(true);
    setError(null);
    try {
      await onDeleteProject(project.id);
      closeModal();
    } catch (err) {
      setError(String(err));
    } finally {
      setIsSaving(false);
    }
  }

  async function submitNameModal(modal: Exclude<ProjectModalState, { type: "delete-project"; project: WorkItem } | null>) {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setError(t.nameRequired);
      return;
    }

    setIsSaving(true);
    setError(null);
    try {
      if (modal.type === "create-project") {
        await onCreateProject(trimmedName);
      } else if (modal.type === "edit-project") {
        await onUpdateProject(modal.project.id, trimmedName);
      } else {
        await onCreateTask(modal.project.id, trimmedName);
      }
      closeModal();
    } catch (err) {
      setError(String(err));
    } finally {
      setIsSaving(false);
    }
  }

  async function toggleTaskCompleted(task: WorkItem, completed: boolean) {
    setError(null);
    try {
      await onSetTaskActive(task.id, !completed);
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <section className="projects-page" id="projects">
      <ProjectsList
        projects={projects}
        selectedProjectId={selectedProject?.id}
        t={t}
        onCreateProject={openCreateProjectModal}
        onSelectProject={onSelectProject}
      />
      <ProjectDetail
        error={error}
        selectedProject={selectedProject}
        showHiddenTasks={showHiddenTasks}
        t={t}
        tasks={visibleTasks}
        onAddTask={openCreateTaskModal}
        onDeleteProject={openDeleteProjectModal}
        onEditProject={openEditProjectModal}
        onToggleHiddenTasks={() => setShowHiddenTasks((value) => !value)}
        onToggleTaskCompleted={toggleTaskCompleted}
      />
      {modal ? renderModal(modal) : null}
    </section>
  );

  function renderModal(activeModal: NonNullable<ProjectModalState>) {
    if (activeModal.type === "delete-project") {
      return (
        <ProjectDeleteModal
          error={error}
          isSaving={isSaving}
          project={activeModal.project}
          t={t}
          onClose={closeModal}
          onSubmit={submitModal}
        />
      );
    }

    return (
      <ProjectNameModal
        error={error}
        isSaving={isSaving}
        modal={activeModal}
        name={name}
        t={t}
        onChange={setName}
        onClose={closeModal}
        onSubmit={submitModal}
      />
    );
  }
}
