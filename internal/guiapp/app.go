package guiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/paths"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/timeutil"
)

type App struct {
	ctx context.Context
}

func New() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

type Dashboard struct {
	Initialized bool          `json:"initialized"`
	DBPath      string        `json:"dbPath"`
	UserEmail   string        `json:"userEmail"`
	Running     *RunningTimer `json:"running"`
	TodayTotal  int64         `json:"todayTotalSeconds"`
}

type RunningTimer struct {
	WorkItemName string `json:"workItemName"`
	StartTimeUTC int64  `json:"startTimeUTC"`
}

type Stopwatch struct {
	ID              int64  `json:"id"`
	WorkItemID      *int64 `json:"workItemId"`
	WorkItemName    string `json:"workItemName"`
	StartDate       string `json:"startDate"`
	StartTime       string `json:"startTime"`
	EndDate         string `json:"endDate"`
	EndTime         string `json:"endTime"`
	DurationSeconds int64  `json:"durationSeconds"`
	Running         bool   `json:"running"`
	Conflicting     bool   `json:"conflicting"`
}

type WorkItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parentId"`
	Depth    int    `json:"depth"`
	Status   string `json:"status"`
}

type StopResult struct {
	WorkItemName      string `json:"workItemName"`
	DurationSeconds   int64  `json:"durationSeconds"`
	TodayTotalSeconds int64  `json:"todayTotalSeconds"`
}

type CreateTimeEntryRequest struct {
	ID            int64  `json:"id"`
	WorkItemID    int64  `json:"workItemId"`
	Description   string `json:"description"`
	StartDate     string `json:"startDate"`
	StartTime     string `json:"startTime"`
	EndDate       string `json:"endDate"`
	EndTime       string `json:"endTime"`
	UntilMidnight bool   `json:"untilMidnight"`
}

type TimeEntry struct {
	ID              int64  `json:"id"`
	WorkItemID      *int64 `json:"workItemId"`
	Description     string `json:"description"`
	StartDate       string `json:"startDate"`
	StartTime       string `json:"startTime"`
	EndDate         string `json:"endDate"`
	EndTime         string `json:"endTime"`
	DurationSeconds int64  `json:"durationSeconds"`
}

type TimeDay struct {
	Date           string      `json:"date"`
	Entries        []TimeEntry `json:"entries"`
	TotalSeconds   int64       `json:"totalSeconds"`
	ProjectSeconds int64       `json:"projectSeconds"`
	AbsenceSeconds int64       `json:"absenceSeconds"`
	WorkSeconds    int64       `json:"workSeconds"`
	BreakSeconds   int64       `json:"breakSeconds"`
}

const stopwatchOverlapErrorCode = "HUMBLEBEE_STOPWATCH_OVERLAP"

func (a *App) GetDashboard() (*Dashboard, error) {
	database, dbPath, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()

	initialized, err := db.IsInitialized(database)
	if err != nil {
		return nil, err
	}
	out := &Dashboard{
		Initialized: initialized,
		DBPath:      dbPath,
	}
	if !initialized {
		return out, nil
	}

	personRepo := repo.NewPersonRepo(database)
	p, err := personRepo.GetDefault()
	if err != nil {
		return nil, err
	}
	if p != nil {
		out.UserEmail = p.Email
	}
	if p == nil {
		return out, nil
	}

	timerRepo := repo.NewTimeEntryRepo(database)
	running, err := timerRepo.FindRunning(p.ID)
	if err != nil {
		return nil, err
	}
	if running != nil {
		name := "Default"
		if running.WorkItemID != nil {
			itemsRepo := repo.NewWorkItemRepo(database)
			item, _ := itemsRepo.GetByID(p.ID, *running.WorkItemID)
			if item != nil {
				name = item.Name
			}
		}
		out.Running = &RunningTimer{
			WorkItemName: name,
			StartTimeUTC: running.StartTime,
		}
	}

	// Today total (current local day window).
	w := timeutil.TodayWindow(time.Now(), time.Local)
	entries, err := timerRepo.ListOverlapping(p.ID, w.Start.UTC().Unix(), w.End.UTC().Unix())
	if err != nil {
		return nil, err
	}
	var total int64
	for _, e := range entries {
		if e.EndTime == nil {
			continue
		}
		total += timeutil.OverlapSeconds(e.StartTime, *e.EndTime, w)
	}
	out.TodayTotal = total
	return out, nil
}

