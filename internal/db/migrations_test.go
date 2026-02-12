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

