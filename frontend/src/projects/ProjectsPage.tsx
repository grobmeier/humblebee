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

import { FormEvent, useEffect, useMemo, useState } from "react";
import { ProjectDeleteModal } from "./ProjectDeleteModal";
import { ProjectDetail } from "./ProjectDetail";
import { ProjectNameModal } from "./ProjectNameModal";
import { ProjectsList } from "./ProjectsList";
import type { DateLanguage } from "../dashboard/dateFormat";
import { isActiveTask, isArchivedWorkItem, type ProjectModalState, type ProjectsPageText, type WorkItem } from "./projectTypes";

type ProjectsPageProps = {
  language: DateLanguage;
  selectedProjectId: number;
  t: ProjectsPageText;
  workItems: WorkItem[];
  onCreateProject: (name: string, sourceProjectId: number) => Promise<void>;
  onCreateTask: (projectId: number, name: string) => Promise<void>;
  onDeleteProject: (projectId: number) => Promise<void>;
  onSelectProject: (projectId: number) => void;
  onSetProjectActive: (projectId: number, active: boolean) => Promise<void>;
  onSetTaskActive: (taskId: number, active: boolean) => Promise<void>;
  onUpdateProject: (projectId: number, name: string) => Promise<void>;
};

export function ProjectsPage({
  language,
  selectedProjectId,
  t,
  workItems,
  onCreateProject,
  onCreateTask,
  onDeleteProject,
  onSelectProject,
  onSetProjectActive,
  onSetTaskActive,
  onUpdateProject
}: ProjectsPageProps) {
  const allProjects = useMemo(
    () => workItems.filter((item) => item.parentId == null && item.name.toLowerCase() !== "default"),
    [workItems]
  );
  const activeProjects = useMemo(() => allProjects.filter(isActiveTask), [allProjects]);
  const [showArchivedProjects, setShowArchivedProjects] = useState(false);
  const projects = showArchivedProjects ? allProjects : activeProjects;
  const selectedProject = projects.find((project) => project.id === selectedProjectId) ?? projects[0] ?? null;
  const tasks = selectedProject ? workItems.filter((item) => item.parentId === selectedProject.id) : [];
  const [modal, setModal] = useState<ProjectModalState>(null);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [copySourceProjectId, setCopySourceProjectId] = useState(0);
  const [showHiddenTasks, setShowHiddenTasks] = useState(false);
  const hasArchivedProjects = allProjects.some(isArchivedWorkItem);
  const hasHiddenTasks = tasks.some((task) => !isActiveTask(task));
  const visibleTasks = selectedProject && isArchivedWorkItem(selectedProject) ? tasks : showHiddenTasks ? tasks : tasks.filter(isActiveTask);

  useEffect(() => {
    if (!selectedProject && projects.length) {
      onSelectProject(projects[0].id);
    }
  }, [onSelectProject, projects, selectedProject]);

  useEffect(() => {
    setShowHiddenTasks(false);
  }, [selectedProject?.id]);

  useEffect(() => {
    if (!hasHiddenTasks && showHiddenTasks) {
      setShowHiddenTasks(false);
    }
  }, [hasHiddenTasks, showHiddenTasks]);

  useEffect(() => {
    if (!hasArchivedProjects && showArchivedProjects) {
      setShowArchivedProjects(false);
    }
  }, [hasArchivedProjects, showArchivedProjects]);

  function openCreateProjectModal() {
    setModal({ type: "create-project" });
    setName("");
    setCopySourceProjectId(0);
    setError(null);
  }

  function openEditProjectModal(project: WorkItem) {
    setModal({ type: "edit-project", project });
    setName(project.name);
    setCopySourceProjectId(0);
    setError(null);
  }

  function openCreateTaskModal(project: WorkItem) {
    setModal({ type: "create-task", project });
    setName("");
    setCopySourceProjectId(0);
    setError(null);
  }

  function openDeleteProjectModal(project: WorkItem) {
    setModal({ type: "delete-project", project });
    setName("");
    setCopySourceProjectId(0);
    setError(null);
  }

  function closeModal() {
    setModal(null);
    setName("");
    setCopySourceProjectId(0);
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
        await onCreateProject(trimmedName, copySourceProjectId);
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

  async function setProjectActive(project: WorkItem, active: boolean) {
    setError(null);
    try {
      await onSetProjectActive(project.id, active);
      if (!active) {
        setShowArchivedProjects(false);
      }
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <section className="projects-page" id="projects">
      <ProjectsList
        canToggleArchivedProjects={hasArchivedProjects}
        language={language}
        projects={projects}
        selectedProjectId={selectedProject?.id}
        showArchivedProjects={showArchivedProjects}
        t={t}
        onCreateProject={openCreateProjectModal}
        onSelectProject={onSelectProject}
        onToggleArchivedProjects={() => setShowArchivedProjects((value) => !value)}
      />
      <ProjectDetail
        error={error}
        language={language}
        selectedProject={selectedProject}
        canToggleHiddenTasks={hasHiddenTasks}
        showHiddenTasks={showHiddenTasks}
        t={t}
        tasks={visibleTasks}
        onAddTask={openCreateTaskModal}
        onArchiveProject={(project) => void setProjectActive(project, false)}
        onDeleteProject={openDeleteProjectModal}
        onEditProject={openEditProjectModal}
        onReactivateProject={(project) => void setProjectActive(project, true)}
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
        copySourceProjectId={copySourceProjectId}
        error={error}
        isSaving={isSaving}
        language={language}
        modal={activeModal}
        name={name}
        projects={activeProjects}
        t={t}
        onChange={setName}
        onClose={closeModal}
        onCopySourceChange={setCopySourceProjectId}
        onSubmit={submitModal}
      />
    );
  }
}