func (a *App) CreateTimeEntry(req CreateTimeEntryRequest) (*TimeEntry, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}

	start, end, err := parseManualEntryTimes(req)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, errors.New("end time must be after start time")
	}

	var workItemID *int64
	if req.WorkItemID != 0 {
		itemsRepo := repo.NewWorkItemRepo(database)
		item, err := itemsRepo.GetByID(personID, req.WorkItemID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, errors.New("work item not found")
		}
		workItemID = &req.WorkItemID
	}

	entriesRepo := repo.NewTimeEntryRepo(database)
	overlaps, err := entriesRepo.HasOverlap(personID, start.UTC().Unix(), end.UTC().Unix())
	if err != nil {
		return nil, err
	}
	if overlaps {
		return nil, errors.New("time entry overlaps with an existing entry")
	}

	_, offsetSec := start.Zone()
	offsetMin := offsetSec / 60
	duration := int64(end.Sub(start).Seconds())
	description := strings.TrimSpace(req.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	endUnix := end.UTC().Unix()
	id, err := entriesRepo.CreateCompleted(model.TimeEntry{
		UUID:        uuid.NewString(),
		PersonID:    personID,
		WorkItemID:  workItemID,
		Description: descriptionPtr,
		StartTime:   start.UTC().Unix(),
		EndTime:     &endUnix,
		Duration:    &duration,
		TZName:      start.Location().String(),
		TZOffsetMin: offsetMin,
		CreatedAt:   time.Now().UTC().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return timeEntryDTO(model.TimeEntry{
		ID:          id,
		WorkItemID:  workItemID,
		Description: descriptionPtr,
		StartTime:   start.UTC().Unix(),
		EndTime:     &endUnix,
		Duration:    &duration,
	}, time.Local), nil
}

func (a *App) UpdateTimeEntry(req CreateTimeEntryRequest) (*TimeEntry, error) {
	if req.ID == 0 {
		return nil, errors.New("time entry id is required")
	}
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}

	start, end, err := parseManualEntryTimes(req)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, errors.New("end time must be after start time")
	}

	var workItemID *int64
	if req.WorkItemID != 0 {
		itemsRepo := repo.NewWorkItemRepo(database)
		item, err := itemsRepo.GetByID(personID, req.WorkItemID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, errors.New("work item not found")
		}
		workItemID = &req.WorkItemID
	}

	entriesRepo := repo.NewTimeEntryRepo(database)
	overlaps, err := entriesRepo.HasOverlapExcluding(personID, req.ID, start.UTC().Unix(), end.UTC().Unix())
	if err != nil {
		return nil, err
	}
	if overlaps {
		return nil, errors.New("time entry overlaps with an existing entry")
	}

	_, offsetSec := start.Zone()
	offsetMin := offsetSec / 60
	duration := int64(end.Sub(start).Seconds())
	description := strings.TrimSpace(req.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	endUnix := end.UTC().Unix()
	err = entriesRepo.UpdateCompleted(model.TimeEntry{
		ID:          req.ID,
		PersonID:    personID,
		WorkItemID:  workItemID,
		Description: descriptionPtr,
		StartTime:   start.UTC().Unix(),
		EndTime:     &endUnix,
		Duration:    &duration,
		TZName:      start.Location().String(),
		TZOffsetMin: offsetMin,
	})
	if err != nil {
		return nil, err
	}

	return timeEntryDTO(model.TimeEntry{
		ID:          req.ID,
		WorkItemID:  workItemID,
		Description: descriptionPtr,
		StartTime:   start.UTC().Unix(),
		EndTime:     &endUnix,
		Duration:    &duration,
	}, time.Local), nil
}

func (a *App) ListStopwatches() ([]Stopwatch, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	entriesRepo := repo.NewTimeEntryRepo(database)
	entries, err := entriesRepo.ListStopwatches(personID, 12)
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	out := make([]Stopwatch, 0, len(entries))
	for _, entry := range entries {
		name := "Default"
		if entry.WorkItemID != nil {
			item, err := itemsRepo.GetByID(personID, *entry.WorkItemID)
			if err == nil && item != nil {
				name = item.Name
			}
		}
		out = append(out, stopwatchDTO(entry, name, time.Local))
	}
	return out, nil
}

func (a *App) DiscardRunningStopwatch() error {
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}

	entriesRepo := repo.NewTimeEntryRepo(database)
	running, err := entriesRepo.FindRunning(personID)
	if err != nil {
		return err
	}
	if running == nil {
		return errors.New("no timer is currently running")
	}
	return entriesRepo.DeleteByID(personID, running.ID)
}

