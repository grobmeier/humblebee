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
	"fmt"
	"sort"
	"strings"

	"github.com/grobmeier/humblebee/internal/model"
)

type WorkItemRepo struct {
	db *sql.DB
}

func NewWorkItemRepo(db *sql.DB) *WorkItemRepo {
	return &WorkItemRepo{db: db}
}

func (r *WorkItemRepo) GetByID(personID, id int64) (*model.WorkItem, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
		FROM workitems
		WHERE person_id = ? AND id = ?
		LIMIT 1`, personID, id)
	return scanWorkItem(row)
}

func (r *WorkItemRepo) ListActive(personID int64) ([]model.WorkItem, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
		FROM workitems
		WHERE person_id = ? AND status = 'ACTIVE'
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.WorkItem
	for rows.Next() {
		item, err := scanWorkItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (r *WorkItemRepo) ListAll(personID int64) ([]model.WorkItem, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
		FROM workitems
		WHERE person_id = ?
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.WorkItem
	for rows.Next() {
		item, err := scanWorkItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *WorkItemRepo) ListProjectItems(personID int64) ([]model.WorkItem, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
		FROM workitems
		WHERE person_id = ? AND status <> 'DELETED'
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.WorkItem
	for rows.Next() {
		item, err := scanWorkItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (r *WorkItemRepo) FindByNameUnderParent(personID int64, parentID *int64, name string) (*model.WorkItem, error) {
	var row *sql.Row
	if parentID == nil {
		row = r.db.QueryRow(`
			SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
			FROM workitems
			WHERE person_id = ? AND parent_id IS NULL AND name = ?
			LIMIT 1`, personID, name)
	} else {
		row = r.db.QueryRow(`
			SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
			FROM workitems
			WHERE person_id = ? AND parent_id = ? AND name = ?
			LIMIT 1`, personID, *parentID, name)
	}
	item, err := scanWorkItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *WorkItemRepo) FindByNameAnyLevel(personID int64, name string) ([]model.WorkItem, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, name, description, parent_id, path, depth, status, color, created_at, updated_at
		FROM workitems
		WHERE person_id = ? AND name = ? AND status = 'ACTIVE'
		ORDER BY depth ASC, name ASC
	`, personID, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.WorkItem
	for rows.Next() {
		item, err := scanWorkItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type CreateWorkItemParams struct {
	PersonID int64
	UUID     string
	Name     string
	ParentID *int64
	Depth    int
	Created  int64
}

func (r *WorkItemRepo) Create(params CreateWorkItemParams) (*model.WorkItem, error) {
	res, err := r.db.Exec(`
		INSERT INTO workitems (uuid, person_id, name, parent_id, depth, status, created_at)
		VALUES (?, ?, ?, ?, ?, 'ACTIVE', ?)
	`, params.UUID, params.PersonID, params.Name, params.ParentID, params.Depth, params.Created)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	var path string
	if params.ParentID == nil {
		path = fmt.Sprintf("%d", id)
	} else {
		var parentPath sql.NullString
		if err := r.db.QueryRow(`SELECT path FROM workitems WHERE person_id = ? AND id = ?`, params.PersonID, *params.ParentID).Scan(&parentPath); err != nil {
			return nil, err
		}
		if !parentPath.Valid || parentPath.String == "" {
			return nil, errors.New("parent has no path")
		}
		path = parentPath.String + "/" + fmt.Sprintf("%d", id)
	}

	if _, err := r.db.Exec(`UPDATE workitems SET path = ? WHERE id = ?`, path, id); err != nil {
		return nil, err
	}
	return r.GetByID(params.PersonID, id)
}

func (r *WorkItemRepo) UpdateName(personID, workItemID int64, name string) (*model.WorkItem, error) {
	res, err := r.db.Exec(`
		UPDATE workitems
		SET name = ?, updated_at = strftime('%s','now')
		WHERE person_id = ? AND id = ? AND status = 'ACTIVE'
	`, name, personID, workItemID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New("work item not found")
	}
	return r.GetByID(personID, workItemID)
}

func (r *WorkItemRepo) SetStatus(personID, workItemID int64, status model.WorkItemStatus) (*model.WorkItem, error) {
	res, err := r.db.Exec(`
		UPDATE workitems
		SET status = ?, updated_at = strftime('%s','now')
		WHERE person_id = ? AND id = ?
	`, string(status), personID, workItemID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New("work item not found")
	}
	return r.GetByID(personID, workItemID)
}

func (r *WorkItemRepo) ArchiveSubtree(personID, workItemID int64) error {
	var path sql.NullString
	var name string
	var parentID sql.NullInt64
	if err := r.db.QueryRow(`
		SELECT path, name, parent_id
		FROM workitems
		WHERE person_id = ? AND id = ?
	`, personID, workItemID).Scan(&path, &name, &parentID); err != nil {
		return err
	}
	if strings.EqualFold(name, "Default") && !parentID.Valid {
		return errors.New("cannot remove the 'Default' work item")
	}
	if !path.Valid || path.String == "" {
		return errors.New("work item missing path")
	}
	p := path.String
	like := p + "/%"
	_, err := r.db.Exec(`
		UPDATE workitems
		SET status = 'ARCHIVED', updated_at = strftime('%s','now')
		WHERE person_id = ?
		  AND (path = ? OR path LIKE ?)
	`, personID, p, like)
	return err
}

func (r *WorkItemRepo) DeleteProjectAndTimeEntries(personID, projectID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var path sql.NullString
	var name string
	var parentID sql.NullInt64
	if err := tx.QueryRow(`
		SELECT path, name, parent_id
		FROM workitems
		WHERE person_id = ? AND id = ? AND status <> 'DELETED'
	`, personID, projectID).Scan(&path, &name, &parentID); err != nil {
		return err
	}
	if parentID.Valid {
		return errors.New("work item is not a project")
	}
	if strings.EqualFold(name, "Default") {
		return errors.New("cannot remove the 'Default' work item")
	}
	if !path.Valid || path.String == "" {
		return errors.New("work item missing path")
	}

	p := path.String
	like := p + "/%"
	if _, err := tx.Exec(`
		DELETE FROM closed_stopwatch_workitems
		WHERE person_id = ?
		  AND workitem_id IN (
			SELECT id FROM workitems
			WHERE person_id = ? AND (path = ? OR path LIKE ?)
		  )
	`, personID, personID, p, like); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM external_mappings
		WHERE local_table = 'time_entries'
		  AND local_id IN (
			SELECT id FROM time_entries
			WHERE person_id = ?
			  AND workitem_id IN (
				SELECT id FROM workitems
				WHERE person_id = ? AND (path = ? OR path LIKE ?)
			  )
		  )
	`, personID, personID, p, like); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM time_entries
		WHERE person_id = ?
		  AND workitem_id IN (
			SELECT id FROM workitems
			WHERE person_id = ? AND (path = ? OR path LIKE ?)
		  )
	`, personID, personID, p, like); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM external_mappings
		WHERE local_table = 'workitems'
		  AND local_id IN (
			SELECT id FROM workitems
			WHERE person_id = ? AND (path = ? OR path LIKE ?)
		  )
	`, personID, p, like); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM workitems
		WHERE person_id = ? AND (path = ? OR path LIKE ?)
	`, personID, p, like); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWorkItem(s scanner) (*model.WorkItem, error) {
	var item model.WorkItem
	var desc sql.NullString
	var parent sql.NullInt64
	var path sql.NullString
	var status string
	var color sql.NullString
	var updated sql.NullInt64
	if err := s.Scan(
		&item.ID,
		&item.UUID,
		&item.PersonID,
		&item.Name,
		&desc,
		&parent,
		&path,
		&item.Depth,
		&status,
		&color,
		&item.CreatedAt,
		&updated,
	); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		item.Description = &v
	}
	if parent.Valid {
		v := parent.Int64
		item.ParentID = &v
	}
	if path.Valid {
		v := path.String
		item.Path = &v
	}
	item.Status = model.WorkItemStatus(status)
	if color.Valid {
		v := color.String
		item.Color = &v
	}
	if updated.Valid {
		v := updated.Int64
		item.UpdatedAt = &v
	}
	return &item, nil
}
