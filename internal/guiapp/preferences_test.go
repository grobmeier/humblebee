package guiapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
)

func testReportPreferences() ReportPreferences {
	return ReportPreferences{Report: "timesheet", Mode: "monthly", Month: 2, StartMonth: 1, EndMonth: 3, Year: 2026, StartDate: "2026-01-01", EndDate: "2026-03-31", ShowDecimal: true}
}

func preferenceTestApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())
	app := New()
	if err := app.Init("preferences@example.com"); err != nil {
		t.Fatal(err)
	}
	return app
}

func readTestPreferences(t *testing.T, app *App) *WorkspacePreferences {
	t.Helper()
	p, err := app.GetWorkspacePreferences()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPreferencesRoundTripAndDatabaseIsolation(t *testing.T) {
	app := preferenceTestApp(t)
	original := readTestPreferences(t, app)
	if original.Reports != nil || original.ManualSelection != nil {
		t.Fatal("expected defaults")
	}
	report := testReportPreferences()
	if err := app.SaveReportPreferences(original.WorkspaceKey, report); err != nil {
		t.Fatal(err)
	}
	loaded := readTestPreferences(t, New())
	if loaded.Reports == nil || *loaded.Reports != report {
		t.Fatalf("round trip: %#v", loaded.Reports)
	}
	path := filepath.Join(t.TempDir(), "second.db")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := app.setSelectedDatabasePath(path); err != nil {
		t.Fatal(err)
	}
	if err := app.Init("preferences@example.com"); err != nil {
		t.Fatal(err)
	}
	second := readTestPreferences(t, app)
	if second.WorkspaceKey == original.WorkspaceKey || second.Reports != nil {
		t.Fatal("database preferences leaked")
	}
	if err := app.SaveReportPreferences(original.WorkspaceKey, report); err == nil {
		t.Fatal("stale write accepted")
	}
	if readTestPreferences(t, app).Reports != nil {
		t.Fatal("stale write changed current workspace")
	}
	if err := app.clearSelectedDatabasePath(); err != nil {
		t.Fatal(err)
	}
	if *readTestPreferences(t, app).Reports != report {
		t.Fatal("original preferences lost")
	}
}

func TestPreferencesProfileIsolation(t *testing.T) {
	app := preferenceTestApp(t)
	before := readTestPreferences(t, app)
	if err := app.SaveReportPreferences(before.WorkspaceKey, testReportPreferences()); err != nil {
		t.Fatal(err)
	}
	database, _, err := app.openDB()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := repo.NewPersonRepo(database).CreateDefault(model.Person{UUID: "second-profile", Email: "second@example.com"}); err != nil {
		t.Fatal(err)
	}
	after := readTestPreferences(t, app)
	if before.WorkspaceKey == after.WorkspaceKey || after.Reports != nil {
		t.Fatal("profile preferences leaked")
	}
	if err := app.SaveReportPreferences(before.WorkspaceKey, testReportPreferences()); err == nil {
		t.Fatal("stale profile write accepted")
	}
}

func TestPreferencesValidateSelectionsAndPreserveSections(t *testing.T) {
	app := preferenceTestApp(t)
	project, err := app.CreateProject("Project")
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateTask(project.ID, "Task")
	if err != nil {
		t.Fatal(err)
	}
	key := readTestPreferences(t, app).WorkspaceKey
	manual := ManualSelectionPreferences{ProjectID: project.ID, TaskID: task.ID}
	if err := app.SaveManualSelectionPreferences(key, manual); err != nil {
		t.Fatal(err)
	}
	report := testReportPreferences()
	report.ProjectID = project.ID
	if err := app.SaveReportPreferences(key, report); err != nil {
		t.Fatal(err)
	}
	loaded := readTestPreferences(t, app)
	if loaded.ManualSelection == nil || *loaded.ManualSelection != manual || loaded.Reports.ProjectID != project.ID {
		t.Fatal("sections not preserved")
	}
	if _, err := app.SetTaskActive(task.ID, false); err != nil {
		t.Fatal(err)
	}
	loaded = readTestPreferences(t, app)
	if loaded.ManualSelection == nil || loaded.ManualSelection.ProjectID != project.ID || loaded.ManualSelection.TaskID != 0 {
		t.Fatal("inactive task must reset, retaining project")
	}
	if _, err := app.SetProjectActive(project.ID, false); err != nil {
		t.Fatal(err)
	}
	if readTestPreferences(t, app).ManualSelection != nil {
		t.Fatal("inactive project must reset")
	}
	report.ProjectID = 999999
	if err := app.SaveReportPreferences(key, report); err != nil {
		t.Fatal(err)
	}
	if readTestPreferences(t, app).Reports.ProjectID != 0 {
		t.Fatal("missing report project must reset")
	}
}

func TestPreferencesMalformedAndOutdatedSettings(t *testing.T) {
	app := preferenceTestApp(t)
	key := readTestPreferences(t, app).WorkspaceKey
	for _, body := range []string{`null`, `"bad"`, `[]`, `{"other":false}`} {
		if err := writeGUISettings(guiSettings{WorkspacePreferences: json.RawMessage(body)}); err != nil {
			t.Fatal(err)
		}
		if readTestPreferences(t, app).Reports != nil {
			t.Fatal("malformed preference restored")
		}
	}
	for _, stored := range []storedPreferences{
		{Version: 99, Reports: json.RawMessage(`{}`)},
		{Version: 1, Reports: json.RawMessage(`"bad"`)},
		{Version: 1, Reports: json.RawMessage(`{"report":"timesheet","year":2026}`)},
	} {
		body, _ := json.Marshal(map[string]storedPreferences{key: stored})
		if err := writeGUISettings(guiSettings{WorkspacePreferences: body}); err != nil {
			t.Fatal(err)
		}
		if readTestPreferences(t, app).Reports != nil {
			t.Fatal("invalid report restored")
		}
	}
	invalid := testReportPreferences()
	invalid.StartDate = "2026-02-30"
	if err := app.SaveReportPreferences(key, invalid); err == nil {
		t.Fatal("invalid date accepted")
	}
	invalid = testReportPreferences()
	invalid.EndMonth = 0
	if err := app.SaveReportPreferences(key, invalid); err == nil {
		t.Fatal("invalid month accepted")
	}
	path, _ := guiSettingsPath()
	if err := os.WriteFile(path, []byte(`{"selectedDatabasePath":`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := app.GetWorkspacePreferences(); err == nil {
		t.Fatal("corrupt selection must not silently select a different database")
	}
}

func TestPreferenceWorkspaceKeyResolvesAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.db")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias.db")
	if err := os.Symlink(path, alias); err != nil {
		t.Skip(err)
	}
	a, err := preferenceWorkspaceKey(path, 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := preferenceWorkspaceKey(alias, 1)
	if err != nil || a != b {
		t.Fatal("alias did not resolve to same workspace")
	}
}

func TestPreferencesConcurrentSwitchNeverWritesNewDatabase(t *testing.T) {
	app := preferenceTestApp(t)
	key := readTestPreferences(t, app).WorkspaceKey
	other := filepath.Join(t.TempDir(), "other.db")
	if err := os.WriteFile(other, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := app.setSelectedDatabasePath(other); err != nil {
		t.Fatal(err)
	}
	if err := app.Init("other@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := app.clearSelectedDatabasePath(); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() { defer workers.Done(); _ = app.SaveReportPreferences(key, testReportPreferences()) }()
	}
	if err := app.setSelectedDatabasePath(other); err != nil {
		t.Fatal(err)
	}
	workers.Wait()
	if readTestPreferences(t, app).Reports != nil {
		t.Fatal("old workspace write leaked across concurrent switch")
	}
}

func TestPreferencesBoundedStorage(t *testing.T) {
	app := preferenceTestApp(t)
	key := readTestPreferences(t, app).WorkspaceKey
	values := map[string]storedPreferences{}
	for i := range 256 {
		values[fmt.Sprint(i)] = storedPreferences{Version: 1}
	}
	body, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeGUISettings(guiSettings{WorkspacePreferences: body}); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReportPreferences(key, testReportPreferences()); err == nil {
		t.Fatal("unbounded workspace growth")
	}
	if err := app.SaveReportPreferences("", testReportPreferences()); err == nil {
		t.Fatal("empty workspace key accepted")
	}
}
