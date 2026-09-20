package guiapp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// Preferences are versioned, typed, and bounded; never store booking content here.
type ReportPreferences struct {
	Report      string `json:"report"`
	Mode        string `json:"mode"`
	Month       int    `json:"month"`
	StartMonth  int    `json:"startMonth"`
	EndMonth    int    `json:"endMonth"`
	Year        int    `json:"year"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	ProjectID   int64  `json:"projectId"`
	ShowDecimal bool   `json:"showDecimal"`
}

type ManualSelectionPreferences struct {
	ProjectID int64 `json:"projectId"`
	TaskID    int64 `json:"taskId"`
}

type WorkspacePreferences struct {
	WorkspaceKey    string                      `json:"workspaceKey"`
	Reports         *ReportPreferences          `json:"reports"`
	ManualSelection *ManualSelectionPreferences `json:"manualSelection"`
}

type storedPreferences struct {
	Version         int             `json:"version"`
	Reports         json.RawMessage `json:"reports,omitempty"`
	ManualSelection json.RawMessage `json:"manualSelection,omitempty"`
}

func preferenceWorkspaceKey(path string, personID int64) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d", resolved, personID)))), nil
}

func (a *App) preferenceContext() (string, error) {
	database, path, err := a.openDB()
	if err != nil {
		return "", err
	}
	defer database.Close()
	id, err := a.defaultPersonID(database)
	if err != nil {
		return "", err
	}
	return preferenceWorkspaceKey(path, id)
}

func readPreferenceMap(settings guiSettings) map[string]storedPreferences {
	values := map[string]storedPreferences{}
	// A damaged preference section must not invalidate database selection.
	if len(settings.WorkspacePreferences) > 1024*1024 || json.Unmarshal(settings.WorkspacePreferences, &values) != nil || values == nil {
		return map[string]storedPreferences{}
	}
	return values
}

func (a *App) GetWorkspacePreferences() (*WorkspacePreferences, error) {
	guiSettingsMu.Lock()
	defer guiSettingsMu.Unlock()
	key, err := a.preferenceContext()
	if err != nil {
		return nil, err
	}
	settings, err := readGUISettings()
	if err != nil {
		return nil, err
	}
	result := &WorkspacePreferences{WorkspaceKey: key}
	stored := readPreferenceMap(settings)[key]
	if stored.Version != 1 {
		return result, nil
	}
	var reports ReportPreferences
	if json.Unmarshal(stored.Reports, &reports) == nil && validateReportPreferences(reports) == nil {
		items, err := a.ListProjectWorkItems()
		if err != nil {
			return nil, err
		}
		found := false
		for _, item := range items {
			if item.ID == reports.ProjectID && item.ParentID == nil {
				found = true
			}
		}
		if !found {
			reports.ProjectID = 0
		}
		result.Reports = &reports
	}
	var manual ManualSelectionPreferences
	if json.Unmarshal(stored.ManualSelection, &manual) == nil && manual.ProjectID > 0 && manual.TaskID >= 0 {
		items, err := a.ListWorkItems()
		if err != nil {
			return nil, err
		}
		projectFound, taskFound := false, false
		for _, item := range items {
			if item.ID == manual.ProjectID && item.ParentID == nil {
				projectFound = true
			}
			if item.ID == manual.TaskID && item.ParentID != nil && *item.ParentID == manual.ProjectID {
				taskFound = true
			}
		}
		if projectFound {
			if !taskFound {
				manual.TaskID = 0
			}
			result.ManualSelection = &manual
		}
	}
	return result, nil
}

func validateReportPreferences(p ReportPreferences) error {
	switch p.Report {
	case "worktime-by-month", "worktime-grouped-by-project", "worktime-project-details", "worktime-task-details", "timesheet":
	default:
		return fmt.Errorf("invalid report")
	}
	if (p.Mode != "monthly" && p.Mode != "daily") || p.Year < 1 || p.Year > 9999 || p.Month < 1 || p.Month > 12 || p.StartMonth < 1 || p.EndMonth > 12 || p.StartMonth > p.EndMonth || p.ProjectID < 0 {
		return fmt.Errorf("invalid report filter")
	}
	start, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return fmt.Errorf("invalid report start date")
	}
	end, err := time.Parse("2006-01-02", p.EndDate)
	if err != nil || end.Before(start) {
		return fmt.Errorf("invalid report end date")
	}
	return nil
}

func (a *App) SaveReportPreferences(expectedWorkspaceKey string, preferences ReportPreferences) error {
	if err := validateReportPreferences(preferences); err != nil {
		return err
	}
	return a.savePreferences(expectedWorkspaceKey, preferences, true)
}

func (a *App) SaveManualSelectionPreferences(expectedWorkspaceKey string, preferences ManualSelectionPreferences) error {
	if preferences.ProjectID <= 0 || preferences.TaskID < 0 {
		return fmt.Errorf("invalid manual selection")
	}
	return a.savePreferences(expectedWorkspaceKey, preferences, false)
}

func (a *App) savePreferences(expected string, value any, reports bool) error {
	guiSettingsMu.Lock()
	defer guiSettingsMu.Unlock()
	key, err := a.preferenceContext()
	if err != nil {
		return err
	}
	if expected == "" || key != expected {
		return fmt.Errorf("workspace changed; preferences were not saved")
	}
	settings, err := readGUISettings()
	if err != nil {
		return err
	}
	values := readPreferenceMap(settings)
	stored, exists := values[key]
	if !exists && len(values) >= 256 {
		return fmt.Errorf("workspace preference limit reached")
	}
	if stored.Version != 1 {
		stored = storedPreferences{Version: 1}
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if reports {
		stored.Reports = body
	} else {
		stored.ManualSelection = body
	}
	values[key] = stored
	settings.WorkspacePreferences, err = json.Marshal(values)
	if err != nil {
		return err
	}
	return writeGUISettings(settings)
}
