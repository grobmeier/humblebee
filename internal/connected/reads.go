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

package connected

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Task struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	Completed bool   `json:"completed"`
}

type ProjectTasks struct {
	ProjectUUID string `json:"projectUuid"`
	ProjectName string `json:"projectName"`
	Active      bool   `json:"active"`
	Tasks       []Task `json:"tasks"`
}

type Entry struct {
	UUID        string `json:"uuid"`
	ProjectUUID string `json:"projectUuid"`
	ProjectName string `json:"projectName"`
	TaskUUID    string `json:"taskUuid"`
	TaskName    string `json:"taskName"`
	Note        string `json:"description"`
	StartAtTS   int64  `json:"startAtTs"`
	EndAtTS     int64  `json:"endAtTs"`
	DurationMS  int64  `json:"durationMs"`
}

func (e *Entry) UnmarshalJSON(data []byte) error {
	type plainEntry Entry
	var wire struct {
		*plainEntry
		Start    *int64 `json:"startAtTs"`
		End      *int64 `json:"endAtTs"`
		Duration *int64 `json:"durationMs"`
	}
	wire.plainEntry = (*plainEntry)(e)
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Start == nil || wire.End == nil || wire.Duration == nil {
		return invalidData()
	}
	e.StartAtTS = *wire.Start
	e.EndAtTS = *wire.End
	e.DurationMS = *wire.Duration
	return nil
}

type Day struct {
	Date            string  `json:"date"`
	TotalDurationMS int64   `json:"totalDurationMs"`
	Entries         []Entry `json:"entries"`
}

type Stopwatch struct {
	ID          int64  `json:"id"`
	ProjectUUID string `json:"projectUuid"`
	TaskUUID    string `json:"taskUuid"`
	ProjectName string `json:"projectName"`
	TaskName    string `json:"title"`
	StartTime   *int64 `json:"startTime"`
	EndTime     *int64 `json:"endTime"`
	DurationMS  int64  `json:"duration"`
}

func (s Stopwatch) Running() bool { return s.StartTime != nil && s.EndTime == nil }

type StopwatchState struct {
	StateToken  string      `json:"stateToken"`
	Stopwatches []Stopwatch `json:"stopwatches"`
}

func validUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func invalidData() error {
	return Failure("client.invalid_response", "Time & Bill returned incomplete or inconsistent data.", 5)
}

func (c *Client) Projects(ctx context.Context, all bool) ([]Project, error) {
	var response struct {
		Projects []struct {
			UUID, Name string
			Active     *bool
		} `json:"projects"`
	}
	if err := c.call(ctx, "get_projects_workspace", map[string]any{}, &response); err != nil {
		return nil, err
	}
	if response.Projects == nil {
		return nil, invalidData()
	}
	result := []Project{}
	for _, p := range response.Projects {
		if !validUUID(p.UUID) || p.Name == "" || p.Active == nil {
			return nil, invalidData()
		}
		if all || *p.Active && p.Name != "@" {
			result = append(result, Project{UUID: p.UUID, Name: p.Name, Active: *p.Active})
		}
	}
	return result, nil
}

func (c *Client) Tasks(ctx context.Context, project string, all bool) (ProjectTasks, error) {
	if !validUUID(project) {
		return ProjectTasks{}, Failure("client.invalid_project", "Use a project UUID from remote projects.", 2)
	}
	var wire struct {
		ProjectUUID string `json:"projectUuid"`
		ProjectName string `json:"projectName"`
		Active      *bool  `json:"active"`
		Tasks       []struct {
			UUID, Name        string
			Active, Completed *bool
		} `json:"tasks"`
	}
	if err := c.call(ctx, "get_project_tasks", map[string]any{"projectUuid": project}, &wire); err != nil {
		return ProjectTasks{}, err
	}
	if wire.ProjectUUID != project || wire.ProjectName == "" || wire.Active == nil || wire.Tasks == nil {
		return ProjectTasks{}, invalidData()
	}
	result := ProjectTasks{ProjectUUID: project, ProjectName: wire.ProjectName, Active: *wire.Active, Tasks: []Task{}}
	for _, t := range wire.Tasks {
		if !validUUID(t.UUID) || t.Name == "" || t.Active == nil || t.Completed == nil {
			return ProjectTasks{}, invalidData()
		}
		if all || *wire.Active && wire.ProjectName != "@" && *t.Active && !*t.Completed {
			result.Tasks = append(result.Tasks, Task{UUID: t.UUID, Name: t.Name, Active: *t.Active, Completed: *t.Completed})
		}
	}
	return result, nil
}

func ValidateDate(value string) error {
	date, err := time.Parse("2006-01-02", value)
	if err != nil || date.Year() < 1 || date.Format("2006-01-02") != value {
		return Failure("client.invalid_date", "Use a YYYY-MM-DD date.", 2)
	}
	return nil
}

func (e Entry) valid() bool {
	return validUUID(e.UUID) && validUUID(e.ProjectUUID) && validUUID(e.TaskUUID) &&
		e.StartAtTS >= -62135596800 && e.EndAtTS <= 253402300799 && e.EndAtTS >= e.StartAtTS &&
		e.DurationMS >= 0 && (e.EndAtTS-e.StartAtTS)*1000 == e.DurationMS
}

func (c *Client) Day(ctx context.Context, date string) (Day, error) {
	if err := ValidateDate(date); err != nil {
		return Day{}, err
	}
	var wire struct {
		Total   *int64  `json:"totalDurationMs"`
		Entries []Entry `json:"entries"`
	}
	if err := c.call(ctx, "get_time_day", map[string]any{"date": date}, &wire); err != nil {
		return Day{}, err
	}
	if wire.Total == nil || *wire.Total < 0 || wire.Entries == nil {
		return Day{}, invalidData()
	}
	var total int64
	for _, e := range wire.Entries {
		if !e.valid() || math.MaxInt64-total < e.DurationMS {
			return Day{}, invalidData()
		}
		total += e.DurationMS
	}
	if total != *wire.Total {
		return Day{}, invalidData()
	}
	return Day{Date: date, TotalDurationMS: total, Entries: wire.Entries}, nil
}

func (c *Client) Stopwatches(ctx context.Context) (StopwatchState, error) {
	var state StopwatchState
	if err := c.call(ctx, "list_stopwatches", map[string]any{}, &state); err != nil {
		return state, err
	}
	if state.StateToken == "" || state.Stopwatches == nil {
		return state, invalidData()
	}
	seen := map[int64]bool{}
	for _, s := range state.Stopwatches {
		if s.ID <= 0 || seen[s.ID] || !validUUID(s.ProjectUUID) || !validUUID(s.TaskUUID) || s.DurationMS < 0 {
			return state, invalidData()
		}
		seen[s.ID] = true
	}
	return state, nil
}
