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

