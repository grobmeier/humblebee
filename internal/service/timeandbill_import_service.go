package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
)

const (
	timeAndBillSourceSystem  = "timeandbill"
	timeAndBillExportFormat  = "timeandbill.humblebee.user-time-export"
	timeAndBillSchemaVersion = 1
)

type TimeAndBillImportService struct {
	db       *sql.DB
	workItem *repo.WorkItemRepo
	confirm  func(string) (bool, error)
	now      func() time.Time
}

type TimeAndBillImportOptions struct {
	DryRun          bool
	AssumeYes       bool
	UpdateExisting  bool
	SkipConflicting bool
}

type TimeAndBillImportSummary struct {
	ExportUUID         string
	AlreadyImported    bool
	ProjectsCreated    int
	ProjectsMapped     int
	ProjectsSkipped    int
	TasksCreated       int
	TasksMapped        int
	TasksSkipped       int
	TimeEntriesCreated int
	TimeEntriesUpdated int
	TimeEntriesSkipped int
	TimeEntryConflicts int
	NeedsConfirmation  int
}

type TimeAndBillImportConflict struct {
	TimeEntryUUID string `json:"timeEntryUuid"`
	ProjectName   string `json:"projectName"`
	TaskName      string `json:"taskName"`
	Start         string `json:"start"`
	End           string `json:"end"`
	LocalEntryID  int64  `json:"localEntryId"`
	LocalStart    int64  `json:"localStart"`
	LocalEnd      int64  `json:"localEnd"`
}

type TimeAndBillImportPreview struct {
	ExportUUID      string                      `json:"exportUuid"`
	ExportedAt      string                      `json:"exportedAt"`
	SourceUserEmail string                      `json:"sourceUserEmail"`
	Summary         TimeAndBillImportSummary    `json:"summary"`
	Conflicts       []TimeAndBillImportConflict `json:"conflicts"`
}

type timeAndBillExport struct {
	SchemaVersion int                         `json:"schemaVersion"`
	Format        string                      `json:"format"`
	ExportUUID    string                      `json:"exportUuid"`
	ExportedAt    string                      `json:"exportedAt"`
	User          timeAndBillExportUser       `json:"user"`
	Customers     []timeAndBillExportCustomer `json:"customers"`
	Projects      []timeAndBillExportProject  `json:"projects"`
	Tasks         []timeAndBillExportTask     `json:"tasks"`
	TimeEntries   []timeAndBillExportEntry    `json:"timeEntries"`
}

type timeAndBillExportUser struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
}

type timeAndBillExportCustomer struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type timeAndBillExportProject struct {
	UUID         string  `json:"uuid"`
	CustomerUUID *string `json:"customerUuid"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Active       bool    `json:"active"`
}

type timeAndBillExportTask struct {
	UUID        string  `json:"uuid"`
	ProjectUUID string  `json:"projectUuid"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Active      bool    `json:"active"`
	Complete    bool    `json:"complete"`
}

type timeAndBillExportEntry struct {
	UUID            string  `json:"uuid"`
	TaskUUID        string  `json:"taskUuid"`
	ProjectUUID     string  `json:"projectUuid"`
	Description     *string `json:"description"`
	Start           string  `json:"start"`
	End             *string `json:"end"`
	DurationSeconds *int64  `json:"durationSeconds"`
	Timezone        string  `json:"timezone"`
}

func NewTimeAndBillImportService(database *sql.DB) *TimeAndBillImportService {
	return &TimeAndBillImportService{
		db:       database,
		workItem: repo.NewWorkItemRepo(database),
		confirm: func(message string) (bool, error) {
			return false, fmt.Errorf("confirmation required: %s", message)
		},
		now: time.Now,
	}
}

func (s *TimeAndBillImportService) SetConfirm(confirm func(string) (bool, error)) {
	s.confirm = confirm
}

func (s *TimeAndBillImportService) ImportFile(personID int64, path string, options TimeAndBillImportOptions) (TimeAndBillImportSummary, error) {
	payload, err := readTimeAndBillExport(path)
	if err != nil {
		return TimeAndBillImportSummary{}, err
	}
	return s.Import(personID, payload, options)
}

