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
		return nil, "", db.WrapBusyError(path, err)
	}
	initialized, err := db.IsInitialized(database)
	if err != nil {
		_ = database.Close()
		return nil, "", db.WrapBusyError(path, err)
	}
	if initialized {
		// Apply idempotent migrations on every run for already-initialized databases.
		if err := db.Migrate(database); err != nil {
			_ = database.Close()
			return nil, "", db.WrapBusyError(path, err)
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