func (a *App) DiscardStopwatch(stopwatchID int64) error {
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}
	return repo.NewTimeEntryRepo(database).CloseStopwatchByEntryID(personID, stopwatchID)
}

func (a *App) DeleteTimeEntry(entryID int64) error {
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}
	entriesRepo := repo.NewTimeEntryRepo(database)
	entry, err := entriesRepo.GetByID(personID, entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return nil
	}
	if entry.EntrySource == "stopwatch" && entry.EndTime != nil {
		return entriesRepo.MarkStopwatchUnbooked(entryID)
	}
	return entriesRepo.DeleteByID(personID, entryID)
}

func (a *App) GetTimeDay(date string) (*TimeDay, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return nil, errors.New("invalid date")
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)

	entriesRepo := repo.NewTimeEntryRepo(database)
	entries, err := entriesRepo.ListOverlapping(personID, start.UTC().Unix(), end.UTC().Unix())
	if err != nil {
		return nil, err
	}

	out := &TimeDay{
		Date:    start.Format("2006-01-02"),
		Entries: make([]TimeEntry, 0, len(entries)),
	}
	for _, entry := range entries {
		if entry.EndTime == nil {
			continue
		}
		seconds := timeutil.OverlapSeconds(entry.StartTime, *entry.EndTime, timeutil.Window{Start: start, End: end})
		out.TotalSeconds += seconds
		out.ProjectSeconds += seconds
		out.WorkSeconds += seconds
		dto := timeEntryDTO(entry, time.Local)
		dto.DurationSeconds = seconds
		out.Entries = append(out.Entries, *dto)
	}
	sort.Slice(out.Entries, func(i, j int) bool {
		if out.Entries[i].StartDate != out.Entries[j].StartDate {
			return out.Entries[i].StartDate < out.Entries[j].StartDate
		}
		return out.Entries[i].StartTime < out.Entries[j].StartTime
	})
	return out, nil
}

func (a *App) Init(email string) error {
	email = strings.TrimSpace(email)
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()

	initialized, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if initialized {
		return errors.New("already initialized")
	}
	if err := db.Migrate(database); err != nil {
		return err
	}
	initSvc := service.NewInitService(database)
	_, _, err = initSvc.Init(service.InitParams{
		Email:           email,
		InitialWorkItem: "",
		Now:             time.Now(),
	})
	return err
}

func parseManualEntryTimes(req CreateTimeEntryRequest) (time.Time, time.Time, error) {
	start, err := parseLocalDateTime(req.StartDate, req.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid start time")
	}
	endDate := req.EndDate
	endTime := req.EndTime
	if req.UntilMidnight {
		endDate = req.StartDate
		endTime = "24:00"
	}
	end, err := parseLocalDateTime(endDate, endTime)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid end time")
	}
	return start, end, nil
}

func parseLocalDateTime(dateValue, timeValue string) (time.Time, error) {
	if timeValue == "24:00" {
		day, err := time.ParseInLocation("2006-01-02", dateValue, time.Local)
		if err != nil {
			return time.Time{}, err
		}
		return day.AddDate(0, 0, 1), nil
	}
	return time.ParseInLocation("2006-01-02 15:04", dateValue+" "+timeValue, time.Local)
}

func timeEntryDTO(entry model.TimeEntry, loc *time.Location) *TimeEntry {
	start := time.Unix(entry.StartTime, 0).In(loc)
	var end time.Time
	if entry.EndTime != nil {
		end = time.Unix(*entry.EndTime, 0).In(loc)
	}
	description := ""
	if entry.Description != nil {
		description = *entry.Description
	}
	duration := int64(0)
	if entry.Duration != nil {
		duration = *entry.Duration
	}
	return &TimeEntry{
		ID:              entry.ID,
		WorkItemID:      entry.WorkItemID,
		Description:     description,
		StartDate:       start.Format("2006-01-02"),
		StartTime:       start.Format("15:04"),
		EndDate:         end.Format("2006-01-02"),
		EndTime:         end.Format("15:04"),
		DurationSeconds: duration,
	}
}

