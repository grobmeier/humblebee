package cli

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"

	_ "modernc.org/sqlite"
)

func TestBackfillTZ(t *testing.T) {
	database, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	personRepo := repo.NewPersonRepo(database)
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

	entries := repo.NewTimeEntryRepo(database)
	start := time.Now().UTC().Add(-2 * time.Hour).Unix()
	end := time.Now().UTC().Add(-1 * time.Hour).Unix()
	id, err := entries.Start(model.TimeEntry{
		UUID:       uuid.NewString(),
		PersonID:   personID,
		WorkItemID: nil,
		StartTime:  start,
		CreatedAt:  start,
		TZName:     "",
		TZOffsetMin: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := entries.Stop(id, end, end-start); err != nil {
		t.Fatal(err)
	}

	// Ensure it is counted as missing.
	missing, err := countMissingTZ(database, personID)
	if err != nil {
		t.Fatal(err)
	}
	if missing != 1 {
		t.Fatalf("expected missing=1, got %d", missing)
	}

	n, err := backfillTZ(database, personID, time.FixedZone("X", 2*3600))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected updated=1, got %d", n)
	}

	missing, err = countMissingTZ(database, personID)
	if err != nil {
		t.Fatal(err)
	}
	if missing != 0 {
		t.Fatalf("expected missing=0, got %d", missing)
	}
}

