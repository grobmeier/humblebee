package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openMemory(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Ping(); err != nil {
		t.Fatal(err)
	}
	return database
}

func TestMigrateAndIsInitialized(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	ok, err := IsInitialized(database)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected not initialized")
	}

	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}
	ok, err = IsInitialized(database)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected initialized")
	}
}

func TestMigrateAddsTimeEntryTZColumnsToExistingDB(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	// Simulate an older database that has time_entries but no tz_* columns.
	if _, err := database.Exec(`
		CREATE TABLE config (key TEXT PRIMARY KEY, value TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER);
		INSERT INTO config (key, value, created_at) VALUES ('schema_version', '1', strftime('%s','now'));

		CREATE TABLE time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			workitem_id INTEGER,
			description TEXT,
			start_time INTEGER NOT NULL,
			end_time INTEGER,
			duration INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		);
	`); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}

	rows, err := database.Query(`PRAGMA table_info(time_entries);`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if !cols["tz_name"] || !cols["tz_offset_minutes"] {
		t.Fatalf("expected tz columns to exist, got %#v", cols)
	}
}
