package cli

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/paths"
	"github.com/grobmeier/humblebee/internal/repo"
)

func openDB() (*sql.DB, string, error) {
	path, err := paths.DBPath()
	if err != nil {
		return nil, "", err
	}
	database, err := db.Open(path)
	if err != nil {
		return nil, "", err
	}
	initialized, err := db.IsInitialized(database)
	if err != nil {
		_ = database.Close()
		return nil, "", err
	}
	if initialized {
		// Apply idempotent migrations on every run for already-initialized databases.
		if err := db.Migrate(database); err != nil {
			_ = database.Close()
			return nil, "", err
		}
	}
	return database, path, nil
}

func requireInitialized(database *sql.DB) error {
	ok, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("HumbleBee is not initialized. Run 'humblebee init' first.")
	}
	return nil
}

func defaultPersonID(database *sql.DB) (int64, error) {
	people := repo.NewPersonRepo(database)
	p, err := people.GetDefault()
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, fmt.Errorf("no default user found; run 'humblebee init' again")
	}
	return p.ID, nil
}
