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

type ReportRequest struct {
	Mode       string `json:"mode"`
	Month      int    `json:"month"`
	StartMonth int    `json:"startMonth"`
	EndMonth   int    `json:"endMonth"`
	Year       int    `json:"year"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
	ProjectID  int64  `json:"projectId"`
	Language   string `json:"language"`
}

type WorktimeByMonthReport struct {
	Empty         bool                `json:"empty"`
	Rows          []WorktimeReportRow `json:"rows"`
	TotalSeconds  int64               `json:"totalSeconds"`
	TotalDuration string              `json:"totalDuration"`
}

type WorktimeGroupedByProjectReport struct {
	Empty         bool                   `json:"empty"`
	Groups        []WorktimeProjectGroup `json:"groups"`
	TotalSeconds  int64                  `json:"totalSeconds"`
	TotalDuration string                 `json:"totalDuration"`
}

type WorktimeProjectGroup struct {
	ProjectID     int64               `json:"projectId"`
	ProjectName   string              `json:"projectName"`
	Rows          []WorktimeReportRow `json:"rows"`
	TotalSeconds  int64               `json:"totalSeconds"`
	TotalDuration string              `json:"totalDuration"`
}

type WorktimeTaskDetailsReport struct {
	Empty         bool                    `json:"empty"`
	Rows          []WorktimeTaskDetailRow `json:"rows"`
	TotalSeconds  int64                   `json:"totalSeconds"`
	TotalDuration string                  `json:"totalDuration"`
}

type WorktimeProjectDetailsReport struct {
	Empty         bool                `json:"empty"`
	Rows          []WorktimeReportRow `json:"rows"`
	TotalSeconds  int64               `json:"totalSeconds"`
	TotalDuration string              `json:"totalDuration"`
}

type WorktimeTaskDetailRow struct {
	ProjectID       int64  `json:"projectId"`
	ProjectName     string `json:"projectName"`
	TaskID          int64  `json:"taskId"`
	TaskName        string `json:"taskName"`
	DurationSeconds int64  `json:"durationSeconds"`
	Duration        string `json:"duration"`
}

type TimesheetReport struct {
	Empty         bool                  `json:"empty"`
	UserName      string                `json:"userName"`
	ProjectRows   []TimesheetProjectRow `json:"projectRows"`
	DailyRows     []TimesheetDailyRow   `json:"dailyRows"`
	TotalSeconds  int64                 `json:"totalSeconds"`
	TotalDuration string                `json:"totalDuration"`
}

type TimesheetProjectRow struct {
	ProjectID       int64  `json:"projectId"`
	ProjectName     string `json:"projectName"`
	DurationSeconds int64  `json:"durationSeconds"`
	Duration        string `json:"duration"`
}

type TimesheetDailyRow struct {
	Date            string `json:"date"`
	TotalSeconds    int64  `json:"totalSeconds"`
	TotalDuration   string `json:"totalDuration"`
	ProjectSeconds  int64  `json:"projectSeconds"`
	ProjectDuration string `json:"projectDuration"`
}

type WorktimeReportRow struct {
	ProjectID       int64  `json:"projectId"`
	ProjectName     string `json:"projectName"`
	TaskID          int64  `json:"taskId"`
	TaskName        string `json:"taskName"`
	Description     string `json:"description"`
	Date            string `json:"date"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	DurationSeconds int64  `json:"durationSeconds"`
	Duration        string `json:"duration"`
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
		return entriesRepo.MarkStopwatchUnbooked(personID, entryID)
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

func (a *App) GetWorktimeByMonthReport(req ReportRequest) (*WorktimeByMonthReport, error) {
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

	window, err := reportWindow(req, time.Local)
	if err != nil {
		return nil, err
	}
	entriesRepo := repo.NewTimeEntryRepo(database)
	entries, err := entriesRepo.ListOverlapping(personID, window.Start.UTC().Unix(), window.End.UTC().Unix())
	if err != nil {
		return nil, err
	}
	itemsRepo := repo.NewWorkItemRepo(database)
	items, err := itemsRepo.ListAll(personID)
	if err != nil {
		return nil, err
	}
	itemByID := make(map[int64]model.WorkItem, len(items))
	for _, item := range items {
		itemByID[item.ID] = item
	}

	report := &WorktimeByMonthReport{}
	for _, entry := range entries {
		if entry.EndTime == nil {
			continue
		}
		seconds := timeutil.OverlapSeconds(entry.StartTime, *entry.EndTime, window)
		if seconds <= 0 {
			continue
		}
		row := worktimeReportRow(entry, seconds, itemByID, time.Local, req.Language)
		report.Rows = append(report.Rows, row)
		report.TotalSeconds += seconds
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		if report.Rows[i].Date != report.Rows[j].Date {
			return report.Rows[i].Date < report.Rows[j].Date
		}
		return report.Rows[i].StartTime < report.Rows[j].StartTime
	})
	report.Empty = len(report.Rows) == 0
	report.TotalDuration = formatReportDuration(report.TotalSeconds)
	return report, nil
}

