import type { Page } from "@playwright/test";

export type MockOptions = {
  reportPreferences?: Record<string, unknown> | null;
};

export async function installHumbleBeeMock(page: Page, options: MockOptions = {}) {
  await page.addInitScript(({ reportPreferences }) => {
    let databasePath = "/test/humblebee-accounting.db";
    const workItems = [
      { id: 1, name: "Accounting", parentId: null, depth: 0, status: "ACTIVE" },
      { id: 2, name: "Monthly reconciliation", parentId: 1, depth: 1, status: "ACTIVE" },
      { id: 3, name: "Advisory", parentId: null, depth: 0, status: "ACTIVE" },
      { id: 4, name: "Client meeting", parentId: 3, depth: 1, status: "ACTIVE" }
    ];
    const entry = {
      id: 1,
      workItemId: 2,
      description: "Original note",
      startDate: "2026-09-20",
      endDate: "2026-09-20",
      startTime: "08:00",
      endTime: "09:00",
      durationSeconds: 3600
    };
    const state = {
      backupRequests: 0,
      createdEntries: [] as Record<string, unknown>[],
      createdProjects: [] as Record<string, unknown>[],
      createdTasks: [] as Record<string, unknown>[],
      deletedEntries: [] as number[],
      discardedStopwatches: [] as number[],
      exportRequests: [] as Record<string, unknown>[],
      importedFiles: [] as string[],
      newsRequests: 0,
      openedURLs: [] as string[],
      savedManualSelections: [] as Record<string, unknown>[],
      selectedDatabasePaths: [] as string[],
      startedWorkItems: [] as number[],
      stoppedStopwatches: 0,
      updatedEntries: [] as Record<string, unknown>[],
      preferences: {
        workspaceKey: "accounting-profile",
        reports: reportPreferences ?? null,
        manualSelection: { projectId: 1, taskId: 2 }
      }
    };
    Object.defineProperty(window, "__humblebeeTest", { value: state, configurable: false });

    const emptyReport = { rows: [], empty: true, totalDuration: "00:00" };
    const stopwatches: Record<string, unknown>[] = [];
    const timeDay = (date: string) => ({
      date,
      entries: [entry],
      totalSeconds: 3600,
      projectSeconds: 3600,
      absenceSeconds: 0,
      workSeconds: 3600,
      breakSeconds: 0
    });
    window.go = {
      guiapp: {
        App: {
          BackupDatabase: async () => {
            state.backupRequests += 1;
            return "/test/backups/humblebee-accounting-20260920.db";
          },
          CreateTimeEntry: async (payload: Record<string, unknown>) => { state.createdEntries.push(payload); },
          CreateProject: async (name: string) => {
            state.createdProjects.push({ name, sourceProjectId: 0 });
            return { id: 5, name, parentId: null, depth: 0, status: "ACTIVE" };
          },
          CreateProjectWithTasks: async (name: string, sourceProjectId: number) => {
            state.createdProjects.push({ name, sourceProjectId });
            return { id: 5, name, parentId: null, depth: 0, status: "ACTIVE" };
          },
          CreateTask: async (projectId: number, name: string) => { state.createdTasks.push({ projectId, name }); },
          DeleteTimeEntry: async (id: number) => { state.deletedEntries.push(id); },
          DiscardStopwatch: async (id: number) => {
            state.discardedStopwatches.push(id);
            const index = stopwatches.findIndex((stopwatch) => stopwatch.id === id);
            if (index >= 0) stopwatches.splice(index, 1);
          },
          ExportTimesheetReport: async (request: Record<string, unknown>) => { state.exportRequests.push({ report: "timesheet", request }); return "/test/timesheet.xlsx"; },
          ExportWorktimeByMonthReport: async (request: Record<string, unknown>) => { state.exportRequests.push({ report: "worktime-by-month", request }); return "/test/month.xlsx"; },
          ExportWorktimeGroupedByProjectReport: async (request: Record<string, unknown>) => { state.exportRequests.push({ report: "worktime-grouped-by-project", request }); return "/test/project.xlsx"; },
          ExportWorktimeProjectDetailsReport: async (request: Record<string, unknown>) => { state.exportRequests.push({ report: "worktime-project-details", request }); return "/test/project-details.xlsx"; },
          ExportWorktimeTaskDetailsReport: async (request: Record<string, unknown>) => { state.exportRequests.push({ report: "worktime-task-details", request }); return "/test/task-details.xlsx"; },
          GetDashboard: async () => ({ initialized: true, dbPath: databasePath, userEmail: "local@example.test", todayTotalSeconds: 3600 }),
          GetDatabaseInfo: async () => ({ path: databasePath, defaultPath: "/test/default.db" }),
          GetNews: async (language: string) => {
            state.newsRequests += 1;
            return {
              items: [{
                guid: `${language}-news`,
                title: "HumbleBee update",
                url: `https://www.timeandbill.de/${language}/humblebee/news/`,
                publishedAt: "2026-09-20T08:00:00Z",
                summary: "A local-first time-tracking update."
              }],
              fetchedAt: "2026-09-20T08:00:00Z",
              cached: false,
              unavailable: false,
              cacheWriteFailed: false
            };
          },
          GetRecentNotes: async () => ["• Prüfung │ März\nZweite Zeile", "Quarterly review"],
          GetTimesheetReport: async () => emptyReport,
          GetTimeDay: async (date: string) => timeDay(date),
          GetWorktimeByMonthReport: async () => emptyReport,
          GetWorktimeGroupedByProjectReport: async () => emptyReport,
          GetWorktimeProjectDetailsReport: async () => emptyReport,
          GetWorktimeTaskDetailsReport: async () => emptyReport,
          GetWorkspacePreferences: async () => state.preferences,
          ListProjectWorkItems: async () => workItems,
          ImportTimeAndBill: async (path: string) => {
            state.importedFiles.push(path);
            return {
              summary: {
                projectsCreated: 1, projectsMapped: 0, projectsSkipped: 0,
                tasksCreated: 2, tasksMapped: 0, tasksSkipped: 0,
                timeEntriesCreated: 3, timeEntriesSkipped: 1, timeEntryConflicts: 1,
                alreadyImported: false
              },
              conflicts: []
            };
          },
          ListStopwatches: async () => stopwatches,
          ListWorkItems: async () => workItems,
          PreviewTimeAndBillImport: async () => ({
            exportUuid: "test-export", sourceUserEmail: "accountant@example.test", exportedAt: "2026-09-20T08:00:00Z", existingTimeEntryCount: 2,
            summary: {
              projectsCreated: 1, projectsMapped: 0, projectsSkipped: 0,
              tasksCreated: 2, tasksMapped: 0, tasksSkipped: 0,
              timeEntriesCreated: 3, timeEntriesSkipped: 1, timeEntryConflicts: 1,
              alreadyImported: false
            },
            conflicts: [{ timeEntryUuid: "conflict", projectName: "Accounting", taskName: "Monthly reconciliation", start: "2026-09-20T08:00:00Z", end: "2026-09-20T09:00:00Z" }]
          }),
          SaveManualSelectionPreferences: async (_workspaceKey: string, selection: Record<string, unknown>) => {
            state.preferences.manualSelection = selection as { projectId: number; taskId: number };
            state.savedManualSelections.push(selection);
          },
          SaveReportPreferences: async (_workspaceKey: string, preferences: Record<string, unknown>) => {
            state.preferences.reports = preferences;
          },
          SelectDatabaseFile: async () => "/test/client-work.db",
          SelectImportFile: async () => "/test/time-and-bill-export.json",
          SetTaskActive: async () => undefined,
          Start: async (workItemId: number) => {
            state.startedWorkItems.push(workItemId);
            stopwatches.splice(0, stopwatches.length, {
              id: 91, workItemId, workItemName: "Monthly reconciliation", startDate: "2026-09-20", startTime: "08:00", endDate: "", endTime: "", durationSeconds: 0, conflicting: false, running: true
            });
          },
          Stop: async () => {
            state.stoppedStopwatches += 1;
            if (stopwatches[0]) Object.assign(stopwatches[0], { running: false, endDate: "2026-09-20", endTime: "09:00", durationSeconds: 3600 });
          },
          SwitchDatabase: async (path: string) => {
            databasePath = path;
            state.selectedDatabasePaths.push(path);
            state.preferences.workspaceKey = "client-work-profile";
            state.preferences.manualSelection = { projectId: 3, taskId: 4 };
            return { path, defaultPath: "/test/default.db" };
          },
          UpdateTimeEntry: async (payload: Record<string, unknown>) => { state.updatedEntries.push(payload); }
        }
      }
    };
    window.runtime = { BrowserOpenURL: (url: string) => { state.openedURLs.push(url); } };
  }, options);
}

export async function mockState<T>(page: Page): Promise<T> {
  return page.evaluate(() => (window as Window & { __humblebeeTest: T }).__humblebeeTest);
}