func (s *TimeAndBillImportService) PreviewFile(personID int64, path string, options TimeAndBillImportOptions) (TimeAndBillImportPreview, error) {
	payload, err := readTimeAndBillExport(path)
	if err != nil {
		return TimeAndBillImportPreview{}, err
	}
	return s.Preview(personID, payload, options)
}

func (s *TimeAndBillImportService) Preview(personID int64, payload timeAndBillExport, options TimeAndBillImportOptions) (TimeAndBillImportPreview, error) {
	options.DryRun = true
	summary, conflicts, err := s.plan(personID, payload, TimeAndBillImportSummary{ExportUUID: payload.ExportUUID}, options)
	if err != nil {
		return TimeAndBillImportPreview{}, err
	}
	return TimeAndBillImportPreview{
		ExportUUID:      payload.ExportUUID,
		ExportedAt:      payload.ExportedAt,
		SourceUserEmail: payload.User.Email,
		Summary:         summary,
		Conflicts:       conflicts,
	}, nil
}

func (s *TimeAndBillImportService) Import(personID int64, payload timeAndBillExport, options TimeAndBillImportOptions) (TimeAndBillImportSummary, error) {
	if err := validateTimeAndBillExport(payload); err != nil {
		return TimeAndBillImportSummary{}, err
	}

	summary := TimeAndBillImportSummary{ExportUUID: payload.ExportUUID}
	alreadyImported, err := s.hasImportRun(payload.ExportUUID)
	if err != nil {
		return summary, err
	}
	summary.AlreadyImported = alreadyImported
	if alreadyImported && !options.UpdateExisting {
		return summary, nil
	}

	if options.DryRun {
		planned, _, err := s.plan(personID, payload, summary, options)
		return planned, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return summary, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.apply(tx, personID, payload, &summary, options); err != nil {
		return summary, err
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return summary, err
	}
	if !alreadyImported {
		if _, err := tx.Exec(`
			INSERT INTO import_runs (export_uuid, source_format, source_user_uuid, source_user_email, imported_at, summary_json)
			VALUES (?, ?, ?, ?, ?, ?)
		`, payload.ExportUUID, payload.Format, payload.User.UUID, payload.User.Email, s.now().Unix(), string(summaryJSON)); err != nil {
			return summary, err
		}
	}
	return summary, tx.Commit()
}

func (s *TimeAndBillImportService) plan(personID int64, payload timeAndBillExport, summary TimeAndBillImportSummary, options TimeAndBillImportOptions) (TimeAndBillImportSummary, []TimeAndBillImportConflict, error) {
	if err := validateTimeAndBillExport(payload); err != nil {
		return summary, nil, err
	}
	alreadyImported, err := s.hasImportRun(payload.ExportUUID)
	if err != nil {
		return summary, nil, err
	}
	summary.AlreadyImported = alreadyImported
	if alreadyImported && !options.UpdateExisting {
		return summary, nil, nil
	}

	projectIDs := map[string]int64{}
	for _, project := range payload.Projects {
		if id, ok, err := s.findMapping(project.UUID, "workitems"); err != nil {
			return summary, nil, err
		} else if ok {
			projectIDs[project.UUID] = id
			summary.ProjectsSkipped++
			continue
		}
		existing, err := s.workItem.FindByNameUnderParent(personID, nil, project.Name)
		if err != nil {
			return summary, nil, err
		}
		if existing != nil {
			summary.NeedsConfirmation++
		} else {
			summary.ProjectsCreated++
		}
	}

	for _, task := range payload.Tasks {
		if _, ok, err := s.findMapping(task.UUID, "workitems"); err != nil {
			return summary, nil, err
		} else if ok {
			summary.TasksSkipped++
		} else {
			summary.TasksCreated++
		}
	}

	conflicts := []TimeAndBillImportConflict{}
	var overlapChecker *importOverlapChecker
	if options.SkipConflicting {
		overlapChecker, err = s.preloadImportOverlaps(personID, payload)
		if err != nil {
			return summary, nil, err
		}
	}
	for _, entry := range payload.TimeEntries {
		if _, ok, err := s.findMapping(entry.UUID, "time_entries"); err != nil {
			return summary, nil, err
		} else if ok {
			if options.UpdateExisting {
				summary.TimeEntriesUpdated++
			} else {
				summary.TimeEntriesSkipped++
			}
		} else {
			if options.SkipConflicting {
				conflict, hasConflict, err := findImportConflict(payload, entry, overlapChecker)
				if err != nil {
					return summary, nil, err
				}
				if hasConflict {
					conflicts = append(conflicts, conflict)
					summary.TimeEntryConflicts++
					continue
				}
			}
			summary.TimeEntriesCreated++
		}
	}
	return summary, conflicts, nil
}

func (s *TimeAndBillImportService) apply(tx *sql.Tx, personID int64, payload timeAndBillExport, summary *TimeAndBillImportSummary, options TimeAndBillImportOptions) error {
	customerIDs := map[string]int64{}
	for _, customer := range payload.Customers {
		if strings.TrimSpace(customer.UUID) == "" || strings.TrimSpace(customer.Name) == "" {
			continue
		}
		status := model.WorkItemStatusActive
		if !customer.Active {
			status = model.WorkItemStatusArchived
		}
		id, created, mapped, err := s.ensureWorkItem(tx, personID, customer.UUID, customer.Name, nil, 0, status, options, true)
		if err != nil {
			return err
		}
		customerIDs[customer.UUID] = id
		if created {
			summary.ProjectsCreated++
		} else if mapped {
			summary.ProjectsMapped++
		} else {
			summary.ProjectsSkipped++
		}
	}

	projectIDs := map[string]int64{}
	projectDepths := map[string]int{}
	projectStatuses := map[string]model.WorkItemStatus{}
	for _, project := range payload.Projects {
		parentID := (*int64)(nil)
		depth := 0
		if project.CustomerUUID != nil {
			if id, ok := customerIDs[*project.CustomerUUID]; ok {
				parentID = &id
				depth = 1
			}
		}
		status := model.WorkItemStatusActive
		if !project.Active {
			status = model.WorkItemStatusArchived
		}
		id, created, mapped, err := s.ensureWorkItem(tx, personID, project.UUID, project.Name, parentID, depth, status, options, true)
		if err != nil {
			return err
		}
		projectIDs[project.UUID] = id
		projectDepths[project.UUID] = depth
		projectStatuses[project.UUID] = status
		if created {
			summary.ProjectsCreated++
		} else if mapped {
			summary.ProjectsMapped++
		} else {
			summary.ProjectsSkipped++
		}
	}

	taskIDs := map[string]int64{}
	for _, task := range payload.Tasks {
		projectID, ok := projectIDs[task.ProjectUUID]
		if !ok {
			return fmt.Errorf("task %q references unknown project UUID %q", task.Name, task.ProjectUUID)
		}
		status := model.WorkItemStatusActive
		if projectStatuses[task.ProjectUUID] != model.WorkItemStatusActive || !task.Active || task.Complete {
			status = model.WorkItemStatusArchived
		}
		id, created, mapped, err := s.ensureWorkItem(tx, personID, task.UUID, task.Name, &projectID, projectDepths[task.ProjectUUID]+1, status, options, false)
		if err != nil {
			return err
		}
		taskIDs[task.UUID] = id
		if created {
			summary.TasksCreated++
		} else if mapped {
			summary.TasksMapped++
		} else {
			summary.TasksSkipped++
		}
	}

	for _, entry := range payload.TimeEntries {
		taskID, ok := taskIDs[entry.TaskUUID]
		if !ok {
			return fmt.Errorf("time entry %s references unknown task UUID %q", entry.UUID, entry.TaskUUID)
		}
		created, updated, skipped, conflicted, err := s.applyTimeEntry(tx, personID, entry, taskID, options)
		if err != nil {
			return err
		}
		if created {
			summary.TimeEntriesCreated++
		}
		if updated {
			summary.TimeEntriesUpdated++
		}
		if skipped {
			summary.TimeEntriesSkipped++
		}
		if conflicted {
			summary.TimeEntryConflicts++
		}
	}
	return nil
}

func (s *TimeAndBillImportService) ensureWorkItem(tx *sql.Tx, personID int64, sourceUUID string, name string, parentID *int64, depth int, status model.WorkItemStatus, options TimeAndBillImportOptions, askOnNameMatch bool) (int64, bool, bool, error) {
	if id, ok, err := findMappingTx(tx, sourceUUID, "workitems"); err != nil {
		return 0, false, false, err
	} else if ok {
		return id, false, false, nil
	}

	existing, err := findWorkItemByNameUnderParentTx(tx, personID, parentID, name)
	if err != nil {
		return 0, false, false, err
	}
	if existing != nil {
		useExisting := !askOnNameMatch || options.AssumeYes
		if askOnNameMatch && !options.AssumeYes {
			useExisting, err = s.confirm(fmt.Sprintf("Use existing project %q for imported Time & Bill project %q?", existing.Name, name))
			if err != nil {
				return 0, false, false, err
			}
		}
		if useExisting {
			if err := insertMappingTx(tx, sourceUUID, "workitems", existing.ID); err != nil {
				return 0, false, false, err
			}
			return existing.ID, false, true, nil
		}
	}

	id, err := createWorkItemTx(tx, personID, uuid.NewString(), name, parentID, depth, status, s.now().Unix())
	if err != nil {
		return 0, false, false, err
	}
	if err := insertMappingTx(tx, sourceUUID, "workitems", id); err != nil {
		return 0, false, false, err
	}
	return id, true, false, nil
}

func (s *TimeAndBillImportService) applyTimeEntry(tx *sql.Tx, personID int64, entry timeAndBillExportEntry, taskID int64, options TimeAndBillImportOptions) (bool, bool, bool, bool, error) {
	existingID, exists, err := findMappingTx(tx, entry.UUID, "time_entries")
	if err != nil {
		return false, false, false, false, err
	}
	start, end, duration, err := parseImportedTimeEntry(entry)
	if err != nil {
		return false, false, false, false, err
	}
	if exists {
		if !options.UpdateExisting {
			return false, false, true, false, nil
		}
		_, err := tx.Exec(`
			UPDATE time_entries
			SET workitem_id = ?, description = ?, start_time = ?, end_time = ?, duration = ?, tz_name = ?, updated_at = ?
			WHERE person_id = ? AND id = ?
		`, taskID, entry.Description, start, end, duration, entry.Timezone, s.now().Unix(), personID, existingID)
		return false, true, false, false, err
	}
	if options.SkipConflicting {
		conflict, err := hasOverlapTx(tx, personID, start, end)
		if err != nil {
			return false, false, false, false, err
		}
		if conflict != nil {
			return false, false, false, true, nil
		}
	}

	res, err := tx.Exec(`
		INSERT INTO time_entries (uuid, person_id, workitem_id, description, start_time, end_time, duration, tz_name, tz_offset_minutes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?)
	`, uuid.NewString(), personID, taskID, entry.Description, start, end, duration, entry.Timezone, s.now().Unix())
	if err != nil {
		return false, false, false, false, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return false, false, false, false, err
	}
	if err := insertMappingTx(tx, entry.UUID, "time_entries", id); err != nil {
		return false, false, false, false, err
	}
	return true, false, false, false, nil
}

func validateTimeAndBillExport(payload timeAndBillExport) error {
	if payload.Format != timeAndBillExportFormat {
		return fmt.Errorf("unsupported import format %q", payload.Format)
	}
	if payload.SchemaVersion != timeAndBillSchemaVersion {
		return fmt.Errorf("unsupported Time & Bill export schema version %d", payload.SchemaVersion)
	}
	if strings.TrimSpace(payload.ExportUUID) == "" {
		return errors.New("exportUuid is required")
	}
	for _, project := range payload.Projects {
		if strings.TrimSpace(project.UUID) == "" || strings.TrimSpace(project.Name) == "" {
			return errors.New("all projects must have uuid and name")
		}
	}
	for _, task := range payload.Tasks {
		if strings.TrimSpace(task.UUID) == "" || strings.TrimSpace(task.ProjectUUID) == "" || strings.TrimSpace(task.Name) == "" {
			return errors.New("all tasks must have uuid, projectUuid, and name")
		}
	}
	for _, entry := range payload.TimeEntries {
		if strings.TrimSpace(entry.UUID) == "" || strings.TrimSpace(entry.TaskUUID) == "" || strings.TrimSpace(entry.Start) == "" {
			return errors.New("all time entries must have uuid, taskUuid, and start")
		}
	}
	return nil
}

func readTimeAndBillExport(path string) (timeAndBillExport, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return timeAndBillExport{}, err
	}
	var payload timeAndBillExport
	if err := json.Unmarshal(body, &payload); err != nil {
		return timeAndBillExport{}, err
	}
	return payload, nil
}