func (a *App) GetWorktimeGroupedByProjectReport(req ReportRequest) (*WorktimeGroupedByProjectReport, error) {
	details, err := a.GetWorktimeByMonthReport(req)
	if err != nil {
		return nil, err
	}
	report := &WorktimeGroupedByProjectReport{
		Empty:         details.Empty,
		TotalSeconds:  details.TotalSeconds,
		TotalDuration: details.TotalDuration,
	}
	groupsByProject := map[int64]*WorktimeProjectGroup{}
	var projectOrder []int64
	for _, row := range details.Rows {
		group := groupsByProject[row.ProjectID]
		if group == nil {
			group = &WorktimeProjectGroup{
				ProjectID:   row.ProjectID,
				ProjectName: row.ProjectName,
			}
			groupsByProject[row.ProjectID] = group
			projectOrder = append(projectOrder, row.ProjectID)
		}
		group.Rows = append(group.Rows, row)
		group.TotalSeconds += row.DurationSeconds
		group.TotalDuration = formatReportDuration(group.TotalSeconds)
	}
	sort.Slice(projectOrder, func(i, j int) bool {
		return strings.ToLower(groupsByProject[projectOrder[i]].ProjectName) < strings.ToLower(groupsByProject[projectOrder[j]].ProjectName)
	})
	for _, projectID := range projectOrder {
		report.Groups = append(report.Groups, *groupsByProject[projectID])
	}
	return report, nil
}

func (a *App) GetWorktimeTaskDetailsReport(req ReportRequest) (*WorktimeTaskDetailsReport, error) {
	details, err := a.GetWorktimeByMonthReport(req)
	if err != nil {
		return nil, err
	}
	report := &WorktimeTaskDetailsReport{}
	projectID := req.ProjectID
	if projectID == 0 {
		projectID = firstReportableProjectID(details.Rows)
	}
	rowsByTask := map[int64]*WorktimeTaskDetailRow{}
	var taskOrder []int64
	for _, row := range details.Rows {
		if projectID != 0 && row.ProjectID != projectID {
			continue
		}
		taskID := row.TaskID
		if taskID == 0 {
			taskID = row.ProjectID
		}
		taskRow := rowsByTask[taskID]
		if taskRow == nil {
			taskName := row.TaskName
			if taskName == "" {
				taskName = row.ProjectName
			}
			taskRow = &WorktimeTaskDetailRow{
				ProjectID:   row.ProjectID,
				ProjectName: row.ProjectName,
				TaskID:      taskID,
				TaskName:    taskName,
			}
			rowsByTask[taskID] = taskRow
			taskOrder = append(taskOrder, taskID)
		}
		taskRow.DurationSeconds += row.DurationSeconds
		taskRow.Duration = formatReportDuration(taskRow.DurationSeconds)
		report.TotalSeconds += row.DurationSeconds
	}
	sort.Slice(taskOrder, func(i, j int) bool {
		left := rowsByTask[taskOrder[i]]
		right := rowsByTask[taskOrder[j]]
		if strings.ToLower(left.ProjectName) != strings.ToLower(right.ProjectName) {
			return strings.ToLower(left.ProjectName) < strings.ToLower(right.ProjectName)
		}
		return strings.ToLower(left.TaskName) < strings.ToLower(right.TaskName)
	})
	for _, taskID := range taskOrder {
		report.Rows = append(report.Rows, *rowsByTask[taskID])
	}
	report.Empty = len(report.Rows) == 0
	report.TotalDuration = formatReportDuration(report.TotalSeconds)
	return report, nil
}

