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
	"errors"

	"github.com/grobmeier/humblebee/internal/model"
)

type PersonRepo struct {
	db *sql.DB
}

func NewPersonRepo(db *sql.DB) *PersonRepo {
	return &PersonRepo{db: db}
}

func (r *PersonRepo) GetDefault() (*model.Person, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, email, COALESCE(username,''), created_at, updated_at, is_active, is_default
		FROM persons
		WHERE is_default = 1 AND is_active = 1
		LIMIT 1`)
	var p model.Person
	var username string
	var updated sql.NullInt64
	var isActive int
	var isDefault int
	if err := row.Scan(&p.ID, &p.UUID, &p.Email, &username, &p.CreatedAt, &updated, &isActive, &isDefault); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.Username = username
	if updated.Valid {
		v := updated.Int64
		p.UpdatedAt = &v
	}
	p.IsActive = isActive == 1
	p.IsDefault = isDefault == 1
	return &p, nil
}

func (r *PersonRepo) CreateDefault(p model.Person) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE persons SET is_default = 0 WHERE is_default = 1`); err != nil {
		return 0, err
	}
	res, err := tx.Exec(`
		INSERT INTO persons (uuid, email, username, created_at, is_active, is_default)
		VALUES (?, ?, ?, ?, 1, 1)
	`, p.UUID, p.Email, p.Username, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