func parseImportedTimeEntry(entry timeAndBillExportEntry) (int64, *int64, *int64, error) {
	start, err := time.Parse(time.RFC3339, entry.Start)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("invalid start for time entry %s: %w", entry.UUID, err)
	}
	var endEpoch *int64
	if entry.End != nil && strings.TrimSpace(*entry.End) != "" {
		end, err := time.Parse(time.RFC3339, *entry.End)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("invalid end for time entry %s: %w", entry.UUID, err)
		}
		value := end.Unix()
		endEpoch = &value
	}
	duration := entry.DurationSeconds
	if duration == nil && endEpoch != nil {
		value := *endEpoch - start.Unix()
		duration = &value
	}
	return start.Unix(), endEpoch, duration, nil
}

type importOverlapChecker struct {
	entries []model.TimeEntry
}

func (s *TimeAndBillImportService) preloadImportOverlaps(personID int64, payload timeAndBillExport) (*importOverlapChecker, error) {
	var minStart int64
	var maxEnd int64
	for _, entry := range payload.TimeEntries {
		start, end, _, err := parseImportedTimeEntry(entry)
		if err != nil {
			return nil, err
		}
		if end == nil {
			continue
		}
		if minStart == 0 || start < minStart {
			minStart = start
		}
		if *end > maxEnd {
			maxEnd = *end
		}
	}
	if minStart == 0 || maxEnd == 0 {
		return &importOverlapChecker{}, nil
	}
	entries, err := repo.NewTimeEntryRepo(s.db).ListOverlapping(personID, minStart, maxEnd)
	if err != nil {
		return nil, err
	}
	return &importOverlapChecker{entries: entries}, nil
}

