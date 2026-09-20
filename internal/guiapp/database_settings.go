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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/grobmeier/humblebee/internal/paths"
)

const guiSettingsFileName = "gui-settings.json"

var guiSettingsMu sync.Mutex

type guiSettings struct {
	SelectedDatabasePath string          `json:"selectedDatabasePath"`
	WorkspacePreferences json.RawMessage `json:"workspacePreferences,omitempty"`
}

func (a *App) databasePath() (string, error) {
	settings, err := readGUISettings()
	if err != nil {
		return "", err
	}
	if settings.SelectedDatabasePath != "" {
		if err := selectedDatabasePathExists(settings.SelectedDatabasePath); err != nil {
			return "", err
		}
		return settings.SelectedDatabasePath, nil
	}
	return a.defaultDatabasePath()
}

func selectedDatabasePathExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("selected database does not exist: %s", path)
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("selected database path is not a file: %s", path)
	}
	return nil
}

func (a *App) defaultDatabasePath() (string, error) {
	return paths.DBPath()
}

func (a *App) setSelectedDatabasePath(path string) error {
	guiSettingsMu.Lock()
	defer guiSettingsMu.Unlock()
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	settings, err := readGUISettings()
	if err != nil {
		return err
	}
	settings.SelectedDatabasePath = absolute
	return writeGUISettings(settings)
}

func (a *App) clearSelectedDatabasePath() error {
	guiSettingsMu.Lock()
	defer guiSettingsMu.Unlock()
	settings, err := readGUISettings()
	if err != nil {
		return err
	}
	settings.SelectedDatabasePath = ""
	return writeGUISettings(settings)
}

func readGUISettings() (guiSettings, error) {
	path, err := guiSettingsPath()
	if err != nil {
		return guiSettings{}, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return guiSettings{}, nil
		}
		return guiSettings{}, err
	}
	var settings guiSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return guiSettings{}, err
	}
	return settings, nil
}

func writeGUISettings(settings guiSettings) error {
	path, err := guiSettingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".gui-settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(append(body, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func guiSettingsPath() (string, error) {
	dir, err := paths.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, guiSettingsFileName), nil
}
