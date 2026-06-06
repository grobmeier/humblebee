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
	"errors"
	"strconv"
)

const schemaVersion = 6

func IsInitialized(db *sql.DB) (bool, error) {
	var v string
	err := db.QueryRow(`SELECT value FROM config WHERE key='schema_version'`).Scan(&v)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	// If config table doesn't exist, treat as not initialized.
	var count int
	if err2 := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='config'`).Scan(&count); err2 != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return false, err
}

func Migrate(db *sql.DB) error {
	current, err := currentSchemaVersion(db)
	if err == nil && current >= schemaVersion {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS persons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			username TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			is_active INTEGER DEFAULT 1,
			is_default INTEGER DEFAULT 0
		);`,
		`CREATE INDEX IF NOT EXISTS idx_persons_email ON persons(email);`,
		`CREATE INDEX IF NOT EXISTS idx_persons_active ON persons(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_persons_default ON persons(is_default);`,

		`CREATE TABLE IF NOT EXISTS workitems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			person_id INTEGER NOT NULL,
			name TEXT NOT NULL COLLATE NOCASE,
			description TEXT,
			parent_id INTEGER,
			path TEXT,
			depth INTEGER DEFAULT 0,
			status TEXT DEFAULT 'ACTIVE',
			color TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES workitems(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_workitems_person ON workitems(person_id);`,
		`CREATE INDEX IF NOT EXISTS idx_workitems_parent ON workitems(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_workitems_status ON workitems(status);`,
		`CREATE INDEX IF NOT EXISTS idx_workitems_path ON workitems(path);`,
		`DROP INDEX IF EXISTS idx_workitems_person_name;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_workitems_person_name_parent
			ON workitems(person_id, name, ifnull(parent_id,0));`,

		`CREATE TABLE IF NOT EXISTS time_entries (
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
			updated_at INTEGER,
			FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE,
			FOREIGN KEY (workitem_id) REFERENCES workitems(id) ON DELETE SET NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_person ON time_entries(person_id);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_workitem ON time_entries(workitem_id);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_start ON time_entries(start_time);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_running ON time_entries(end_time) WHERE end_time IS NULL;`,
		`DROP INDEX IF EXISTS idx_time_entries_one_running_per_person;`,

		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS import_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			export_uuid TEXT UNIQUE NOT NULL,
			source_format TEXT NOT NULL,
			source_user_uuid TEXT,
			source_user_email TEXT,
			imported_at INTEGER NOT NULL,
			summary_json TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS external_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_system TEXT NOT NULL,
			source_uuid TEXT NOT NULL,
			local_table TEXT NOT NULL,
			local_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			UNIQUE (source_system, source_uuid, local_table)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_external_mappings_local
			ON external_mappings(local_table, local_id);`,
		`CREATE TABLE IF NOT EXISTS closed_stopwatch_workitems (
			person_id INTEGER NOT NULL,
			workitem_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY (person_id, workitem_id),
			FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE
		);`,
		`INSERT OR IGNORE INTO config (key, value, created_at) VALUES ('schema_version', '3', strftime('%s','now'));`,
		`INSERT OR IGNORE INTO config (key, value, created_at) VALUES ('initialized_at', strftime('%s','now'), strftime('%s','now'));`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	// Ensure new columns exist on older databases.
	if err := ensureTimeEntryTZColumns(tx); err != nil {
		return err
	}
	if err := ensureTimeEntrySourceColumn(tx); err != nil {
		return err
	}

	// Basic schema_version check; future migrations can build on this.
	_, err = tx.Exec(`UPDATE config SET value = ? WHERE key='schema_version'`, schemaVersion)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func currentSchemaVersion(db *sql.DB) (int, error) {
	var value string
	if err := db.QueryRow(`SELECT value FROM config WHERE key='schema_version'`).Scan(&value); err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func ensureTimeEntrySourceColumn(tx *sql.Tx) error {
	exists, err := timeEntryColumnExists(tx, "entry_source")
	if err != nil {
		return err
	}
	if exists {
		return backfillLegacyRunningStopwatches(tx)
	}
	if _, err := tx.Exec(`ALTER TABLE time_entries ADD COLUMN entry_source TEXT NOT NULL DEFAULT 'manual';`); err != nil {
		return err
	}
	return backfillLegacyRunningStopwatches(tx)
}

func backfillLegacyRunningStopwatches(tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE time_entries SET entry_source = 'stopwatch' WHERE end_time IS NULL AND entry_source = 'manual';`)
	return err
}

func ensureTimeEntryTZColumns(tx *sql.Tx) error {
	tzNameExists, err := timeEntryColumnExists(tx, "tz_name")
	if err != nil {
		return err
	}
	tzOffsetExists, err := timeEntryColumnExists(tx, "tz_offset_minutes")
	if err != nil {
		return err
	}
	if !tzNameExists {
		if _, err := tx.Exec(`ALTER TABLE time_entries ADD COLUMN tz_name TEXT NOT NULL DEFAULT '';`); err != nil {
			return err
		}
	}
	if !tzOffsetExists {
		if _, err := tx.Exec(`ALTER TABLE time_entries ADD COLUMN tz_offset_minutes INTEGER NOT NULL DEFAULT 0;`); err != nil {
			return err
		}
	}
	return nil
}

func timeEntryColumnExists(tx *sql.Tx, column string) (bool, error) {
	rows, err := tx.Query(`PRAGMA table_info(time_entries);`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}
