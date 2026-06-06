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
		INSERT INTO time_entries (uuid, person_id, start_time, end_time, duration, created_at)
			VALUES ('running-1', 1, 1000, NULL, NULL, 1000);
		INSERT INTO time_entries (uuid, person_id, start_time, end_time, duration, created_at)
			VALUES ('completed-1', 1, 1000, 1600, 600, 1000);
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
	if !cols["entry_source"] {
		t.Fatalf("expected entry_source column to exist, got %#v", cols)
	}

	var runningSource string
	if err := database.QueryRow(`SELECT entry_source FROM time_entries WHERE uuid = 'running-1'`).Scan(&runningSource); err != nil {
		t.Fatal(err)
	}
	if runningSource != "stopwatch" {
		t.Fatalf("expected legacy running row to be restored as stopwatch, got %q", runningSource)
	}
	var completedSource string
	if err := database.QueryRow(`SELECT entry_source FROM time_entries WHERE uuid = 'completed-1'`).Scan(&completedSource); err != nil {
		t.Fatal(err)
	}
	if completedSource != "manual" {
		t.Fatalf("expected completed legacy row to stay manual, got %q", completedSource)
	}
}

func TestMigrateBackfillsRunningStopwatchWhenEntrySourceAlreadyExists(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if _, err := database.Exec(`
		CREATE TABLE config (key TEXT PRIMARY KEY, value TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER);
		INSERT INTO config (key, value, created_at) VALUES ('schema_version', '5', strftime('%s','now'));

		CREATE TABLE time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			workitem_id INTEGER,
			description TEXT,
			start_time INTEGER NOT NULL,
			end_time INTEGER,
			duration INTEGER,
			entry_source TEXT NOT NULL DEFAULT 'manual',
			tz_name TEXT NOT NULL DEFAULT '',
			tz_offset_minutes INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		);
		INSERT INTO time_entries (uuid, person_id, start_time, end_time, duration, entry_source, created_at)
			VALUES ('running-1', 1, 1000, NULL, NULL, 'manual', 1000);
	`); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}

	var source string
	if err := database.QueryRow(`SELECT entry_source FROM time_entries WHERE uuid = 'running-1'`).Scan(&source); err != nil {
		t.Fatal(err)
	}
	if source != "stopwatch" {
		t.Fatalf("expected existing entry_source database to restore running stopwatch, got %q", source)
	}
}

func TestMigrateSkipsCurrentSchema(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE migration_probe (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := database.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='migration_probe'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected current schema migration to leave unrelated tables unchanged")
	}
}