func findImportConflict(payload timeAndBillExport, entry timeAndBillExportEntry, overlapChecker *importOverlapChecker) (TimeAndBillImportConflict, bool, error) {
	start, end, _, err := parseImportedTimeEntry(entry)
	if err != nil {
		return TimeAndBillImportConflict{}, false, err
	}
	overlap := overlapChecker.firstOverlap(start, end)
	if overlap == nil {
		return TimeAndBillImportConflict{}, false, nil
	}
	projectName, taskName := importNames(payload, entry)
	localEnd := int64(0)
	if overlap.EndTime != nil {
		localEnd = *overlap.EndTime
	}
	endValue := ""
	if entry.End != nil {
		endValue = *entry.End
	}
	return TimeAndBillImportConflict{
		TimeEntryUUID: entry.UUID,
		ProjectName:   projectName,
		TaskName:      taskName,
		Start:         entry.Start,
		End:           endValue,
		LocalEntryID:  overlap.ID,
		LocalStart:    overlap.StartTime,
		LocalEnd:      localEnd,
	}, true, nil
}

func (c *importOverlapChecker) firstOverlap(start int64, end *int64) *model.TimeEntry {
	if c == nil || end == nil {
		return nil
	}
	for i := range c.entries {
		entry := &c.entries[i]
		if entry.EndTime != nil && entry.StartTime < *end && *entry.EndTime > start {
			return entry
		}
	}
	return nil
}

