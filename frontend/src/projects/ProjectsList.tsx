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

import { SideNavigationList } from "../components/SideNavigationList";
import type { DateLanguage } from "../dashboard/dateFormat";
import { labelWorkItemName } from "../dashboard/workItemUtils";
import { EyeIcon } from "./ProjectIcons";
import { isArchivedWorkItem, type ProjectsPageText, type WorkItem } from "./projectTypes";

type ProjectsListProps = {
  canToggleArchivedProjects: boolean;
  language: DateLanguage;
  projects: WorkItem[];
  selectedProjectId?: number;
  showArchivedProjects: boolean;
  t: ProjectsPageText;
  onCreateProject: () => void;
  onSelectProject: (projectId: number) => void;
  onToggleArchivedProjects: () => void;
};

export function ProjectsList({
  canToggleArchivedProjects,
  language,
  projects,
  selectedProjectId,
  showArchivedProjects,
  t,
  onCreateProject,
  onSelectProject,
  onToggleArchivedProjects
}: ProjectsListProps) {
  return (
    <SideNavigationList
      action={
        <div className="side-navigation-actions">
          <button
            className={`icon-button eye-button ${showArchivedProjects && canToggleArchivedProjects ? "active" : ""}`}
            type="button"
            disabled={!canToggleArchivedProjects}
            onClick={onToggleArchivedProjects}
            aria-label={t.showArchivedProjects}
            title={t.showArchivedProjects}
          >
            <EyeIcon />
          </button>
          <button className="primary-button" type="button" onClick={onCreateProject}>
            {t.addProject}
          </button>
        </div>
      }
      ariaLabel={t.projectList}
      emptyText={t.emptyProjects}
      items={projects.map((project) => ({
        className: isArchivedWorkItem(project) ? "is-archived-project" : undefined,
        id: project.id,
        label: labelWorkItemName(project.name, language)
      }))}
      selectedId={selectedProjectId}
      title={t.projectList}
      onSelect={(id) => onSelectProject(Number(id))}
    />
  );
}
