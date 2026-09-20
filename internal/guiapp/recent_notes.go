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
	"errors"
)

// GetRecentNotes returns up to five exact, nonblank booked notes for the current
// database/profile/task, rejecting requests queued for another database.
// Task zero means Default (a NULL workitem_id). Callers must discard results
// if their UI context changes while awaiting the response.
func (a *App) GetRecentNotes(expectedDatabasePath string, taskID int64) ([]string, error) {
	databaseSelectionMu.RLock()
	defer databaseSelectionMu.RUnlock()
	if taskID < 0 {
		return nil, errors.New("invalid task id")
	}
	if _, err := a.expectedDatabasePath(expectedDatabasePath); err != nil {
		return nil, err
	}
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
	// SQLite's default trim removes only ASCII spaces. Match Go's Unicode
	// whitespace set for blank filtering without altering the returned note.
	const whitespace = "\t\n\v\f\r \u0085\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000"
	rows, err := database.Query(`
		SELECT description FROM (
			SELECT description, created_at, id,
			       ROW_NUMBER() OVER (PARTITION BY description ORDER BY created_at DESC, id DESC) AS position
			FROM time_entries
			WHERE person_id = ? AND (workitem_id = ? OR (? = 0 AND workitem_id IS NULL))
			  AND end_time IS NOT NULL
			  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
			  AND description IS NOT NULL AND trim(description, ?) != ''
		) WHERE position = 1 ORDER BY created_at DESC, id DESC LIMIT 5
	`, personID, taskID, taskID, whitespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := make([]string, 0, 5)
	for rows.Next() {
		var note string
		if err := rows.Scan(&note); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}