func (a *App) GetWorktimeProjectDetailsReport(req ReportRequest) (*WorktimeProjectDetailsReport, error) {
	details, err := a.GetWorktimeByMonthReport(req)
	if err != nil {
		return nil, err
	}
	report := &WorktimeProjectDetailsReport{}
	if req.ProjectID == 0 {
		report.Empty = true
		report.TotalDuration = formatReportDuration(0)
		return report, nil
	}
	for _, row := range details.Rows {
		if row.ProjectID != req.ProjectID {
			continue
		}
		report.Rows = append(report.Rows, row)
		report.TotalSeconds += row.DurationSeconds
	}
	report.Empty = len(report.Rows) == 0
	report.TotalDuration = formatReportDuration(report.TotalSeconds)
	return report, nil
}

func firstReportableProjectID(rows []WorktimeReportRow) int64 {
	if len(rows) == 0 {
		return 0
	}
	projects := map[int64]string{}
	var ids []int64
	for _, row := range rows {
		if _, exists := projects[row.ProjectID]; exists {
			continue
		}
		projects[row.ProjectID] = row.ProjectName
		ids = append(ids, row.ProjectID)
	}
	sort.Slice(ids, func(i, j int) bool {
		return strings.ToLower(projects[ids[i]]) < strings.ToLower(projects[ids[j]])
	})
	return ids[0]
}

func (a *App) GetTimesheetReport(req ReportRequest) (*TimesheetReport, error) {
	database, _, err := a.openDB()
	if err != nil {
		return nil, err
	}
	defer database.Close()
	if err := a.requireInitialized(database); err != nil {
		return nil, err
	}
	personRepo := repo.NewPersonRepo(database)
	person, err := personRepo.GetDefault()
	if err != nil {
		return nil, err
	}
	userName := ""
	if person != nil {
		userName = person.Email
	}

	details, err := a.GetWorktimeByMonthReport(req)
	if err != nil {
		return nil, err
	}
	report := &TimesheetReport{
		Empty:         details.Empty,
		UserName:      userName,
		TotalSeconds:  details.TotalSeconds,
		TotalDuration: details.TotalDuration,
	}
	if req.Mode == "daily" && person != nil {
		dailyRows, err := timesheetDailyRows(database, person.ID, req, time.Local)
		if err != nil {
			return nil, err
		}
		report.DailyRows = dailyRows
		return report, nil
	}
	rowsByProject := map[int64]*TimesheetProjectRow{}
	var projectOrder []int64
	for _, row := range details.Rows {
		projectRow := rowsByProject[row.ProjectID]
		if projectRow == nil {
			projectRow = &TimesheetProjectRow{
				ProjectID:   row.ProjectID,
				ProjectName: row.ProjectName,
			}
			rowsByProject[row.ProjectID] = projectRow
			projectOrder = append(projectOrder, row.ProjectID)
		}
		projectRow.DurationSeconds += row.DurationSeconds
		projectRow.Duration = formatReportDuration(projectRow.DurationSeconds)
	}
	sort.Slice(projectOrder, func(i, j int) bool {
		return strings.ToLower(rowsByProject[projectOrder[i]].ProjectName) < strings.ToLower(rowsByProject[projectOrder[j]].ProjectName)
	})
	for _, projectID := range projectOrder {
		report.ProjectRows = append(report.ProjectRows, *rowsByProject[projectID])
	}
	return report, nil
}

