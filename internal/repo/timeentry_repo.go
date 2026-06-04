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
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
		FROM time_entries
		WHERE person_id = ? AND end_time IS NULL AND entry_source = 'stopwatch'
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

func (r *TimeEntryRepo) GetByID(personID, entryID int64) (*model.TimeEntry, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
		FROM time_entries
		WHERE person_id = ? AND id = ?
		LIMIT 1
	`, personID, entryID)
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
		INSERT INTO time_entries (uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at)
		VALUES (?, ?, ?, ?, ?, NULL, NULL, 'stopwatch', ?, ?, ?)
	`, e.UUID, e.PersonID, e.WorkItemID, e.Description, e.StartTime, e.TZName, e.TZOffsetMin, e.CreatedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TimeEntryRepo) ListStopwatches(personID int64, limit int) ([]model.TimeEntry, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.uuid, e.person_id, e.workitem_id, e.description, e.start_time, e.end_time, e.duration, e.entry_source, e.tz_name, e.tz_offset_minutes, e.created_at, e.updated_at
		FROM time_entries e
		WHERE e.person_id = ?
		  AND e.entry_source IN ('stopwatch', 'stopwatch_conflict', 'stopwatch_unbooked')
		  AND NOT EXISTS (
			SELECT 1
			FROM closed_stopwatch_workitems c
			WHERE c.person_id = e.person_id
			  AND c.workitem_id = coalesce(e.workitem_id, 0)
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM time_entries newer
			WHERE newer.person_id = e.person_id
			  AND newer.entry_source IN ('stopwatch', 'stopwatch_conflict', 'stopwatch_unbooked')
			  AND coalesce(newer.workitem_id, 0) = coalesce(e.workitem_id, 0)
			  AND (
				coalesce(newer.end_time, newer.start_time) > coalesce(e.end_time, e.start_time)
				OR (
					coalesce(newer.end_time, newer.start_time) = coalesce(e.end_time, e.start_time)
					AND newer.id > e.id
				)
			  )
		  )
		ORDER BY e.end_time IS NULL DESC, coalesce(e.end_time, e.start_time) DESC
		LIMIT ?
	`, personID, limit)
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

func (r *TimeEntryRepo) ReopenStopwatchWorkItem(personID int64, workItemID *int64) error {
	_, err := r.db.Exec(`
		DELETE FROM closed_stopwatch_workitems
		WHERE person_id = ? AND workitem_id = ?
	`, personID, stopwatchWorkItemKey(workItemID))
	return err
}

func (r *TimeEntryRepo) CloseStopwatchByEntryID(personID, entryID int64) error {
	entry, err := r.GetByID(personID, entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return errors.New("stopwatch not found")
	}
	if err := r.DeleteByID(personID, entry.ID); err != nil {
		return err
	}
	return r.CloseStopwatchWorkItem(personID, entry.WorkItemID)
}

func (r *TimeEntryRepo) CloseStopwatchWorkItem(personID int64, workItemID *int64) error {
	_, err := r.db.Exec(`
		INSERT INTO closed_stopwatch_workitems (person_id, workitem_id, created_at)
		VALUES (?, ?, strftime('%s','now'))
		ON CONFLICT(person_id, workitem_id) DO UPDATE SET created_at = excluded.created_at
	`, personID, stopwatchWorkItemKey(workItemID))
	return err
}

func (r *TimeEntryRepo) CreateCompleted(e model.TimeEntry) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO time_entries (uuid, person_id, workitem_id, description, start_time, end_time, duration, tz_name, tz_offset_minutes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.UUID, e.PersonID, e.WorkItemID, e.Description, e.StartTime, e.EndTime, e.Duration, e.TZName, e.TZOffsetMin, e.CreatedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TimeEntryRepo) UpdateCompleted(e model.TimeEntry) error {
	res, err := r.db.Exec(`
		UPDATE time_entries
		SET workitem_id = ?, description = ?, start_time = ?, end_time = ?, duration = ?, tz_name = ?, tz_offset_minutes = ?, updated_at = strftime('%s','now')
		WHERE id = ?
		  AND person_id = ?
		  AND end_time IS NOT NULL
	`, e.WorkItemID, e.Description, e.StartTime, e.EndTime, e.Duration, e.TZName, e.TZOffsetMin, e.ID, e.PersonID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("time entry not found")
	}
	return nil
}

func (r *TimeEntryRepo) Stop(entryID int64, endTime int64, duration int64) error {
	_, err := r.db.Exec(`
		UPDATE time_entries
		SET end_time = ?, duration = ?, updated_at = strftime('%s','now')
		WHERE id = ?
	`, endTime, duration, entryID)
	return err
}

func (r *TimeEntryRepo) MarkStopwatchConflict(entryID int64, endTime int64, duration int64) error {
	_, err := r.db.Exec(`
		UPDATE time_entries
		SET end_time = ?, duration = ?, entry_source = 'stopwatch_conflict', updated_at = strftime('%s','now')
		WHERE id = ?
	`, endTime, duration, entryID)
	return err
}

func (r *TimeEntryRepo) MarkStopwatchUnbooked(entryID int64) error {
	_, err := r.db.Exec(`
		UPDATE time_entries
		SET entry_source = 'stopwatch_unbooked', updated_at = strftime('%s','now')
		WHERE id = ?
	`, entryID)
	return err
}

func (r *TimeEntryRepo) HasOverlap(personID int64, windowStart, windowEnd int64) (bool, error) {
	return r.HasOverlapExcluding(personID, 0, windowStart, windowEnd)
}

func (r *TimeEntryRepo) HasOverlapExcluding(personID int64, excludedEntryID int64, windowStart, windowEnd int64) (bool, error) {
	var count int
	if err := r.db.QueryRow(`
		SELECT count(*)
		FROM time_entries
		WHERE person_id = ?
		  AND id != ?
		  AND end_time IS NOT NULL
		  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
		  AND start_time < ?
		  AND end_time > ?
	`, personID, excludedEntryID, windowEnd, windowStart).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TimeEntryRepo) ListOverlapping(personID int64, windowStart, windowEnd int64) ([]model.TimeEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
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

func (r *TimeEntryRepo) ListOverlappingForWorkItem(personID int64, workItemID *int64, windowStart, windowEnd int64) ([]model.TimeEntry, error) {
	if workItemID == nil {
		rows, err := r.db.Query(`
			SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
			FROM time_entries
			WHERE person_id = ?
			  AND end_time IS NOT NULL
			  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
			  AND workitem_id IS NULL
			  AND start_time < ?
			  AND end_time > ?
			ORDER BY start_time ASC
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

	rows, err := r.db.Query(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
		  AND workitem_id = ?
		  AND start_time < ?
		  AND end_time > ?
		ORDER BY start_time ASC
	`, personID, *workItemID, windowEnd, windowStart)
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

func (r *TimeEntryRepo) DeleteByID(personID, entryID int64) error {
	_, err := r.db.Exec(`DELETE FROM time_entries WHERE person_id = ? AND id = ?`, personID, entryID)
	return err
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
	var entrySource sql.NullString
	var tzName sql.NullString
	var tzOffset sql.NullInt64
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
		&entrySource,
		&tzName,
		&tzOffset,
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
	if entrySource.Valid {
		e.EntrySource = entrySource.String
	}
	if tzName.Valid {
		e.TZName = tzName.String
	}
	if tzOffset.Valid {
		e.TZOffsetMin = int(tzOffset.Int64)
	}
	if updated.Valid {
		v := updated.Int64
		e.UpdatedAt = &v
	}
	return &e, nil
}

func stopwatchWorkItemKey(workItemID *int64) int64 {
	if workItemID == nil {
		return 0
	}
	return *workItemID
}
