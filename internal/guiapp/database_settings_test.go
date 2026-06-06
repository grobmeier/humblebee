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
	"path/filepath"
	"testing"
)

func TestDatabasePathUsesDefaultWhenNoGUISettingExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HUMBLEBEE_HOME", home)

	app := New()
	path, err := app.databasePath()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(home, "humblebee.db") {
		t.Fatalf("expected default database path, got %q", path)
	}
}

func TestDatabasePathPersistsSelectedPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HUMBLEBEE_HOME", home)

	app := New()
	selected := filepath.Join(t.TempDir(), "other.db")
	if err := app.setSelectedDatabasePath(selected); err != nil {
		t.Fatal(err)
	}

	path, err := app.databasePath()
	if err != nil {
		t.Fatal(err)
	}
	if path != selected {
		t.Fatalf("expected selected database path, got %q", path)
	}

	reloaded := New()
	reloadedPath, err := reloaded.databasePath()
	if err != nil {
		t.Fatal(err)
	}
	if reloadedPath != selected {
		t.Fatalf("expected selected database path after reload, got %q", reloadedPath)
	}
}

func TestClearSelectedDatabasePathReturnsToDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HUMBLEBEE_HOME", home)

	app := New()
	if err := app.setSelectedDatabasePath(filepath.Join(t.TempDir(), "other.db")); err != nil {
		t.Fatal(err)
	}
	if err := app.clearSelectedDatabasePath(); err != nil {
		t.Fatal(err)
	}
	path, err := app.databasePath()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(home, "humblebee.db") {
		t.Fatalf("expected default database path after clear, got %q", path)
	}
}
