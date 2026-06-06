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
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DatabaseInfo struct {
	Path        string `json:"path"`
	DefaultPath string `json:"defaultPath"`
	Initialized bool   `json:"initialized"`
}

type ImportPreview struct {
	ExportUUID             string           `json:"exportUuid"`
	ExportedAt             string           `json:"exportedAt"`
	SourceUserEmail        string           `json:"sourceUserEmail"`
	ExistingTimeEntryCount int              `json:"existingTimeEntryCount"`
	Summary                ImportSummary    `json:"summary"`
	Conflicts              []ImportConflict `json:"conflicts"`
}

type ImportResult struct {
	Summary   ImportSummary    `json:"summary"`
	Conflicts []ImportConflict `json:"conflicts"`
}

type ImportSummary struct {
	ExportUUID         string `json:"exportUuid"`
	AlreadyImported    bool   `json:"alreadyImported"`
	ProjectsCreated    int    `json:"projectsCreated"`
	ProjectsMapped     int    `json:"projectsMapped"`
	ProjectsSkipped    int    `json:"projectsSkipped"`
	TasksCreated       int    `json:"tasksCreated"`
	TasksMapped        int    `json:"tasksMapped"`
	TasksSkipped       int    `json:"tasksSkipped"`
	TimeEntriesCreated int    `json:"timeEntriesCreated"`
	TimeEntriesUpdated int    `json:"timeEntriesUpdated"`
	TimeEntriesSkipped int    `json:"timeEntriesSkipped"`
	TimeEntryConflicts int    `json:"timeEntryConflicts"`
	NeedsConfirmation  int    `json:"needsConfirmation"`
}

type ImportConflict struct {
	TimeEntryUUID string `json:"timeEntryUuid"`
	ProjectName   string `json:"projectName"`
	TaskName      string `json:"taskName"`
	Start         string `json:"start"`
	End           string `json:"end"`
	LocalEntryID  int64  `json:"localEntryId"`
	LocalStart    int64  `json:"localStart"`
	LocalEnd      int64  `json:"localEnd"`
}

func (a *App) GetDatabaseInfo() (*DatabaseInfo, error) {
	return a.databaseInfo()
}

func (a *App) SelectImportFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Time & Bill export",
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON files (*.json)",
			Pattern:     "*.json",
		}},
	})
}

func (a *App) SelectDatabaseFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select HumbleBee database",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite database (*.db;*.sqlite;*.sqlite3)", Pattern: "*.db;*.sqlite;*.sqlite3"},
			{DisplayName: "All files (*.*)", Pattern: "*.*"},
		},
	})
}

func (a *App) SelectNewDatabaseFile() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Create HumbleBee database",
		DefaultFilename:      "humblebee.db",
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{{
			DisplayName: "SQLite database (*.db)",
			Pattern:     "*.db",
		}},
	})
}

func (a *App) SwitchDatabase(path string) (*DatabaseInfo, error) {
	path, err := normalizeRequiredPath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("database path must be a file")
	}
	if err := validateDatabasePath(path); err != nil {
		return nil, err
	}
	if err := a.setSelectedDatabasePath(path); err != nil {
		return nil, err
	}
	return a.databaseInfo()
}

func (a *App) CreateDatabase(path string) (*DatabaseInfo, error) {
	path, err := normalizeRequiredPath(path)
	if err != nil {
		return nil, err
	}
	database, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	_ = database.Close()
	if err := a.setSelectedDatabasePath(path); err != nil {
		return nil, err
	}
	return a.databaseInfo()
}

func (a *App) UseDefaultDatabase() (*DatabaseInfo, error) {
	if err := a.clearSelectedDatabasePath(); err != nil {
		return nil, err
	}
	return a.databaseInfo()
}

