// Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package guiapp

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
)

func TestStartSwitchesFromRunningStopwatchToNewStopwatch(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	workItemID := createGUIAppTestWorkItem(t, app, "Client work")

	if err := app.Start(0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	if err := app.Start(workItemID); err != nil {
		t.Fatal(err)
	}

	stopwatches, err := app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 2 {
		t.Fatalf("expected stopped old stopwatch and running new stopwatch, got %d", len(stopwatches))
	}
	if !stopwatches[0].Running {
		t.Fatal("expected newest stopwatch to be running")
	}
	if stopwatches[0].WorkItemID == nil || *stopwatches[0].WorkItemID != workItemID {
		t.Fatalf("expected running stopwatch work item %d, got %#v", workItemID, stopwatches[0].WorkItemID)
	}
	if stopwatches[1].Running {
		t.Fatal("expected previous stopwatch to be booked")
	}
	if stopwatches[1].DurationSeconds <= 0 {
		t.Fatalf("expected booked stopwatch duration, got %d", stopwatches[1].DurationSeconds)
	}
}

func TestStartSwitchReturnsOverlapModalErrorWhenRunningStopwatchConflicts(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	workItemID := createGUIAppTestWorkItem(t, app, "Client work")

	if err := app.Start(0); err != nil {
		t.Fatal(err)
	}
	runningStart := runningStopwatchStart(t, app)
	createGUIAppTestCompletedEntry(t, app, workItemID, runningStart, runningStart+120)
	time.Sleep(1100 * time.Millisecond)

	err := app.Start(workItemID)
	if err == nil {
		t.Fatal("expected stopwatch overlap error")
	}
	if !strings.Contains(err.Error(), stopwatchOverlapErrorCode) {
		t.Fatalf("expected structured stopwatch overlap error, got %v", err)
	}

	stopwatches, err := app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 2 {
		t.Fatalf("expected the new running stopwatch and old conflicting stopwatch, got %d", len(stopwatches))
	}
	if !stopwatches[0].Running {
		t.Fatal("expected new stopwatch to be running after conflict")
	}
	if stopwatches[0].WorkItemID == nil || *stopwatches[0].WorkItemID != workItemID {
		t.Fatalf("expected running stopwatch work item %d, got %#v", workItemID, stopwatches[0].WorkItemID)
	}
	if !stopwatches[1].Conflicting {
		t.Fatal("expected old stopwatch to stay visible as conflicting")
	}
	if stopwatches[1].Running {
		t.Fatal("expected conflicting stopwatch not to be running")
	}
}

func TestDeleteTimeEntryDiscardsStopwatchWithoutBooking(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := app.Start(0); err != nil {
		t.Fatal(err)
	}
	stopwatches, err := app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 1 {
		t.Fatalf("expected running stopwatch, got %d", len(stopwatches))
	}

	if err := app.DeleteTimeEntry(stopwatches[0].ID); err != nil {
		t.Fatal(err)
	}
	stopwatches, err = app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 0 {
		t.Fatalf("expected discarded stopwatch not to be listed, got %d", len(stopwatches))
	}
}

func TestDeleteTimeEntryUnbooksStoppedStopwatchButKeepsStopwatchCard(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := app.Start(0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := app.Stop(); err != nil {
		t.Fatal(err)
	}
	stopwatches, err := app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 1 {
		t.Fatalf("expected stopped stopwatch, got %d", len(stopwatches))
	}

	if err := app.DeleteTimeEntry(stopwatches[0].ID); err != nil {
		t.Fatal(err)
	}
	stopwatches, err = app.ListStopwatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 1 {
		t.Fatalf("expected stopwatch card to remain after unbooking, got %d", len(stopwatches))
	}
	if stopwatches[0].Running {
		t.Fatal("expected unbooked stopwatch card to remain stopped")
	}
	day, err := app.GetTimeDay(stopwatches[0].StartDate)
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 0 {
		t.Fatalf("expected unbooked stopwatch not to be listed as time entry, got %d", len(day.Entries))
	}
}

func TestDeleteTimeEntryDeletesCompletedEntry(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	workItemID := createGUIAppTestWorkItem(t, app, "Client work")
	entry, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  workItemID,
		StartDate:   "2026-05-12",
		StartTime:   "09:00",
		EndDate:     "2026-05-12",
		EndTime:     "10:00",
		Description: "Entry to delete",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := app.DeleteTimeEntry(entry.ID); err != nil {
		t.Fatal(err)
	}
	day, err := app.GetTimeDay("2026-05-12")
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 0 {
		t.Fatalf("expected deleted time entry not to be listed, got %d", len(day.Entries))
	}
}

func TestCreateProjectUpdateProjectAndCreateTask(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}

	project, err := app.CreateProject("Client A")
	if err != nil {
		t.Fatal(err)
	}
	if project.ID == 0 || project.Name != "Client A" || project.ParentID != nil || project.Depth != 0 {
		t.Fatalf("unexpected project: %#v", project)
	}

	updated, err := app.UpdateProject(project.ID, "Client B")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Client B" {
		t.Fatalf("expected renamed project, got %#v", updated)
	}

	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if task.ParentID == nil || *task.ParentID != project.ID || task.Depth != 1 {
		t.Fatalf("unexpected task: %#v", task)
	}

	items, err := app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if !containsGUIWorkItem(items, project.ID, "Client B") {
		t.Fatalf("expected renamed project in work item list: %#v", items)
	}
	if !containsGUIWorkItem(items, task.ID, "Research") {
		t.Fatalf("expected task in work item list: %#v", items)
	}
}

