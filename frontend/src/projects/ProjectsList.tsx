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