func (a *App) PreviewTimeAndBillImport(path string) (*ImportPreview, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	existingCount, err := countExistingTimeEntries(database, personID)
	if err != nil {
		return nil, err
	}
	importer := service.NewTimeAndBillImportService(database)
	preview, err := importer.PreviewFile(personID, path, service.TimeAndBillImportOptions{SkipConflicting: true})
	if err != nil {
		return nil, err
	}
	return &ImportPreview{
		ExportUUID:             preview.ExportUUID,
		ExportedAt:             preview.ExportedAt,
		SourceUserEmail:        preview.SourceUserEmail,
		ExistingTimeEntryCount: existingCount,
		Summary:                importSummaryDTO(preview.Summary),
		Conflicts:              importConflictDTOs(preview.Conflicts),
	}, nil
}

func (a *App) ImportTimeAndBill(path string) (*ImportResult, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	importer := service.NewTimeAndBillImportService(database)
	preview, err := importer.PreviewFile(personID, path, service.TimeAndBillImportOptions{SkipConflicting: true})
	if err != nil {
		return nil, err
	}
	summary, err := importer.ImportFile(personID, path, service.TimeAndBillImportOptions{
		AssumeYes:       true,
		SkipConflicting: true,
	})
	if err != nil {
		return nil, err
	}
	return &ImportResult{
		Summary:   importSummaryDTO(summary),
		Conflicts: importConflictDTOs(preview.Conflicts),
	}, nil
}

func (a *App) databaseInfo() (*DatabaseInfo, error) {
	path, err := a.databasePath()
	if err != nil {
		return nil, err
	}
	defaultPath, err := a.defaultDatabasePath()
	if err != nil {
		return nil, err
	}
	database, err := db.Open(path)
	if err != nil {
		return nil, db.WrapBusyError(path, err)
	}
	defer database.Close()
	initialized, err := db.IsInitialized(database)
	if err != nil {
		return nil, db.WrapBusyError(path, err)
	}
	if initialized {
		if err := db.Migrate(database); err != nil {
			return nil, db.WrapBusyError(path, err)
		}
	}
	return &DatabaseInfo{Path: path, DefaultPath: defaultPath, Initialized: initialized}, nil
}

func normalizeRequiredPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("database path is required")
	}
	return filepath.Abs(path)
}

func validateDatabasePath(path string) error {
	database, err := db.Open(path)
	if err != nil {
		return err
	}
	defer database.Close()
	initialized, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if initialized {
		return db.Migrate(database)
	}
	return nil
}

func countExistingTimeEntries(database *sql.DB, personID int64) (int, error) {
	var count int
	err := database.QueryRow(`
		SELECT count(*)
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
	`, personID).Scan(&count)
	return count, err
}

func importSummaryDTO(summary service.TimeAndBillImportSummary) ImportSummary {
	return ImportSummary{
		ExportUUID:         summary.ExportUUID,
		AlreadyImported:    summary.AlreadyImported,
		ProjectsCreated:    summary.ProjectsCreated,
		ProjectsMapped:     summary.ProjectsMapped,
		ProjectsSkipped:    summary.ProjectsSkipped,
		TasksCreated:       summary.TasksCreated,
		TasksMapped:        summary.TasksMapped,
		TasksSkipped:       summary.TasksSkipped,
		TimeEntriesCreated: summary.TimeEntriesCreated,
		TimeEntriesUpdated: summary.TimeEntriesUpdated,
		TimeEntriesSkipped: summary.TimeEntriesSkipped,
		TimeEntryConflicts: summary.TimeEntryConflicts,
		NeedsConfirmation:  summary.NeedsConfirmation,
	}
}

func importConflictDTOs(conflicts []service.TimeAndBillImportConflict) []ImportConflict {
	out := make([]ImportConflict, 0, len(conflicts))
	for _, conflict := range conflicts {
		out = append(out, ImportConflict{
			TimeEntryUUID: conflict.TimeEntryUUID,
			ProjectName:   conflict.ProjectName,
			TaskName:      conflict.TaskName,
			Start:         conflict.Start,
			End:           conflict.End,
			LocalEntryID:  conflict.LocalEntryID,
			LocalStart:    conflict.LocalStart,
			LocalEnd:      conflict.LocalEnd,
		})
	}
	return out
}
