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
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func insertRecentNote(t *testing.T, database *sql.DB, person int64, task interface{}, note interface{}, created int64, source string, end interface{}) int64 {
	t.Helper()
	result, err := database.Exec(`INSERT INTO time_entries (uuid, person_id, workitem_id, description, start_time, end_time, created_at, entry_source) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), person, task, note, 1000-created, end, created, source)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestRecentNotesFilteringOrderingAndFidelity(t *testing.T) {
	a, database, source := backupFixture(t)
	person, err := a.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	project, err := a.CreateProject("Project")
	if err != nil {
		t.Fatal(err)
	}
	task, err := a.CreateTask(project.ID, "Task")
	if err != nil {
		t.Fatal(err)
	}
	otherTask, err := a.CreateTask(project.ID, "Other")
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(1); i <= 7; i++ {
		insertRecentNote(t, database, person, task.ID, fmt.Sprintf("note %d", i), i, "manual", 2000)
	}
	full := "  Gr\u00fc\u00dfe \u65e5\u672c\u8a9e\nsecond line\n  "
	insertRecentNote(t, database, person, task.ID, full, 8, "manual", 2000)
	insertRecentNote(t, database, person, task.ID, "note 3", 9, "stopwatch", 2000)
	insertRecentNote(t, database, person, task.ID, "tie first", 10, "manual", 2000)
	insertRecentNote(t, database, person, task.ID, "tie last", 10, "manual", 2000)
	for _, blank := range []interface{}{nil, "", " \t\r\n", "\u00a0\u2003\u3000"} {
		insertRecentNote(t, database, person, task.ID, blank, 100, "manual", 2000)
	}
	for _, source := range []string{"stopwatch_conflict", "stopwatch_unbooked"} {
		insertRecentNote(t, database, person, task.ID, source, 101, source, 2000)
	}
	insertRecentNote(t, database, person, task.ID, "running", 102, "stopwatch", nil)
	insertRecentNote(t, database, person, otherTask.ID, "other task", 103, "manual", 2000)
	insertRecentNote(t, database, person, nil, "default note", 104, "manual", 2000)
	if _, err := database.Exec(`INSERT INTO persons (id, uuid, email, created_at) VALUES (99, 'other', 'other@example.test', 0)`); err != nil {
		t.Fatal(err)
	}
	insertRecentNote(t, database, 99, task.ID, "other profile", 105, "manual", 2000)
	want := []string{"tie last", "tie first", "note 3", full, "note 7"}
	got, err := a.GetRecentNotes(source, task.ID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("notes: %#v %v, want %#v", got, err, want)
	}
	got, err = a.GetRecentNotes(source, 0)
	if err != nil || !reflect.DeepEqual(got, []string{"default note"}) {
		t.Fatalf("default: %#v %v", got, err)
	}
	got, err = a.GetRecentNotes(source, 9999)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty: %#v %v", got, err)
	}
}

func TestRecentNotesReflectEditsDeletesAndDatabaseContext(t *testing.T) {
	a, database, source := backupFixture(t)
	person, err := a.defaultPersonID(database)
	if err != nil {
		t.Fatal(err)
	}
	id := insertRecentNote(t, database, person, nil, "original", 1, "manual", 2000)
	got, err := a.GetRecentNotes(source, 0)
	if err != nil || !reflect.DeepEqual(got, []string{"original"}) {
		t.Fatalf("initial: %#v %v", got, err)
	}
	if _, err := database.Exec("UPDATE time_entries SET description = 'edited' WHERE id = ?", id); err != nil {
		t.Fatal(err)
	}
	got, err = a.GetRecentNotes(source, 0)
	if err != nil || !reflect.DeepEqual(got, []string{"edited"}) {
		t.Fatalf("edited: %#v %v", got, err)
	}
	if _, err := database.Exec("DELETE FROM time_entries WHERE id = ?", id); err != nil {
		t.Fatal(err)
	}
	got, err = a.GetRecentNotes(source, 0)
	if err != nil || len(got) != 0 {
		t.Fatalf("deleted: %#v %v", got, err)
	}
	insertRecentNote(t, database, person, nil, "only first database", 2, "manual", 2000)
	if _, err := a.CreateDatabase(filepath.Join(t.TempDir(), "second.db")); err != nil {
		t.Fatal(err)
	}
	if err := a.Init("second@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetRecentNotes(source, 0); err == nil {
		t.Fatal("accepted stale context")
	}
	second, err := a.databasePath()
	if err != nil {
		t.Fatal(err)
	}
	got, err = a.GetRecentNotes(second, 0)
	if err != nil || len(got) != 0 {
		t.Fatalf("database isolation: %#v %v", got, err)
	}
	if _, err := a.GetRecentNotes("", 0); err == nil {
		t.Fatal("accepted missing expected path")
	}
	if _, err := a.GetRecentNotes(second, -1); err == nil {
		t.Fatal("accepted negative task")
	}
}
