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