func hasOverlapTx(tx *sql.Tx, personID int64, start int64, end *int64) (*model.TimeEntry, error) {
	if end == nil {
		return nil, nil
	}
	row := tx.QueryRow(`
		SELECT id, uuid, person_id, workitem_id, description, start_time, end_time, duration, entry_source, tz_name, tz_offset_minutes, created_at, updated_at
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND entry_source NOT IN ('stopwatch_conflict', 'stopwatch_unbooked')
		  AND start_time < ?
		  AND end_time > ?
		ORDER BY start_time ASC
		LIMIT 1
	`, personID, *end, start)
	entry, err := scanImportedTimeEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return entry, nil
}

type importedTimeEntryScanner interface {
	Scan(dest ...any) error
}

func scanImportedTimeEntry(s importedTimeEntryScanner) (*model.TimeEntry, error) {
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

func importNames(payload timeAndBillExport, entry timeAndBillExportEntry) (string, string) {
	projectName := ""
	taskName := ""
	for _, project := range payload.Projects {
		if project.UUID == entry.ProjectUUID {
			projectName = project.Name
			break
		}
	}
	for _, task := range payload.Tasks {
		if task.UUID == entry.TaskUUID {
			taskName = task.Name
			break
		}
	}
	return projectName, taskName
}

func (s *TimeAndBillImportService) hasImportRun(exportUUID string) (bool, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM import_runs WHERE export_uuid = ?`, exportUUID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *TimeAndBillImportService) findMapping(sourceUUID string, localTable string) (int64, bool, error) {
	var id int64
	err := s.db.QueryRow(`
		SELECT local_id
		FROM external_mappings
		WHERE source_system = ? AND source_uuid = ? AND local_table = ?
	`, timeAndBillSourceSystem, sourceUUID, localTable).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func insertMappingTx(tx *sql.Tx, sourceUUID string, localTable string, localID int64) error {
	_, err := tx.Exec(`
		INSERT INTO external_mappings (source_system, source_uuid, local_table, local_id, created_at)
		VALUES (?, ?, ?, ?, strftime('%s','now'))
		ON CONFLICT(source_system, source_uuid, local_table)
		DO UPDATE SET local_id = excluded.local_id, updated_at = strftime('%s','now')
	`, timeAndBillSourceSystem, sourceUUID, localTable, localID)
	return err
}

func findMappingTx(tx *sql.Tx, sourceUUID string, localTable string) (int64, bool, error) {
	var id int64
	err := tx.QueryRow(`
		SELECT local_id
		FROM external_mappings
		WHERE source_system = ? AND source_uuid = ? AND local_table = ?
	`, timeAndBillSourceSystem, sourceUUID, localTable).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

type importedWorkItemMatch struct {
	ID   int64
	Name string
}

func findWorkItemByNameUnderParentTx(tx *sql.Tx, personID int64, parentID *int64, name string) (*importedWorkItemMatch, error) {
	var row *sql.Row
	if parentID == nil {
		row = tx.QueryRow(`
			SELECT id, name
			FROM workitems
			WHERE person_id = ? AND parent_id IS NULL AND name = ?
			LIMIT 1`, personID, name)
	} else {
		row = tx.QueryRow(`
			SELECT id, name
			FROM workitems
			WHERE person_id = ? AND parent_id = ? AND name = ?
			LIMIT 1`, personID, *parentID, name)
	}
	var item importedWorkItemMatch
	if err := row.Scan(&item.ID, &item.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func createWorkItemTx(tx *sql.Tx, personID int64, localUUID string, name string, parentID *int64, depth int, status model.WorkItemStatus, createdAt int64) (int64, error) {
	res, err := tx.Exec(`
		INSERT INTO workitems (uuid, person_id, name, parent_id, depth, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, localUUID, personID, name, parentID, depth, string(status), createdAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	path := fmt.Sprintf("%d", id)
	if parentID != nil {
		var parentPath string
		if err := tx.QueryRow(`SELECT path FROM workitems WHERE person_id = ? AND id = ?`, personID, *parentID).Scan(&parentPath); err != nil {
			return 0, err
		}
		if parentPath == "" {
			return 0, errors.New("parent has no path")
		}
		path = parentPath + "/" + fmt.Sprintf("%d", id)
	}
	if _, err := tx.Exec(`UPDATE workitems SET path = ? WHERE person_id = ? AND id = ?`, path, personID, id); err != nil {
		return 0, err
	}
	return id, nil
}