func TestCreateProjectWithTasksCopiesOnlyActiveTasks(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	source, err := app.CreateProject("Template")
	if err != nil {
		t.Fatal(err)
	}
	activeTask, err := app.CreateTask(source.ID, "Accounting")
	if err != nil {
		t.Fatal(err)
	}
	completedTask, err := app.CreateTask(source.ID, "Old setup")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTaskActive(completedTask.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  activeTask.ID,
		StartDate:   "2026-06-04",
		StartTime:   "09:00",
		EndDate:     "2026-06-04",
		EndTime:     "10:00",
		Description: "Template history",
	}); err != nil {
		t.Fatal(err)
	}

	target, err := app.CreateProjectWithTasks("New client", source.ID)
	if err != nil {
		t.Fatal(err)
	}

	items, err := app.ListProjectWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if !containsChildWorkItem(items, target.ID, "Accounting") {
		t.Fatalf("expected copied active task under new project: %#v", items)
	}
	if containsChildWorkItem(items, target.ID, "Old setup") {
		t.Fatalf("expected completed task not to be copied: %#v", items)
	}
	day, err := app.GetTimeDay("2026-06-04")
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 1 {
		t.Fatalf("expected source time entry to remain single and not be copied, got %#v", day.Entries)
	}
	if day.Entries[0].WorkItemID == nil || *day.Entries[0].WorkItemID != activeTask.ID {
		t.Fatalf("expected original time entry to stay on source task, got %#v", day.Entries[0])
	}
}

func TestCreateProjectWithTasksRejectsInvalidSourceProject(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	source, err := app.CreateProject("Template")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(source.ID, "Accounting")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := app.CreateProjectWithTasks("New client", task.ID); err == nil {
		t.Fatal("expected task source to be rejected")
	}
	if _, err := app.CreateProjectWithTasks("New client", 999999); err == nil {
		t.Fatal("expected missing source project to be rejected")
	}
}

