package db

import (
	"database/sql"
	"errors"
)

const schemaVersion = 1

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
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE,
			FOREIGN KEY (workitem_id) REFERENCES workitems(id) ON DELETE SET NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_person ON time_entries(person_id);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_workitem ON time_entries(workitem_id);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_start ON time_entries(start_time);`,
		`CREATE INDEX IF NOT EXISTS idx_time_entries_running ON time_entries(end_time) WHERE end_time IS NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_time_entries_one_running_per_person
			ON time_entries(person_id)
			WHERE end_time IS NULL;`,

		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		);`,
		`INSERT OR IGNORE INTO config (key, value, created_at) VALUES ('schema_version', '1', strftime('%s','now'));`,
		`INSERT OR IGNORE INTO config (key, value, created_at) VALUES ('initialized_at', strftime('%s','now'), strftime('%s','now'));`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	// Basic schema_version check; future migrations can build on this.
	_, err = tx.Exec(`UPDATE config SET value = ? WHERE key='schema_version'`, schemaVersion)
	if err != nil {
		return err
	}

	return tx.Commit()
}
