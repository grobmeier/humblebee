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
