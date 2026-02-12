package repo

import (
	"database/sql"
	"errors"

	"github.com/grobmeier/humblebee/internal/model"
)

type TimeEntryRepo struct {
	db *sql.DB
}

func NewTimeEntryRepo(db *sql.DB) *TimeEntryRepo {
	return &TimeEntryRepo{db: db}
}

func (r *TimeEntryRepo) FindRunning(personID int64) (*model.TimeEntry, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, created_at, updated_at
		FROM time_entries
		WHERE person_id = ? AND end_time IS NULL
		LIMIT 1
	`, personID)
	entry, err := scanTimeEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return entry, nil
}

func (r *TimeEntryRepo) Start(e model.TimeEntry) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO time_entries (uuid, person_id, workitem_id, description, start_time, end_time, duration, created_at)
		VALUES (?, ?, ?, ?, ?, NULL, NULL, ?)
	`, e.UUID, e.PersonID, e.WorkItemID, e.Description, e.StartTime, e.CreatedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TimeEntryRepo) Stop(entryID int64, endTime int64, duration int64) error {
	_, err := r.db.Exec(`
		UPDATE time_entries
		SET end_time = ?, duration = ?, updated_at = strftime('%s','now')
		WHERE id = ?
	`, endTime, duration, entryID)
	return err
}

func (r *TimeEntryRepo) ListOverlapping(personID int64, windowStart, windowEnd int64) ([]model.TimeEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, created_at, updated_at
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND start_time < ?
		  AND end_time > ?
	`, personID, windowEnd, windowStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.TimeEntry
	for rows.Next() {
		entry, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type timeEntryScanner interface {
	Scan(dest ...any) error
}

func scanTimeEntry(s timeEntryScanner) (*model.TimeEntry, error) {
	var e model.TimeEntry
	var workItemID sql.NullInt64
	var desc sql.NullString
	var end sql.NullInt64
	var dur sql.NullInt64
	var updated sql.NullInt64
	if err := s.Scan(
		&e.ID,
		&e.UUID,
		&e.PersonID,
		&workItemID,
		&desc,
		&e.StartTime,
		&end,
		&dur,
		&e.CreatedAt,
		&updated,
	); err != nil {
		return nil, err
	}
	if workItemID.Valid {
		v := workItemID.Int64
		e.WorkItemID = &v
	}
	if desc.Valid {
		v := desc.String
		e.Description = &v
	}
	if end.Valid {
		v := end.Int64
		e.EndTime = &v
	}
	if dur.Valid {
		v := dur.Int64
		e.Duration = &v
	}
	if updated.Valid {
		v := updated.Int64
		e.UpdatedAt = &v
	}
	return &e, nil
}

