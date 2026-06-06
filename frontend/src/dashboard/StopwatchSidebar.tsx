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

import { formatDisplayDate, type DateLanguage } from "./dateFormat";
import { displayWorkItem, isReservedAbsenceWorkItemName, labelWorkItemName, workItemPath } from "./workItemUtils";

type WorkItem = { id: number; name: string; parentId?: number | null; depth: number; status?: string };

type Stopwatch = {
  durationSeconds: number;
  endDate: string;
  endTime: string;
  id: number;
  conflicting: boolean;
  running: boolean;
  startDate: string;
  startTime: string;
  workItemId?: number;
  workItemName: string;
};

type StopwatchSidebarProps = {
  selectedWorkItemId: number;
  stopwatches: Stopwatch[];
  language: DateLanguage;
  nowTimestamp: number;
  workItems: WorkItem[];
  onBookStopwatch: (stopwatch: Stopwatch) => void;
  onSelectWorkItem: (workItemId: number) => void;
  onDiscardStopwatch: (stopwatchId: number) => void;
  onStart: (workItemId?: number) => void;
  onStop: () => void;
  t: {
    createStopwatch: string;
    book: string;
    discardRunning: string;
    selectWorkItem: string;
    start: string;
    stopStopwatch: string;
  };
};

export function StopwatchSidebar({
  selectedWorkItemId,
  stopwatches,
  language,
  nowTimestamp,
  workItems,
  onBookStopwatch,
  onSelectWorkItem,
  onDiscardStopwatch,
  onStart,
  onStop,
  t
}: StopwatchSidebarProps) {
  const openWorkItemIds = new Set(stopwatches.map((stopwatch) => stopwatch.workItemId ?? 0));
  const availableWorkItems = workItems.filter(
    (workItem) =>
      isActiveWorkItem(workItem) &&
      workItem.name.toLowerCase() !== "default" &&
      !isReservedAbsenceWorkItemName(workItem.name) &&
      !openWorkItemIds.has(workItem.id)
  );
  const selectedWorkItemAvailable = selectedWorkItemId === 0 || availableWorkItems.some((workItem) => workItem.id === selectedWorkItemId);
  const selectedValue = selectedWorkItemAvailable ? selectedWorkItemId : 0;
  const groupedWorkItems = groupWorkItemsForStopwatch(availableWorkItems, workItems, language);

  return (
    <aside className="stopwatch-panel">
      <div className="stopwatch-create">
        <h2>{t.createStopwatch}</h2>
        <select value={selectedValue} onChange={(event) => onSelectWorkItem(Number(event.target.value))} aria-label={t.selectWorkItem}>
          <option value={0}></option>
          {groupedWorkItems.ungrouped.map((workItem) => (
            <option key={workItem.id} value={workItem.id}>
              {formatWorkItemOption(workItem, workItems, language)}
            </option>
          ))}
          {groupedWorkItems.groups.map((group) => (
            <optgroup key={group.project.id} label={labelWorkItemName(group.project.name, language)}>
              {group.items.map((workItem) => (
                <option key={workItem.id} value={workItem.id}>
                  {formatGroupedWorkItemOption(workItem, workItems, language)}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
        <div className="timer-actions">
          <button className="primary-button" disabled={selectedValue === 0} onClick={() => onStart(selectedValue)} type="button">
            {t.start}
          </button>
        </div>
      </div>

      {stopwatches.map((stopwatch) => {
        const display = displayWorkItem(stopwatch.workItemId ?? 0, workItems, language);
        const projectName = display.projectName === "Default" && stopwatch.workItemName ? stopwatch.workItemName : display.projectName;
        return (
          <div
            className={`timer-card ${stopwatch.running ? "active" : "book"} ${stopwatch.conflicting ? "conflict" : ""}`}
            key={stopwatch.id}
            style={{ borderLeftColor: stopwatch.conflicting ? "#d77" : "#5bb75b" }}
          >
            <div className="timer-card-dismiss-row">
              <button className="discard-button" type="button" onClick={() => onDiscardStopwatch(stopwatch.id)} aria-label={t.discardRunning} title={t.discardRunning}>
                ×
              </button>
            </div>
            <div className="timer-card-main">
              <div>
                <strong>{projectName}</strong>
                {display.taskName ? <span>{display.taskName}</span> : null}
                {!stopwatch.running ? <span>{formatDisplayDate(stopwatch.startDate, language)}</span> : null}
              </div>
            </div>
            <div className="timer-card-times">
              <span>{stopwatch.startTime}</span>
              {!stopwatch.running ? <span>{stopwatch.endTime}</span> : null}
              <span>{formatStopwatchDuration(stopwatch, nowTimestamp)}</span>
            </div>
            <div className="timer-card-actions">
              {stopwatch.conflicting ? (
                <button className="book-button" type="button" onClick={() => onBookStopwatch(stopwatch)}>
                  {t.book}
                </button>
              ) : null}
              {stopwatch.running ? (
                <button className="stop-button" type="button" onClick={onStop} aria-label={t.stopStopwatch} title={t.stopStopwatch}>
                  {t.stopStopwatch}
                </button>
              ) : (
                <button className="play-button" type="button" onClick={() => onStart(stopwatch.workItemId ?? 0)} aria-label={t.start}>
                  {t.start}
                </button>
              )}
            </div>
          </div>
        );
      })}
    </aside>
  );
}

function isActiveWorkItem(workItem: WorkItem): boolean {
  return (workItem.status ?? "ACTIVE") === "ACTIVE";
}

type StopwatchWorkItemGroup = {
  project: WorkItem;
  items: WorkItem[];
};

function groupWorkItemsForStopwatch(availableWorkItems: WorkItem[], allWorkItems: WorkItem[], language: DateLanguage): { groups: StopwatchWorkItemGroup[]; ungrouped: WorkItem[] } {
  const availableByID = new Map(availableWorkItems.map((workItem) => [workItem.id, workItem]));
  const projectChildren = new Map<number, WorkItem[]>();
  const ungrouped: WorkItem[] = [];

  for (const workItem of availableWorkItems) {
    const path = workItemPath(workItem.id, allWorkItems);
    const project = path[0];
    if (!project || project.id === workItem.id) {
      continue;
    }
    const existing = projectChildren.get(project.id) ?? [];
    existing.push(workItem);
    projectChildren.set(project.id, existing);
  }

  for (const workItem of availableWorkItems) {
    const path = workItemPath(workItem.id, allWorkItems);
    const project = path[0];
    if (!project) {
      ungrouped.push(workItem);
      continue;
    }
    if (project.id !== workItem.id) {
      continue;
    }
    if (!projectChildren.has(workItem.id)) {
      ungrouped.push(workItem);
    }
  }

  const groups = Array.from(projectChildren.entries())
    .map(([projectID, items]) => {
      const project = availableByID.get(projectID) ?? allWorkItems.find((workItem) => workItem.id === projectID);
      return project ? { project, items: sortWorkItemsByLabel(items, language) } : null;
    })
    .filter((group): group is StopwatchWorkItemGroup => group !== null)
    .sort((a, b) => compareWorkItemLabels(a.project, b.project, language));

  return {
    groups,
    ungrouped: sortWorkItemsByLabel(ungrouped, language)
  };
}

function formatWorkItemOption(workItem: WorkItem, workItems: WorkItem[], language: DateLanguage): string {
  const display = displayWorkItem(workItem.id, workItems, language);
  if (display.taskName) {
    return `${display.projectName} - ${display.taskName}`;
  }
  return display.projectName;
}

function formatGroupedWorkItemOption(workItem: WorkItem, workItems: WorkItem[], language: DateLanguage): string {
  const path = workItemPath(workItem.id, workItems);
  if (path.length <= 1) {
    return labelWorkItemName(workItem.name, language);
  }
  return path.slice(1).map((pathItem) => labelWorkItemName(pathItem.name, language)).join(" - ");
}

function sortWorkItemsByLabel(workItems: WorkItem[], language: DateLanguage): WorkItem[] {
  return [...workItems].sort((a, b) => compareWorkItemLabels(a, b, language));
}

function compareWorkItemLabels(a: WorkItem, b: WorkItem, language: DateLanguage): number {
  return labelWorkItemName(a.name, language).localeCompare(labelWorkItemName(b.name, language), language);
}

function formatStopwatchDuration(stopwatch: Stopwatch, nowTimestamp: number): string {
  if (stopwatch.running) {
    const start = new Date(`${stopwatch.startDate}T${stopwatch.startTime}:00`).getTime();
    if (!Number.isNaN(start)) {
      return formatSeconds((nowTimestamp - start) / 1000);
    }
  }
  return formatSeconds(stopwatch.durationSeconds);
}

function formatSeconds(total: number): string {
  const seconds = Math.max(0, Math.floor(total));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours === 0) return `${minutes}m`;
  return `${hours}h ${String(minutes).padStart(2, "0")}m`;
}