func timesheetDailyRows(database *sql.DB, personID int64, req ReportRequest, loc *time.Location) ([]TimesheetDailyRow, error) {
	window, err := reportWindow(req, loc)
	if err != nil {
		return nil, err
	}
	entries, err := repo.NewTimeEntryRepo(database).ListOverlapping(personID, window.Start.UTC().Unix(), window.End.UTC().Unix())
	if err != nil {
		return nil, err
	}
	secondsByDay := map[string]int64{}
	for _, entry := range entries {
		if entry.EndTime == nil {
			continue
		}
		entryLoc := timeutil.LocationForEntry(entry.TZName, entry.TZOffsetMin, loc)
		for day, seconds := range timeutil.SplitByLocalDay(entry.StartTime, *entry.EndTime, entryLoc) {
			if seconds <= 0 {
				continue
			}
			dayStart, err := time.ParseInLocation("2006-01-02", day, entryLoc)
			if err != nil {
				continue
			}
			if dayStart.Before(window.Start.In(entryLoc)) || !dayStart.Before(window.End.In(entryLoc)) {
				continue
			}
			secondsByDay[day] += seconds
		}
	}
	days := make([]string, 0, len(secondsByDay))
	for day := range secondsByDay {
		days = append(days, day)
	}
	sort.Strings(days)
	rows := make([]TimesheetDailyRow, 0, len(days))
	for _, day := range days {
		seconds := secondsByDay[day]
		rows = append(rows, TimesheetDailyRow{
			Date:            day,
			TotalSeconds:    seconds,
			TotalDuration:   formatReportDuration(seconds),
			ProjectSeconds:  seconds,
			ProjectDuration: formatReportDuration(seconds),
		})
	}
	return rows, nil
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

func reportWindow(req ReportRequest, loc *time.Location) (timeutil.Window, error) {
	if req.Mode == "daily" {
		start, err := time.ParseInLocation("2006-01-02", req.StartDate, loc)
		if err != nil {
			return timeutil.Window{}, errors.New("invalid report start date")
		}
		end, err := time.ParseInLocation("2006-01-02", req.EndDate, loc)
		if err != nil {
			return timeutil.Window{}, errors.New("invalid report end date")
		}
		return timeutil.Window{Start: start, End: end.AddDate(0, 0, 1)}, nil
	}
	startMonth, endMonth, err := normalizeReportMonths(req)
	if err != nil {
		return timeutil.Window{}, err
	}
	start := time.Date(req.Year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)
	end := time.Date(req.Year, time.Month(endMonth), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
	return timeutil.Window{Start: start, End: end}, nil
}

func normalizeReportMonths(req ReportRequest) (int, int, error) {
	startMonth := req.StartMonth
	endMonth := req.EndMonth
	if startMonth == 0 && endMonth == 0 {
		startMonth = req.Month
		endMonth = req.Month
	}
	if startMonth == 0 {
		startMonth = endMonth
	}
	if endMonth == 0 {
		endMonth = startMonth
	}
	if req.Year == 0 || startMonth < 1 || startMonth > 12 || endMonth < 1 || endMonth > 12 || startMonth > endMonth {
		return 0, 0, errors.New("invalid report month")
	}
	return startMonth, endMonth, nil
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

func worktimeReportRow(entry model.TimeEntry, seconds int64, itemByID map[int64]model.WorkItem, loc *time.Location, language string) WorktimeReportRow {
	start := time.Unix(entry.StartTime, 0).In(loc)
	end := time.Unix(*entry.EndTime, 0).In(loc)
	description := ""
	if entry.Description != nil {
		description = *entry.Description
	}
	projectID := int64(0)
	projectName := "Default"
	taskID := int64(0)
	taskName := ""
	if entry.WorkItemID != nil {
		item, ok := itemByID[*entry.WorkItemID]
		if ok {
			if item.ParentID == nil {
				projectID = item.ID
				projectName = item.Name
			} else {
				taskID = item.ID
				taskName = item.Name
				if parent, ok := itemByID[*item.ParentID]; ok {
					projectID = parent.ID
					projectName = parent.Name
				}
			}
		}
	}
	return WorktimeReportRow{
		ProjectID:       projectID,
		ProjectName:     reportWorkItemLabel(projectName, language),
		TaskID:          taskID,
		TaskName:        reportWorkItemLabel(taskName, language),
		Description:     description,
		Date:            start.Format("2006-01-02"),
		StartTime:       start.Format("15:04"),
		EndTime:         end.Format("15:04"),
		DurationSeconds: seconds,
		Duration:        formatReportDuration(seconds),
	}
}

func reportWorkItemLabel(name string, language string) string {
	if language == "en" {
		switch name {
		case "@":
			return "Absences"
		case "@break":
			return "Break"
		case "@overtime":
			return "Overtime compensation"
		case "@public_holiday":
			return "Public holiday"
		case "@sick_leave":
			return "Sick leave"
		case "@vacation":
			return "Vacation"
		}
		return name
	}
	switch name {
	case "@":
		return "Abwesenheiten"
	case "@break":
		return "Pause"
	case "@overtime":
		return "Überstundenausgleich"
	case "@public_holiday":
		return "Feiertag"
	case "@sick_leave":
		return "Krankheit"
	case "@vacation":
		return "Urlaub"
	default:
		return name
	}
}

func formatReportDuration(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
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
	items, err := itemsRepo.ListProjectItems(personID)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.WorkItem, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	included := map[int64]bool{}
	for _, item := range items {
		if !workItemHasActivePath(item, byID) {
			continue
		}
		includeWorkItemPath(item, byID, included)
	}
	out := make([]WorkItem, 0, len(included))
	for _, item := range items {
		if included[item.ID] {
			// Expose Default row too (GUI can choose to hide it and use Default semantics).
			out = append(out, *workItemDTO(item))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func workItemHasActivePath(item model.WorkItem, byID map[int64]model.WorkItem) bool {
	if item.Status != model.WorkItemStatusActive {
		return false
	}
	current := item
	for current.ParentID != nil {
		parent, ok := byID[*current.ParentID]
		if !ok || parent.Status != model.WorkItemStatusActive {
			return false
		}
		current = parent
	}
	return true
}

func includeWorkItemPath(item model.WorkItem, byID map[int64]model.WorkItem, included map[int64]bool) {
	current := item
	for {
		included[current.ID] = true
		if current.ParentID == nil {
			return
		}
		parent, ok := byID[*current.ParentID]
		if !ok {
			return
		}
		current = parent
	}
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

func (a *App) CreateProjectWithTasks(name string, sourceProjectID int64) (*WorkItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}
	if sourceProjectID == 0 {
		return a.CreateProject(name)
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
	source, err := itemsRepo.GetByID(personID, sourceProjectID)
	if err != nil {
		return nil, err
	}
	if source == nil || source.ParentID != nil || strings.EqualFold(source.Name, "Default") || source.Status != model.WorkItemStatusActive {
		return nil, errors.New("source project not found")
	}
	items, err := itemsRepo.ListProjectItems(personID)
	if err != nil {
		return nil, err
	}
	var taskNames []string
	for _, item := range items {
		if item.ParentID == nil || *item.ParentID != sourceProjectID || item.Status != model.WorkItemStatusActive {
			continue
		}
		taskNames = append(taskNames, item.Name)
	}
	sort.Slice(taskNames, func(i, j int) bool {
		return strings.ToLower(taskNames[i]) < strings.ToLower(taskNames[j])
	})

	target, err := a.CreateProject(name)
	if err != nil {
		return nil, err
	}
	for _, taskName := range taskNames {
		if _, err := a.CreateTask(target.ID, taskName); err != nil {
			_ = a.DeleteProject(target.ID)
			return nil, err
		}
	}
	return target, nil
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
		if err := timerRepo.MarkStopwatchConflict(personID, running.ID, end, end-running.StartTime); err != nil {
			return err
		}
		return stopwatchOverlapError(running, now, time.Local)
	}
	return timerRepo.Stop(personID, running.ID, end, end-running.StartTime)
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
		if err := entriesRepo.MarkStopwatchConflict(personID, running.ID, end, end-running.StartTime); err != nil {
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
	dbPath, err := a.databasePath()
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
