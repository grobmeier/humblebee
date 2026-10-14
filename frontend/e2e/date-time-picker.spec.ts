import { expect, test } from "@playwright/test";
import { installHumbleBeeMock } from "./humblebeeMock";

test("date and time pickers survive modal teardown without browser errors", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await installHumbleBeeMock(page);
  await page.goto("/");

  for (let attempt = 0; attempt < 3; attempt++) {
    await page.getByRole("button", { name: "Zeit erfassen" }).click();
    const modal = page.locator(".modal-form");
    await modal.getByPlaceholder("TT.MM.JJJJ").first().click();
    await expect(page.locator(".flatpickr-calendar.open .flatpickr-days")).toBeVisible();
    await modal.getByPlaceholder("HH:MM").first().click();
    await expect(page.locator(".flatpickr-calendar.open .flatpickr-time")).toBeVisible();
    await modal.getByPlaceholder("HH:MM").first().fill("10:15");
    await expect(modal.getByPlaceholder("HH:MM").first()).toHaveValue("10:15");
    await modal.getByRole("checkbox", { name: "Bis Mitternacht?" }).check();
    await expect(modal.getByPlaceholder("TT.MM.JJJJ")).toHaveCount(1);
    await modal.getByRole("button", { name: "x", exact: true }).click();
    await expect(modal).toBeHidden();
    await expect(page.locator(".flatpickr-calendar")).toHaveCount(0);
    await page.evaluate(() => new Promise<void>((resolve) => {
      requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
    }));
  }

  expect(errors).toEqual([]);
});