func stopwatchDTO(entry model.TimeEntry, workItemName string, loc *time.Location) Stopwatch {
	start := time.Unix(entry.StartTime, 0).In(loc)
	var end time.Time
	if entry.EndTime != nil {
		end = time.Unix(*entry.EndTime, 0).In(loc)
	}
	duration := int64(0)
	if entry.Duration != nil {
		duration = *entry.Duration
	}
	return Stopwatch{
		ID:              entry.ID,
		WorkItemID:      entry.WorkItemID,
		WorkItemName:    workItemName,
		StartDate:       start.Format("2006-01-02"),
		StartTime:       start.Format("15:04"),
		EndDate:         end.Format("2006-01-02"),
		EndTime:         end.Format("15:04"),
		DurationSeconds: duration,
		Running:         entry.EndTime == nil,
		Conflicting:     entry.EntrySource == "stopwatch_conflict",
	}
}

func (a *App) ListWorkItems() ([]WorkItem, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	items, err := itemsRepo.ListActive(personID)
	if err != nil {
		return nil, err
	}
	out := make([]WorkItem, 0, len(items))
	for _, it := range items {
		// Expose Default row too (GUI can choose to hide it and use Default semantics).
		out = append(out, *workItemDTO(it))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (a *App) ListProjectWorkItems() ([]WorkItem, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	items, err := repo.NewWorkItemRepo(database).ListProjectItems(personID)
	if err != nil {
		return nil, err
	}
	out := make([]WorkItem, 0, len(items))
	for _, item := range items {
		out = append(out, *workItemDTO(item))
	}
	return out, nil
}

func (a *App) CreateProject(name string) (*WorkItem, error) {
	return a.createWorkItem(name, nil)
}

func (a *App) UpdateProject(projectID int64, name string) (*WorkItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}

	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	project, err := itemsRepo.GetByID(personID, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil || project.ParentID != nil || strings.EqualFold(project.Name, "Default") {
		return nil, errors.New("project not found")
	}
	updated, err := itemsRepo.UpdateName(personID, projectID, name)
	if err != nil {
		return nil, err
	}
	return workItemDTO(*updated), nil
}

func (a *App) DeleteProject(projectID int64) error {
	if projectID == 0 {
		return errors.New("project is required")
	}
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	project, err := itemsRepo.GetByID(personID, projectID)
	if err != nil {
		return err
	}
	if project == nil || project.ParentID != nil || strings.EqualFold(project.Name, "Default") {
		return errors.New("project not found")
	}
	return itemsRepo.DeleteProjectAndTimeEntries(personID, projectID)
}

func (a *App) CreateTask(projectID int64, name string) (*WorkItem, error) {
	if projectID == 0 {
		return nil, errors.New("project is required")
	}
	return a.createWorkItem(name, &projectID)
}

func (a *App) SetTaskActive(taskID int64, active bool) (*WorkItem, error) {
	if taskID == 0 {
		return nil, errors.New("task is required")
	}
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	task, err := itemsRepo.GetByID(personID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil || task.ParentID == nil {
		return nil, errors.New("task not found")
	}
	status := model.WorkItemStatusArchived
	if active {
		status = model.WorkItemStatusActive
	}
	updated, err := itemsRepo.SetStatus(personID, taskID, status)
	if err != nil {
		return nil, err
	}
	return workItemDTO(*updated), nil
}

func (a *App) createWorkItem(name string, parentID *int64) (*WorkItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("work item name is required")
	}

	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}

	depth := 0
	itemsRepo := repo.NewWorkItemRepo(database)
	if parentID != nil {
		parent, err := itemsRepo.GetByID(personID, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.ParentID != nil || strings.EqualFold(parent.Name, "Default") {
			return nil, errors.New("project not found")
		}
		depth = parent.Depth + 1
	}

	created, err := itemsRepo.Create(repo.CreateWorkItemParams{
		PersonID: personID,
		UUID:     uuid.NewString(),
		Name:     name,
		ParentID: parentID,
		Depth:    depth,
		Created:  time.Now().UTC().Unix(),
	})
	if err != nil {
		return nil, err
	}
	return workItemDTO(*created), nil
}

func (a *App) Start(workItemID int64) error {
	database, _, err := a.openDB()
	if err != nil {
		return err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return err
	}

	timer := service.NewTimerService(database)
	now := time.Now()
	timerRepo := repo.NewTimeEntryRepo(database)
	running, err := timerRepo.FindRunning(personID)
	if err != nil {
		return err
	}
	var idPtr *int64
	// GUI uses 0 to mean Default (NULL workitem_id).
	if workItemID != 0 {
		idPtr = &workItemID
	}
	_, err = timer.Start(service.StartParams{
		PersonID:   personID,
		WorkItemID: idPtr,
		Now:        now,
	})
	if err != nil {
		return err
	}
	if err := timerRepo.ReopenStopwatchWorkItem(personID, idPtr); err != nil {
		return err
	}

	if running == nil {
		return nil
	}

	end := now.UTC().Unix()
	if end <= running.StartTime {
		return timerRepo.DeleteByID(personID, running.ID)
	}
	overlaps, err := timerRepo.HasOverlap(personID, running.StartTime, end)
	if err != nil {
		return err
	}
	if overlaps {
		if err := timerRepo.MarkStopwatchConflict(running.ID, end, end-running.StartTime); err != nil {
			return err
		}
		return stopwatchOverlapError(running, now, time.Local)
	}
	return timerRepo.Stop(running.ID, end, end-running.StartTime)
}

func workItemDTO(item model.WorkItem) *WorkItem {
	return &WorkItem{
		ID:       item.ID,
		Name:     item.Name,
		ParentID: item.ParentID,
		Depth:    item.Depth,
		Status:   string(item.Status),
	}
}

func (a *App) Stop() (*StopResult, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personID, err := a.defaultPersonID(database)
	if err != nil {
		return nil, err
	}

	timer := service.NewTimerService(database)
	now := time.Now()
	entriesRepo := repo.NewTimeEntryRepo(database)
	running, err := entriesRepo.FindRunning(personID)
	if err != nil {
		return nil, err
	}
	if running == nil {
		return nil, errors.New("no timer is currently running")
	}
	end := now.UTC().Unix()
	overlaps, err := entriesRepo.HasOverlap(personID, running.StartTime, end)
	if err != nil {
		return nil, err
	}
	if overlaps {
		if err := entriesRepo.MarkStopwatchConflict(running.ID, end, end-running.StartTime); err != nil {
			return nil, err
		}
		return nil, stopwatchOverlapError(running, now, time.Local)
	}

	res, err := timer.Stop(personID, now, time.Local)
	if err != nil {
		return nil, err
	}

	name := "Default"
	if res.StoppedEntry.WorkItemID != nil {
		itemsRepo := repo.NewWorkItemRepo(database)
		item, _ := itemsRepo.GetByID(personID, *res.StoppedEntry.WorkItemID)
		if item != nil {
			name = item.Name
		}
	}
	return &StopResult{
		WorkItemName:      name,
		DurationSeconds:   res.DurationSec,
		TodayTotalSeconds: res.TodayTotal,
	}, nil
}

func stopwatchOverlapError(running *model.TimeEntry, end time.Time, loc *time.Location) error {
	start := time.Unix(running.StartTime, 0).In(loc)
	end = end.In(loc)
	workItemID := int64(0)
	if running.WorkItemID != nil {
		workItemID = *running.WorkItemID
	}
	return fmt.Errorf(
		"%s\nStopwatchID: %d\nWorkItemID: %d\nStartDate: %s\nStartTime: %s\nEndDate: %s\nEndTime: %s\nDetails: The running stopwatch overlaps with already booked time.",
		stopwatchOverlapErrorCode,
		running.ID,
		workItemID,
		start.Format("2006-01-02"),
		start.Format("15:04"),
		end.Format("2006-01-02"),
		end.Format("15:04"),
	)
}

func (a *App) openDB() (*sql.DB, string, error) {
	dbPath, err := paths.DBPath()
	if err != nil {
		return nil, "", err
	}
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, "", db.WrapBusyError(dbPath, err)
	}
	initialized, err := db.IsInitialized(database)
	if err != nil {
		_ = database.Close()
		return nil, "", db.WrapBusyError(dbPath, err)
	}
	if initialized {
		if err := db.Migrate(database); err != nil {
			_ = database.Close()
			return nil, "", db.WrapBusyError(dbPath, err)
		}
	}
	return database, dbPath, nil
}

func (a *App) requireInitialized(database *sql.DB) error {
	ok, err := db.IsInitialized(database)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not initialized")
	}
	return nil
}

func (a *App) defaultPersonID(database *sql.DB) (int64, error) {
	people := repo.NewPersonRepo(database)
	p, err := people.GetDefault()
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, fmt.Errorf("no default user found; run init")
	}
	return p.ID, nil
}
