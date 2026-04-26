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
		UUID:      uuid.NewString(),
		PersonID:  personID,
		WorkItemID: nil,
		StartTime: start,
		CreatedAt: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := entries.Stop(id, end, end-start); err != nil {
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