func TestCreateProjectWithTasksRollsBackProjectWhenTaskCopyFails(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	source, err := app.CreateProject("Template")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTask(source.ID, "Accounting"); err != nil {
		t.Fatal(err)
	}

	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	personID, err := app.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.NewWorkItemRepo(database).Create(repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     "",
		ParentID: &source.ID,
		Depth:    1,
		Created:  time.Now().UTC().Unix(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := app.CreateProjectWithTasks("New client", source.ID); err == nil {
		t.Fatal("expected invalid copied task name to fail")
	}
	items, err := app.ListProjectWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if containsTopLevelWorkItem(items, "New client") {
		t.Fatalf("expected failed task copy to roll back target project: %#v", items)
	}
}

func TestGetWorktimeByMonthReportShowsMonthlyTimeRows(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  task.ID,
		StartDate:   "2026-06-04",
		StartTime:   "09:00",
		EndDate:     "2026-06-04",
		EndTime:     "10:30",
		Description: "Discovery",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeByMonthReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Empty {
		t.Fatal("expected report with rows")
	}
	if report.TotalSeconds != 5400 || report.TotalDuration != "01:30" {
		t.Fatalf("expected 01:30 total, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected one row, got %#v", report.Rows)
	}
	row := report.Rows[0]
	if row.ProjectName != "Client" || row.TaskName != "Research" || row.Description != "Discovery" {
		t.Fatalf("unexpected row labels: %#v", row)
	}
	if row.Date != "2026-06-04" || row.StartTime != "09:00" || row.EndTime != "10:30" || row.Duration != "01:30" {
		t.Fatalf("unexpected row times: %#v", row)
	}
}

func TestGetWorktimeByMonthReportSupportsMonthlyRange(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	for _, startDate := range []string{"2026-05-20", "2026-06-04", "2026-07-01"} {
		if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
			WorkItemID: task.ID,
			StartDate:  startDate,
			StartTime:  "09:00",
			EndDate:    startDate,
			EndTime:    "10:00",
		}); err != nil {
			t.Fatal(err)
		}
	}

	report, err := app.GetWorktimeByMonthReport(ReportRequest{
		Mode:       "monthly",
		StartMonth: 5,
		EndMonth:   6,
		Year:       2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalSeconds != 7200 || report.TotalDuration != "02:00" {
		t.Fatalf("expected May and June total only, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
	if len(report.Rows) != 2 {
		t.Fatalf("expected two rows for May and June, got %#v", report.Rows)
	}
}

func TestGetWorktimeProjectDetailsReportShowsSelectedProjectEntries(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	client, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	research, err := app.CreateTask(client.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	internal, err := app.CreateProject("Internal")
	if err != nil {
		t.Fatal(err)
	}
	planning, err := app.CreateTask(internal.ID, "Planning")
	if err != nil {
		t.Fatal(err)
	}
	note := "Line one │ äöü & <tag>\nLine two"
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  research.ID,
		StartDate:   "2026-06-04",
		StartTime:   "09:00",
		EndDate:     "2026-06-04",
		EndTime:     "10:30",
		Description: note,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: planning.ID,
		StartDate:  "2026-06-04",
		StartTime:  "11:00",
		EndDate:    "2026-06-04",
		EndTime:    "12:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeProjectDetailsReport(ReportRequest{
		Mode:      "monthly",
		Month:     6,
		Year:      2026,
		ProjectID: client.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Empty {
		t.Fatal("expected project detail rows")
	}
	if report.TotalSeconds != 5400 || report.TotalDuration != "01:30" {
		t.Fatalf("expected 01:30 total, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected only selected project row, got %#v", report.Rows)
	}
	row := report.Rows[0]
	if row.ProjectName != "Client" || row.TaskName != "Research" {
		t.Fatalf("unexpected project detail row labels: %#v", row)
	}
	if row.Description != note {
		t.Fatalf("expected note to be preserved, got %q", row.Description)
	}
}

func TestGetWorktimeProjectDetailsReportWithoutProjectHasZeroDuration(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeProjectDetailsReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Empty {
		t.Fatalf("expected report without selected project to be empty: %#v", report)
	}
	if report.TotalSeconds != 0 || report.TotalDuration != "00:00" {
		t.Fatalf("expected zero total duration, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
}

func TestGetWorktimeProjectDetailsReportAppliesDateRange(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: task.ID,
		StartDate:  "2026-06-03",
		StartTime:  "09:00",
		EndDate:    "2026-06-03",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: task.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "11:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeProjectDetailsReport(ReportRequest{
		Mode:      "daily",
		StartDate: "2026-06-04",
		EndDate:   "2026-06-04",
		ProjectID: project.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected one date range row, got %#v", report.Rows)
	}
	if report.Rows[0].Date != "2026-06-04" || report.TotalDuration != "02:00" {
		t.Fatalf("unexpected date range report: %#v", report)
	}
}

func TestReportWorkItemLabelLocalizesTimeAndBillReservedNames(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{name: "@", language: "de", want: "Abwesenheiten"},
		{name: "@break", language: "de", want: "Pause"},
		{name: "@overtime", language: "de", want: "Überstundenausgleich"},
		{name: "@public_holiday", language: "en", want: "Public holiday"},
		{name: "@sick_leave", language: "en", want: "Sick leave"},
		{name: "@vacation", language: "en", want: "Vacation"},
		{name: "Client", language: "de", want: "Client"},
	}

	for _, tt := range tests {
		if got := reportWorkItemLabel(tt.name, tt.language); got != tt.want {
			t.Fatalf("reportWorkItemLabel(%q, %q) = %q, want %q", tt.name, tt.language, got, tt.want)
		}
	}
}

func TestGetWorktimeGroupedByProjectReportGroupsRowsAndTotals(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	client, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	research, err := app.CreateTask(client.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	internal, err := app.CreateProject("Internal")
	if err != nil {
		t.Fatal(err)
	}
	planning, err := app.CreateTask(internal.ID, "Planning")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  research.ID,
		StartDate:   "2026-06-04",
		StartTime:   "09:00",
		EndDate:     "2026-06-04",
		EndTime:     "10:00",
		Description: "Discovery",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: planning.ID,
		StartDate:  "2026-06-04",
		StartTime:  "10:30",
		EndDate:    "2026-06-04",
		EndTime:    "11:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeGroupedByProjectReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Empty {
		t.Fatal("expected grouped report with rows")
	}
	if len(report.Groups) != 2 {
		t.Fatalf("expected two project groups, got %#v", report.Groups)
	}
	if report.Groups[0].ProjectName != "Client" || report.Groups[0].TotalDuration != "01:00" || len(report.Groups[0].Rows) != 1 {
		t.Fatalf("unexpected first group: %#v", report.Groups[0])
	}
	if report.Groups[1].ProjectName != "Internal" || report.Groups[1].TotalDuration != "00:30" || len(report.Groups[1].Rows) != 1 {
		t.Fatalf("unexpected second group: %#v", report.Groups[1])
	}
}

func TestGetWorktimeTaskDetailsReportAggregatesSelectedProjectTasks(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	research, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: research.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: research.ID,
		StartDate:  "2026-06-04",
		StartTime:  "10:30",
		EndDate:    "2026-06-04",
		EndTime:    "11:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeTaskDetailsReport(ReportRequest{
		Mode:      "monthly",
		Month:     6,
		Year:      2026,
		ProjectID: project.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Empty {
		t.Fatal("expected task detail rows")
	}
	if report.TotalSeconds != 5400 || report.TotalDuration != "01:30" {
		t.Fatalf("expected 01:30 total, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected one task aggregate row, got %#v", report.Rows)
	}
	row := report.Rows[0]
	if row.ProjectName != "Client" || row.TaskName != "Research" || row.Duration != "01:30" {
		t.Fatalf("unexpected task aggregate row: %#v", row)
	}
}

func TestGetWorktimeTaskDetailsReportDefaultsToFirstReportableProject(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	alpha, err := app.CreateProject("Alpha")
	if err != nil {
		t.Fatal(err)
	}
	alphaTask, err := app.CreateTask(alpha.ID, "Discovery")
	if err != nil {
		t.Fatal(err)
	}
	zeta, err := app.CreateProject("Zeta")
	if err != nil {
		t.Fatal(err)
	}
	zetaTask, err := app.CreateTask(zeta.ID, "Maintenance")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: zetaTask.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: alphaTask.ID,
		StartDate:  "2026-06-04",
		StartTime:  "10:30",
		EndDate:    "2026-06-04",
		EndTime:    "11:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetWorktimeTaskDetailsReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected only first reportable project, got %#v", report.Rows)
	}
	if report.Rows[0].ProjectName != "Alpha" || report.Rows[0].Duration != "00:30" {
		t.Fatalf("expected Alpha default project, got %#v", report.Rows[0])
	}
}

func TestGetTimesheetReportShowsMonthlyProjectTotals(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	client, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	research, err := app.CreateTask(client.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	internal, err := app.CreateProject("Internal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: research.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: internal.ID,
		StartDate:  "2026-06-04",
		StartTime:  "10:30",
		EndDate:    "2026-06-04",
		EndTime:    "12:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetTimesheetReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Empty {
		t.Fatal("expected timesheet rows")
	}
	if report.TotalSeconds != 9000 || report.TotalDuration != "02:30" {
		t.Fatalf("expected 02:30 total, got %d %q", report.TotalSeconds, report.TotalDuration)
	}
	if len(report.ProjectRows) != 2 {
		t.Fatalf("expected two project totals, got %#v", report.ProjectRows)
	}
	if report.ProjectRows[0].ProjectName != "Client" || report.ProjectRows[0].Duration != "01:00" {
		t.Fatalf("unexpected first project total: %#v", report.ProjectRows[0])
	}
	if report.ProjectRows[1].ProjectName != "Internal" || report.ProjectRows[1].Duration != "01:30" {
		t.Fatalf("unexpected second project total: %#v", report.ProjectRows[1])
	}
}

func TestGetTimesheetReportSplitsDateRangeTotalsByDay(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: project.ID,
		StartDate:  "2026-06-04",
		StartTime:  "23:00",
		EndDate:    "2026-06-05",
		EndTime:    "01:00",
	}); err != nil {
		t.Fatal(err)
	}

	report, err := app.GetTimesheetReport(ReportRequest{
		Mode:      "daily",
		StartDate: "2026-06-04",
		EndDate:   "2026-06-05",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.DailyRows) != 2 {
		t.Fatalf("expected two daily rows, got %#v", report.DailyRows)
	}
	if report.DailyRows[0].Date != "2026-06-04" || report.DailyRows[0].ProjectDuration != "01:00" {
		t.Fatalf("unexpected first daily row: %#v", report.DailyRows[0])
	}
	if report.DailyRows[1].Date != "2026-06-05" || report.DailyRows[1].ProjectDuration != "01:00" {
		t.Fatalf("unexpected second daily row: %#v", report.DailyRows[1])
	}
	if report.TotalDuration != "02:00" {
		t.Fatalf("expected 02:00 total, got %q", report.TotalDuration)
	}
}

func TestExportWorktimeByMonthReportWritesExcelFile(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: project.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}

	path, err := app.ExportWorktimeByMonthReport(ReportRequest{
		Mode:  "monthly",
		Month: 6,
		Year:  2026,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".xlsx" {
		t.Fatalf("expected .xlsx path, got %q", path)
	}
	worksheet := readXLSXWorksheet(t, path)
	if !strings.Contains(worksheet, "Client") || !strings.Contains(worksheet, "01:00") {
		t.Fatalf("expected worksheet to contain report data, got %s", worksheet)
	}
}

func TestExportWorktimeByMonthReportUsesGermanLabels(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID: project.ID,
		StartDate:  "2026-06-04",
		StartTime:  "09:00",
		EndDate:    "2026-06-04",
		EndTime:    "10:00",
	}); err != nil {
		t.Fatal(err)
	}

	path, err := app.ExportWorktimeByMonthReport(ReportRequest{
		Mode:     "monthly",
		Month:    6,
		Year:     2026,
		Language: "de",
	})
	if err != nil {
		t.Fatal(err)
	}
	worksheet := readXLSXWorksheet(t, path)
	if !strings.Contains(worksheet, "Projekt") || !strings.Contains(worksheet, "Aufgabe") || !strings.Contains(worksheet, "Dauer") {
		t.Fatalf("expected worksheet to contain German labels, got %s", worksheet)
	}
	if strings.Contains(worksheet, "Duration") {
		t.Fatalf("expected worksheet not to contain English duration label, got %s", worksheet)
	}
}

func TestExportWorktimeProjectDetailsReportPreservesMultilineNotes(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	note := "Line one │ äöü & <tag>\nLine two • \"quoted\""
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  task.ID,
		StartDate:   "2026-06-04",
		StartTime:   "09:00",
		EndDate:     "2026-06-04",
		EndTime:     "10:00",
		Description: note,
	}); err != nil {
		t.Fatal(err)
	}

	path, err := app.ExportWorktimeProjectDetailsReport(ReportRequest{
		Mode:      "monthly",
		Month:     6,
		Year:      2026,
		ProjectID: project.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.Base(path), fmt.Sprintf("worktime-project-details-%d", project.ID)) {
		t.Fatalf("expected project-specific export filename, got %q", path)
	}
	worksheet := readXLSXWorksheet(t, path)
	for _, want := range []string{"Line one │ äöü", "Line two • &quot;quoted&quot;", "&amp;", "&lt;tag&gt;"} {
		if !strings.Contains(worksheet, want) {
			t.Fatalf("expected worksheet to contain %q, got %s", want, worksheet)
		}
	}
	if !strings.Contains(worksheet, "xml:space=\"preserve\"") {
		t.Fatalf("expected worksheet cells to preserve whitespace, got %s", worksheet)
	}
}

func TestCreateTaskRejectsTaskParent(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTask(task.ID, "Nested"); err == nil {
		t.Fatal("expected nested task creation to fail")
	}
}

func TestSetTaskActiveHidesAndRestoresTaskInStopwatchWorkItems(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}

	hidden, err := app.SetTaskActive(task.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if hidden.Status != string(model.WorkItemStatusArchived) {
		t.Fatalf("expected archived task, got %#v", hidden)
	}
	activeItems, err := app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if containsGUIWorkItem(activeItems, task.ID, "Research") {
		t.Fatalf("expected archived task to be hidden from active work items: %#v", activeItems)
	}
	projectItems, err := app.ListProjectWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if !containsGUIWorkItem(projectItems, task.ID, "Research") {
		t.Fatalf("expected archived task to stay visible in project management: %#v", projectItems)
	}

	restored, err := app.SetTaskActive(task.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Status != string(model.WorkItemStatusActive) {
		t.Fatalf("expected active task, got %#v", restored)
	}
	activeItems, err = app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if !containsGUIWorkItem(activeItems, task.ID, "Research") {
		t.Fatalf("expected restored task in active work items: %#v", activeItems)
	}
}

func TestListWorkItemsIncludesDisplayParentsAndHidesArchivedParentChildren(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	activeProject, err := app.CreateProject("Active Client")
	if err != nil {
		t.Fatal(err)
	}
	activeTask, err := app.CreateTask(activeProject.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	archivedProject, err := app.CreateProject("Old Client")
	if err != nil {
		t.Fatal(err)
	}
	archivedChild, err := app.CreateTask(archivedProject.ID, "Still Active Child")
	if err != nil {
		t.Fatal(err)
	}
	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	personID, err := app.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.NewWorkItemRepo(database).SetStatus(personID, archivedProject.ID, model.WorkItemStatusArchived); err != nil {
		t.Fatal(err)
	}

	activeItems, err := app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if !containsGUIWorkItem(activeItems, activeProject.ID, "Active Client") {
		t.Fatalf("expected active parent for project-task display context: %#v", activeItems)
	}
	if !containsGUIWorkItem(activeItems, activeTask.ID, "Research") {
		t.Fatalf("expected active task in stopwatch work items: %#v", activeItems)
	}
	if containsGUIWorkItem(activeItems, archivedProject.ID, "Old Client") || containsGUIWorkItem(activeItems, archivedChild.ID, "Still Active Child") {
		t.Fatalf("expected archived parent subtree to be hidden from stopwatch work items: %#v", activeItems)
	}
}

func TestDeleteProjectDeletesTasksAndTimeEntries(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	project, err := app.CreateProject("Client")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Research")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.CreateTimeEntry(CreateTimeEntryRequest{
		WorkItemID:  task.ID,
		StartDate:   "2026-05-12",
		StartTime:   "09:00",
		EndDate:     "2026-05-12",
		EndTime:     "10:00",
		Description: "Project time",
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.DeleteProject(project.ID); err != nil {
		t.Fatal(err)
	}
	activeItems, err := app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if containsGUIWorkItem(activeItems, project.ID, "Client") || containsGUIWorkItem(activeItems, task.ID, "Research") {
		t.Fatalf("expected deleted project subtree to be absent from active work items: %#v", activeItems)
	}
	projectItems, err := app.ListProjectWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	if containsGUIWorkItem(projectItems, project.ID, "Client") || containsGUIWorkItem(projectItems, task.ID, "Research") {
		t.Fatalf("expected deleted project subtree to be absent from project management: %#v", projectItems)
	}
	day, err := app.GetTimeDay("2026-05-12")
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 0 {
		t.Fatalf("expected deleting project to delete its time entries, got %d", len(day.Entries))
	}
}

func TestUpdateProjectRejectsDefaultWorkItem(t *testing.T) {
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())

	app := New()
	if err := app.Init("user@example.com"); err != nil {
		t.Fatal(err)
	}
	items, err := app.ListWorkItems()
	if err != nil {
		t.Fatal(err)
	}
	var defaultID int64
	for _, item := range items {
		if item.Name == "Default" && item.ParentID == nil {
			defaultID = item.ID
		}
	}
	if defaultID == 0 {
		t.Fatal("expected Default work item")
	}
	if _, err := app.UpdateProject(defaultID, "Renamed"); err == nil {
		t.Fatal("expected Default project update to fail")
	}
}

func createGUIAppTestWorkItem(t *testing.T, app *App, name string) int64 {
	t.Helper()
	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	personID, err := app.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	item, err := repo.NewWorkItemRepo(database).Create(repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     name,
		Created:  time.Now().UTC().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return item.ID
}

func containsGUIWorkItem(items []WorkItem, id int64, name string) bool {
	for _, item := range items {
		if item.ID == id && item.Name == name {
			return true
		}
	}
	return false
}

func containsChildWorkItem(items []WorkItem, parentID int64, name string) bool {
	for _, item := range items {
		if item.ParentID != nil && *item.ParentID == parentID && item.Name == name {
			return true
		}
	}
	return false
}

func containsTopLevelWorkItem(items []WorkItem, name string) bool {
	for _, item := range items {
		if item.ParentID == nil && item.Name == name {
			return true
		}
	}
	return false
}

func createGUIAppTestCompletedEntry(t *testing.T, app *App, workItemID int64, start int64, end int64) {
	t.Helper()
	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	personID, err := app.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	duration := end - start
	if _, err := repo.NewTimeEntryRepo(database).CreateCompleted(model.TimeEntry{
		UUID:       uuid.NewString(),
		PersonID:   personID,
		WorkItemID: &workItemID,
		StartTime:  start,
		EndTime:    &end,
		Duration:   &duration,
		CreatedAt:  start,
	}); err != nil {
		t.Fatal(err)
	}
}

func runningStopwatchStart(t *testing.T, app *App) int64 {
	t.Helper()
	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	personID, err := app.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	running, err := repo.NewTimeEntryRepo(database).FindRunning(personID)
	if err != nil {
		t.Fatal(err)
	}
	if running == nil {
		t.Fatal("expected running stopwatch")
	}
	return running.StartTime
}

func readXLSXWorksheet(t *testing.T, path string) string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	t.Fatalf("worksheet not found in %s", path)
	return ""
}
