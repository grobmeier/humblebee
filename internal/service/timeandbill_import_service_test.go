package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
	_ "modernc.org/sqlite"
)

func TestTimeAndBillImportCreatesRecordsAndTracksExport(t *testing.T) {
	database, personID := setupImportTestDB(t)
	importer := NewTimeAndBillImportService(database)
	importer.now = func() time.Time { return time.Unix(1774512000, 0) }

	summary, err := importer.Import(personID, testTimeAndBillExport(), TimeAndBillImportOptions{AssumeYes: true})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ProjectsCreated != 1 || summary.TasksCreated != 1 || summary.TimeEntriesCreated != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}

	assertCount(t, database, `SELECT count(*) FROM import_runs WHERE export_uuid = 'export-1'`, 1)
	assertCount(t, database, `SELECT count(*) FROM external_mappings`, 3)
	assertCount(t, database, `SELECT count(*) FROM workitems WHERE name = 'Project A'`, 1)
	assertCount(t, database, `SELECT count(*) FROM workitems WHERE name = 'Task A'`, 1)
	assertCount(t, database, `SELECT count(*) FROM time_entries`, 1)

	reimport, err := importer.Import(personID, testTimeAndBillExport(), TimeAndBillImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reimport.AlreadyImported || reimport.TimeEntriesCreated != 0 {
		t.Fatalf("expected duplicate export detection, got %#v", reimport)
	}
}

func TestTimeAndBillImportUpdatesOnlyExistingTimeEntries(t *testing.T) {
	database, personID := setupImportTestDB(t)
	importer := NewTimeAndBillImportService(database)
	importer.now = func() time.Time { return time.Unix(1774512000, 0) }

	if _, err := importer.Import(personID, testTimeAndBillExport(), TimeAndBillImportOptions{AssumeYes: true}); err != nil {
		t.Fatal(err)
	}

	updated := testTimeAndBillExport()
	description := "Updated entry"
	updated.TimeEntries[0].Description = &description
	end := "2026-03-26T10:30:00Z"
	updated.TimeEntries[0].End = &end
	duration := int64(5400)
	updated.TimeEntries[0].DurationSeconds = &duration

	summary, err := importer.Import(personID, updated, TimeAndBillImportOptions{UpdateExisting: true, AssumeYes: true})
	if err != nil {
		t.Fatal(err)
	}
	if summary.TimeEntriesUpdated != 1 || summary.ProjectsSkipped != 1 || summary.TasksSkipped != 1 {
		t.Fatalf("unexpected update summary: %#v", summary)
	}

	var storedDescription string
	var storedDuration int64
	if err := database.QueryRow(`SELECT description, duration FROM time_entries LIMIT 1`).Scan(&storedDescription, &storedDuration); err != nil {
		t.Fatal(err)
	}
	if storedDescription != "Updated entry" || storedDuration != 5400 {
		t.Fatalf("expected updated time entry, got description=%q duration=%d", storedDescription, storedDuration)
	}
}

func TestTimeAndBillImportDryRunRequiresConfirmationForProjectNameMatch(t *testing.T) {
	database, personID := setupImportTestDB(t)
	workItems := repo.NewWorkItemRepo(database)
	if _, err := workItems.Create(repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     "local-project",
		Name:     "Project A",
		Created:  1774512000,
	}); err != nil {
		t.Fatal(err)
	}

	importer := NewTimeAndBillImportService(database)
	summary, err := importer.Import(personID, testTimeAndBillExport(), TimeAndBillImportOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if summary.NeedsConfirmation != 1 {
		t.Fatalf("expected one project confirmation, got %#v", summary)
	}
}

func setupImportTestDB(t *testing.T) (*sql.DB, int64) {
	t.Helper()
	database, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	personID, err := repo.NewPersonRepo(database).CreateDefault(model.Person{
		UUID:      "person-1",
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: 1774512000,
	})
	if err != nil {
		t.Fatal(err)
	}
	return database, personID
}

func testTimeAndBillExport() timeAndBillExport {
	description := "Initial entry"
	end := "2026-03-26T10:00:00Z"
	duration := int64(3600)
	return timeAndBillExport{
		SchemaVersion: 1,
		Format:        timeAndBillExportFormat,
		ExportUUID:    "export-1",
		ExportedAt:    "2026-03-26T10:00:00Z",
		User: timeAndBillExportUser{
			UUID:  "tab-user-1",
			Email: "user@example.com",
		},
		Projects: []timeAndBillExportProject{{
			UUID:   "project-1",
			Name:   "Project A",
			Active: true,
		}},
		Tasks: []timeAndBillExportTask{{
			UUID:        "task-1",
			ProjectUUID: "project-1",
			Name:        "Task A",
			Active:      true,
		}},
		TimeEntries: []timeAndBillExportEntry{{
			UUID:            "time-1",
			TaskUUID:        "task-1",
			ProjectUUID:     "project-1",
			Description:     &description,
			Start:           "2026-03-26T09:00:00Z",
			End:             &end,
			DurationSeconds: &duration,
			Timezone:        "Europe/Berlin",
		}},
	}
}

func assertCount(t *testing.T, database *sql.DB, query string, expected int) {
	t.Helper()
	var count int
	if err := database.QueryRow(query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("expected count %d for %s, got %d", expected, query, count)
	}
}
