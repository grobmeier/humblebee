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
