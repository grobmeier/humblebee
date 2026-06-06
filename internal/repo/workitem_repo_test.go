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
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/model"

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

func TestWorkItemUniqueAtRootCaseInsensitive(t *testing.T) {
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

	workRepo := NewWorkItemRepo(database)
	_, err = workRepo.Create(CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     "Client",
		ParentID: nil,
		Depth:    0,
		Created:  time.Now().UTC().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = workRepo.Create(CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     "client",
		ParentID: nil,
		Depth:    0,
		Created:  time.Now().UTC().Unix(),
	})
	if err == nil {
		t.Fatalf("expected unique constraint error")
	}
}

func TestDeleteProjectAndTimeEntriesRejectsTask(t *testing.T) {
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

	workRepo := NewWorkItemRepo(database)
	project, err := workRepo.Create(CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     "Client",
		Depth:    0,
		Created:  time.Now().UTC().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err := workRepo.Create(CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     "Research",
		ParentID: &project.ID,
		Depth:    1,
		Created:  time.Now().UTC().Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = workRepo.DeleteProjectAndTimeEntries(personID, task.ID)
	if err == nil || err.Error() != "work item is not a project" {
		t.Fatalf("expected work item kind error, got %v", err)
	}
}
