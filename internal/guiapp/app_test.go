package guiapp

import (
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
