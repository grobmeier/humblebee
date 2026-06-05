export type WorkItemNode = {
  depth: number;
  id: number;
  name: string;
  parentId?: number | null;
};

export type WorkItemLabelLanguage = "de" | "en";

export type WorkItemDisplay = {
  projectName: string;
  taskName: string;
};

const reservedWorkItemLabels: Record<string, Record<WorkItemLabelLanguage, string>> = {
  "@": {
    de: "Abwesenheiten",
    en: "Absences"
  },
  "@break": {
    de: "Pause",
    en: "Break"
  },
  "@overtime": {
    de: "Überstundenausgleich",
    en: "Overtime compensation"
  },
  "@public_holiday": {
    de: "Feiertag",
    en: "Public holiday"
  },
  "@sick_leave": {
    de: "Krankheit",
    en: "Sick leave"
  },
  "@vacation": {
    de: "Urlaub",
    en: "Vacation"
  }
};

export function displayWorkItem(workItemId: number, workItems: WorkItemNode[], language: WorkItemLabelLanguage = "de"): WorkItemDisplay {
  const path = workItemPath(workItemId, workItems);
  if (!path.length) {
    return {
      projectName: "Default",
      taskName: ""
    };
  }

  return {
    projectName: labelWorkItemName(path[0]?.name ?? "Default", language),
    taskName: path[1] ? labelWorkItemName(path[1].name, language) : ""
  };
}

export function labelWorkItemName(name: string, language: WorkItemLabelLanguage = "de"): string {
  return reservedWorkItemLabels[name]?.[language] ?? name;
}

export function isReservedAbsenceWorkItemName(name: string): boolean {
  return name === "@" || name.startsWith("@");
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
