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

export type WorkItem = {
  id: number;
  name: string;
  parentId?: number | null;
  depth: number;
  status?: string;
};

export type ProjectsPageText = {
  addProject: string;
  addTask: string;
  cancel: string;
  completedTask: string;
  copyTasksFrom: string;
  createProject: string;
  createTask: string;
  deleteProject: string;
  deleteProjectConfirm: string;
  deleteProjectTitle: string;
  deleteProjectWarning: string;
  editProject: string;
  emptyProjects: string;
  emptyTasks: string;
  name: string;
  nameRequired: string;
  noTaskTemplate: string;
  projectList: string;
  saveProject: string;
  selectProject: string;
  showHiddenTasks: string;
  tasks: string;
};

export type ProjectFormModalState =
  | { type: "create-project" }
  | { type: "edit-project"; project: WorkItem }
  | { type: "create-task"; project: WorkItem };

export type ProjectModalState = ProjectFormModalState | { type: "delete-project"; project: WorkItem } | null;

export function isActiveTask(item: WorkItem): boolean {
  return (item.status ?? "ACTIVE") === "ACTIVE";
}
