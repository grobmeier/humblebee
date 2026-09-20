import test from "node:test";
import assert from "node:assert/strict";
import { restoreReportPreferences } from "./reportPreferences.ts";

const project = { id: 4, name: "Project", depth: 0, parentId: null };
const preferences = {
  report: "worktime-project-details", mode: "daily", month: 2, startMonth: 1,
  endMonth: 3, year: 2026, startDate: "2026-02-01", endDate: "2026-02-28",
  projectId: 4, showDecimal: true
};

test("restores complete valid preferences without mutating input", () => {
  assert.deepEqual(restoreReportPreferences(preferences, [project]), preferences);
  assert.notEqual(restoreReportPreferences(preferences, [project]), preferences);
});

test("deleted project resets without choosing a replacement", () => {
  assert.equal(restoreReportPreferences(preferences, [{ ...project, id: 5 }]).projectId, 0);
  assert.equal(restoreReportPreferences(preferences, [{ ...project, parentId: 1 }]).projectId, 0);
});

test("malformed or outdated values retain normal defaults", () => {
  for (const value of [null, undefined, "bad", {}, [],
    { ...preferences, report: "old-report" }, { ...preferences, mode: "annual" },
    { ...preferences, month: 13 }, { ...preferences, startMonth: 4 },
    { ...preferences, year: 0 }, { ...preferences, year: "2026" },
    { ...preferences, projectId: -1 }, { ...preferences, projectId: 1.5 },
    { ...preferences, startDate: "2026-02-30" }, { ...preferences, endDate: "2026-01-01" },
    { ...preferences, showDecimal: "true" }]) {
    assert.equal(restoreReportPreferences(value, [project]), null, JSON.stringify(value));
  }
});
