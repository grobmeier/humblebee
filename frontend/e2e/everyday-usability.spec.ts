import { expect, test } from "@playwright/test";
import { installHumbleBeeMock, mockState } from "./humblebeeMock";

test.describe("everyday usability", () => {
  test("uses German labels and aligns compact and multiline modal fields", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");

    await expect(page.getByRole("heading", { name: "Zeiteinträge" })).toBeVisible();
    await expect(page.locator('[aria-label="Zeitübersicht"]')).toBeVisible();
    await page.getByRole("button", { name: "Zeit erfassen" }).click();

    const modal = page.locator(".modal-form");
    await expect(modal.getByText("Projekt", { exact: true })).toBeVisible();
    await expect(modal.getByText("Tätigkeit", { exact: true })).toBeVisible();
    const rows = modal.locator(".tab-form-row");
    await expect(rows.nth(3)).toHaveCSS("align-items", "center");
    await expect(rows.nth(4)).toHaveCSS("align-items", "center");
    await expect(rows.nth(5)).toHaveCSS("align-items", "start");

    await modal.getByRole("button", { name: "×" }).click();
    await page.goto("/#reports/worktime-by-month");
    await expect(page.getByRole("option", { name: "März" }).first()).toBeVisible();
  });

  test("prefills and remembers the last successful manual project and task", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");
    await page.getByRole("button", { name: "Zeit erfassen" }).click();

    const modal = page.locator(".modal-form");
    const selects = modal.locator("select");
    await expect(selects.nth(0)).toHaveValue("1");
    await expect(selects.nth(1)).toHaveValue("2");

    await selects.nth(0).selectOption("3");
    await expect(selects.nth(1)).toHaveValue("4");
    await modal.getByRole("button", { name: "Speichern" }).click();
    await expect(modal).toBeHidden();

    const state = await mockState<{ savedManualSelections: { projectId: number; taskId: number }[] }>(page);
    expect(state.savedManualSelections).toEqual([{ projectId: 3, taskId: 4 }]);
  });

  test("reuses a recent Unicode multiline note only after replacement confirmation", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");
    await page.getByRole("button", { name: "Zeiteintrag duplizieren" }).click();

    const modal = page.locator(".modal-form");
    const note = modal.locator("textarea");
    await expect(note).toHaveValue("Original note");
    await modal.locator("summary").click();
    await modal.getByRole("button", { name: "• Prüfung │ März\nZweite Zeile" }).click();
    await expect(note).toHaveValue("Original note");
    await modal.getByRole("button", { name: "Ersetzen" }).click();
    await expect(note).toHaveValue("• Prüfung │ März\nZweite Zeile");

    await modal.getByRole("button", { name: "Speichern" }).click();
    const state = await mockState<{ createdEntries: { id: number; description: string }[] }>(page);
    expect(state.createdEntries).toHaveLength(1);
    expect(state.createdEntries[0]).toMatchObject({ id: 0, description: "• Prüfung │ März\nZweite Zeile" });
  });

  test("requires an active replacement when duplicating an archived task entry", async ({ page }) => {
    await installHumbleBeeMock(page, { archivedEntrySource: true });
    await page.goto("/");
    await page.getByRole("button", { name: "Zeiteintrag duplizieren" }).click();

    const modal = page.locator(".modal-form");
    await expect(modal.getByText("„Closed client - Historical reconciliation“ ist nicht mehr aktiv.")).toBeVisible();
    const selects = modal.locator("select");
    await expect(selects.nth(0)).toHaveValue("0");
    await expect(selects.nth(1)).toBeDisabled();
    await expect(modal.getByRole("button", { name: "Speichern" })).toBeDisabled();

    await selects.nth(0).selectOption("3");
    await expect(selects.nth(1)).toHaveValue("4");
    await modal.getByRole("button", { name: "Speichern" }).click();
    const state = await mockState<{ createdEntries: { workItemId: number }[] }>(page);
    expect(state.createdEntries).toEqual([expect.objectContaining({ workItemId: 4 })]);
  });

  test("restores the saved project report filter without choosing another project", async ({ page }) => {
    await installHumbleBeeMock(page, {
      reportPreferences: {
        report: "worktime-project-details",
        mode: "monthly",
        month: 3,
        startMonth: 3,
        endMonth: 5,
        year: 2026,
        startDate: "2026-03-01",
        endDate: "2026-05-31",
        projectId: 1,
        showDecimal: true
      }
    });
    await page.goto("/#reports");

    await expect(page).toHaveURL(/#reports\/worktime-project-details$/);
    await expect(page.getByRole("heading", { name: "Projektdetails" })).toBeVisible();
    const filters = page.locator(".report-filter-controls select");
    await expect(filters.nth(0)).toHaveValue("1");
    await expect(filters.nth(1)).toHaveValue("3");
    await expect(filters.nth(2)).toHaveValue("5");
    await expect(page.getByRole("button", { name: "0:00" })).toBeVisible();
  });

  test("shows backup success without switching the active database", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");
    await page.getByRole("button", { name: "Datenbank wechseln" }).click();
    await page.getByRole("button", { name: "Datenbank sichern" }).click();
    await expect(page.getByText("/test/backups/humblebee-accounting-20260920.db", { exact: true })).toBeVisible();

    const state = await mockState<{ backupRequests: number }>(page);
    expect(state.backupRequests).toBe(1);
    await expect(page.getByText("humblebee-accounting.db", { exact: true })).toBeVisible();
  });

  test("does not fetch news until the user opens the news panel", async ({ page }) => {
    await installHumbleBeeMock(page);
    await page.goto("/");
    expect((await mockState<{ newsRequests: number }>(page)).newsRequests).toBe(0);

    await page.getByRole("button", { name: "Neuigkeiten von Time & Bill (verbindet mit timeandbill.de)" }).click();
    await expect(page.getByRole("link", { name: "HumbleBee update" })).toBeVisible();
    expect((await mockState<{ newsRequests: number }>(page)).newsRequests).toBe(1);
  });
});
