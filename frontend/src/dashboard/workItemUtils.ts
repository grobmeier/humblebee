export type WorkItemNode = {
  depth: number;
  id: number;
  name: string;
  parentId?: number | null;
};

export type WorkItemDisplay = {
  projectName: string;
  taskName: string;
};

export function displayWorkItem(workItemId: number, workItems: WorkItemNode[]): WorkItemDisplay {
  const path = workItemPath(workItemId, workItems);
  if (!path.length) {
    return {
      projectName: "Default",
      taskName: ""
    };
  }

  return {
    projectName: path[0]?.name ?? "Default",
    taskName: path[1]?.name ?? ""
  };
}

export function workItemPath(workItemId: number, workItems: WorkItemNode[]): WorkItemNode[] {
  const byId = new Map(workItems.map((workItem) => [workItem.id, workItem]));
  const path: WorkItemNode[] = [];
  let current = byId.get(workItemId);

  while (current) {
    path.unshift(current);
    current = current.parentId ? byId.get(current.parentId) : undefined;
  }

  return path;
}
