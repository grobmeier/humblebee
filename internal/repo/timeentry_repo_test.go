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

package repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/model"
)

func TestTimeEntryDeleteByID(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Now().UTC().Add(-2 * time.Hour).Unix()
	end := time.Now().UTC().Add(-1 * time.Hour).Unix()
	id, err := entries.Start(model.TimeEntry{
		UUID:       uuid.NewString(),
		PersonID:   personID,
		WorkItemID: nil,
		StartTime:  start,
		CreatedAt:  start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := entries.Stop(personID, id, end, end-start); err != nil {
		t.Fatal(err)
	}

	windowStart := time.Now().UTC().Add(-24 * time.Hour).Unix()
	windowEnd := time.Now().UTC().Add(24 * time.Hour).Unix()
	list, err := entries.ListOverlappingForWorkItem(personID, nil, windowStart, windowEnd)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}

	if err := entries.DeleteByID(personID, id); err != nil {
		t.Fatal(err)
	}
	list, err = entries.ListOverlappingForWorkItem(personID, nil, windowStart, windowEnd)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(list))
	}
}

func TestTimeEntryCreateCompletedAndOverlap(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC).Unix()
	end := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC).Unix()
	duration := end - start
	endPtr := end
	durationPtr := duration
	id, err := entries.CreateCompleted(model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  personID,
		StartTime: start,
		EndTime:   &endPtr,
		Duration:  &durationPtr,
		CreatedAt: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}

	overlaps, err := entries.HasOverlap(personID, start+60, end+60)
	if err != nil {
		t.Fatal(err)
	}
	if !overlaps {
		t.Fatal("expected overlapping interval to be detected")
	}

	overlaps, err = entries.HasOverlap(personID, end, end+3600)
	if err != nil {
		t.Fatal(err)
	}
	if overlaps {
		t.Fatal("expected adjacent interval not to overlap")
	}

	overlaps, err = entries.HasOverlapExcluding(personID, id, start+60, end-60)
	if err != nil {
		t.Fatal(err)
	}
	if overlaps {
		t.Fatal("expected edited entry not to overlap with itself")
	}
}

func TestTimeEntryRunningStopwatchDoesNotOverlapManualEntry(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC).Unix()
	if _, err := entries.Start(model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  personID,
		StartTime: start,
		CreatedAt: start,
	}); err != nil {
		t.Fatal(err)
	}

	overlaps, err := entries.HasOverlap(personID, start+60, start+3600)
	if err != nil {
		t.Fatal(err)
	}
	if overlaps {
		t.Fatal("expected running stopwatch not to block manual entry")
	}
}

func TestTimeEntryCloseStopwatchDeletesStoppedStopwatch(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC).Unix()
	end := start + 3600
	id, err := entries.Start(model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  personID,
		StartTime: start,
		CreatedAt: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := entries.Stop(personID, id, end, end-start); err != nil {
		t.Fatal(err)
	}
	if err := entries.CloseStopwatchByEntryID(personID, id); err != nil {
		t.Fatal(err)
	}

	entry, err := entries.GetByID(personID, id)
	if err != nil {
		t.Fatal(err)
	}
	if entry != nil {
		t.Fatal("expected discarded stopwatch entry to be deleted")
	}
	stopwatches, err := entries.ListStopwatches(personID, 12)
	if err != nil {
		t.Fatal(err)
	}
	if len(stopwatches) != 0 {
		t.Fatalf("expected discarded stopwatch not to be listed, got %d", len(stopwatches))
	}
}

func TestTimeEntryConflictingStopwatchDoesNotOverlapManualEntry(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC).Unix()
	id, err := entries.Start(model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  personID,
		StartTime: start,
		CreatedAt: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := entries.MarkStopwatchConflict(personID, id, start+3600, 3600); err != nil {
		t.Fatal(err)
	}

	overlaps, err := entries.HasOverlap(personID, start+60, start+1800)
	if err != nil {
		t.Fatal(err)
	}
	if overlaps {
		t.Fatal("expected conflicting stopwatch not to block manual entry")
	}
}

func TestTimeEntryStopwatchStateUpdatesAreScopedToPerson(t *testing.T) {
	database := openMemory(t)
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := NewPersonRepo(database)
	personID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "user@example.com",
		Username:  "user",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	otherPersonID, err := personRepo.CreateDefault(model.Person{
		UUID:      uuid.NewString(),
		Email:     "other@example.com",
		Username:  "other",
		CreatedAt: time.Now().UTC().Unix(),
		IsActive:  true,
		IsDefault: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := NewTimeEntryRepo(database)
	start := time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC).Unix()
	id, err := entries.Start(model.TimeEntry{
		UUID:      uuid.NewString(),
		PersonID:  personID,
		StartTime: start,
		CreatedAt: start,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := entries.MarkStopwatchConflict(otherPersonID, id, start+3600, 3600); err == nil {
		t.Fatal("expected conflict update with wrong person to fail")
	}
	entry, err := entries.GetByID(personID, id)
	if err != nil {
		t.Fatal(err)
	}
	if entry == nil || entry.EndTime != nil || entry.EntrySource != "stopwatch" {
		t.Fatalf("expected wrong-person conflict update not to change stopwatch, got %#v", entry)
	}

	if err := entries.MarkStopwatchUnbooked(otherPersonID, id); err == nil {
		t.Fatal("expected unbook update with wrong person to fail")
	}
	entry, err = entries.GetByID(personID, id)
	if err != nil {
		t.Fatal(err)
	}
	if entry == nil || entry.EntrySource != "stopwatch" {
		t.Fatalf("expected wrong-person unbook update not to change stopwatch, got %#v", entry)
	}
}
