import { expect, test } from "@playwright/test";
import { installHumbleBeeMock, mockState } from "./humblebeeMock";

test.describe("core workflows", () => {
  test("edits and deletes a booked time entry through the expected bridge calls", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await page.getByRole("button", { name: /08:00 - 09:00/ }).click();
    const modal = page.locator(".modal-form");
    await modal.locator("textarea").fill("Edited note");
    await modal.getByRole("button", { name: "Speichern" }).click();

    let state = await mockState<{ updatedEntries: { id: number; description: string }[] }>(page);
    expect(state.updatedEntries).toEqual([expect.objectContaining({ id: 1, description: "Edited note" })]);

    await page.getByRole("button", { name: "Zeiteintrag löschen" }).click();
    state = await mockState<{ deletedEntries: number[] }>(page);
    expect(state.deletedEntries).toEqual([1]);
  });

  test("starts, stops, and discards a stopwatch", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await page.getByRole("combobox", { name: "Stoppuhr-Aufgabe" }).selectOption("2");
    await page.locator(".stopwatch-create").getByRole("button", { name: "Starten" }).click();
    const timer = page.locator(".timer-card");
    await expect(timer.getByRole("button", { name: "Stoppen" })).toBeVisible();

    await timer.getByRole("button", { name: "Stoppen" }).click();
    await expect(timer.getByRole("button", { name: "Starten" })).toBeVisible();
    await timer.getByRole("button", { name: "Stoppuhr verwerfen" }).click();
    await expect(timer).toHaveCount(0);

    const state = await mockState<{ startedWorkItems: number[]; stoppedStopwatches: number; discardedStopwatches: number[] }>(page);
    expect(state).toMatchObject({ startedWorkItems: [2], stoppedStopwatches: 1, discardedStopwatches: [91] });
  });

  test("creates a project from a task template and adds a task", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/#projects");

    await page.getByRole("button", { name: "Projekt +" }).click();
    let modal = page.locator(".modal-form");
    await modal.locator("input").fill("Year-end review");
    await modal.locator("select").selectOption("1");
    await modal.getByRole("button", { name: "Projekt anlegen" }).click();

    await page.getByRole("button", { name: "Aufgabe +" }).click();
    modal = page.locator(".modal-form");
    await modal.locator("input").fill("VAT filing");
    await modal.getByRole("button", { name: "Aufgabe anlegen" }).click();

    const state = await mockState<{
      createdProjects: { name: string; sourceProjectId: number }[];
      createdTasks: { projectId: number; name: string }[];
    }>(page);
    expect(state.createdProjects).toEqual([{ name: "Year-end review", sourceProjectId: 1 }]);
    expect(state.createdTasks).toEqual([{ projectId: 1, name: "VAT filing" }]);
  });

  test("exports the selected project-details report", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/#reports/worktime-project-details");

    const filters = page.locator(".report-filter-controls select");
    await filters.nth(0).selectOption("1");
    await page.getByRole("button", { name: "Excel exportieren" }).click();
    await expect(page.getByText("/test/project-details.xlsx", { exact: true })).toBeVisible();

    const state = await mockState<{ exportRequests: { report: string; request: { projectId: number } }[] }>(page);
    expect(state.exportRequests).toEqual([expect.objectContaining({
      report: "worktime-project-details",
      request: expect.objectContaining({ projectId: 1 })
    })]);
  });

  test("previews an import before importing it once", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await page.getByRole("button", { name: "Time & Bill importieren" }).click();
    const modal = page.locator(".modal-form");
    await modal.getByRole("button", { name: "Datei wählen" }).click();
    await expect(modal.getByText("/test/time-and-bill-export.json", { exact: true })).toBeVisible();
    await modal.getByRole("button", { name: "Vorschau" }).click();
    await expect(modal.getByText("Anlegen: 3", { exact: true })).toBeVisible();
    await expect(modal.getByText("Diese Datenbank enthält bereits 2 gebuchte Zeiteinträge.")).toBeVisible();
    await modal.getByRole("button", { name: "Importieren" }).click();
    await expect(modal.getByText("Import abgeschlossen.", { exact: true })).toBeVisible();
    await expect(modal.getByRole("button", { name: "Datei wählen" })).toHaveCount(0);

    const state = await mockState<{ importedFiles: string[] }>(page);
    expect(state.importedFiles).toEqual(["/test/time-and-bill-export.json"]);
  });

  test("switches the database and updates its visible identity", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await page.getByRole("button", { name: "Datenbank wechseln" }).click();
    await page.getByRole("button", { name: "Andere Datenbank öffnen" }).click();
    await expect(page.getByText("client-work.db", { exact: true })).toBeVisible();

    const state = await mockState<{ selectedDatabasePaths: string[] }>(page);
    expect(state.selectedDatabasePaths).toEqual(["/test/client-work.db"]);
  });

  test("switches dashboard and modal labels to English", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await page.getByRole("button", { name: "EN", exact: true }).click();
    await expect(page.getByRole("link", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Add time" })).toBeVisible();
    await page.getByRole("button", { name: "Add time" }).click();
    await expect(page.getByRole("heading", { name: "Record time entry" })).toBeVisible();
  });
});
