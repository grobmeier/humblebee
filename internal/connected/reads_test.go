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
	"net/http"
	"net/http/httptest"
	"testing"
)

const fixtureProject = "11111111-1111-4111-8111-111111111111"
const fixtureTask = "22222222-2222-4222-8222-222222222222"
const fixtureEntry = "33333333-3333-4333-8333-333333333333"

func toolServer(t *testing.T, respond func(string, map[string]any) any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     int `json:"id"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		body := respond(request.Params.Name, request.Params.Arguments)
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": body}})
	}))
}

func TestReadOperationsFilterUnavailableTasksAndUseExplicitDayUnits(t *testing.T) {
	note := "\u2022 Gr\u00fc\u00dfe\n\u2502 Second line"
	server := toolServer(t, func(tool string, args map[string]any) any {
		switch tool {
		case "get_projects_workspace":
			return map[string]any{"projects": []any{
				map[string]any{"uuid": fixtureProject, "name": "Client accounts", "active": true},
				map[string]any{"uuid": fixtureTask, "name": "Archived", "active": false}}}
		case "get_project_tasks":
			if args["projectUuid"] != fixtureProject {
				t.Error("wrong project")
			}
			return map[string]any{"projectUuid": fixtureProject, "projectName": "Client accounts", "active": true, "tasks": []any{
				map[string]any{"uuid": fixtureTask, "name": "Review", "active": true, "completed": false},
				map[string]any{"uuid": fixtureEntry, "name": "Done", "active": true, "completed": true}}}
		case "get_time_day":
			if args["date"] != "2026-03-14" {
				t.Error("wrong day")
			}
			return map[string]any{"durationMs": 3600, "totalDurationMs": 3600000, "entries": []any{
				map[string]any{"uuid": fixtureEntry, "projectUuid": fixtureProject, "taskUuid": fixtureTask, "projectName": "Client accounts", "taskName": "Review",
					"description": note, "startAtTs": 1773478800, "endAtTs": 1773482400, "durationMs": 3600000}}}
		default:
			t.Errorf("unexpected read: %s", tool)
			return nil
		}
	})
	defer server.Close()
	client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	projects, err := client.Projects(context.Background(), false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("projects: %+v %v", projects, err)
	}
	tasks, err := client.Tasks(context.Background(), fixtureProject, false)
	if err != nil || len(tasks.Tasks) != 1 {
		t.Fatalf("tasks: %+v %v", tasks, err)
	}
	day, err := client.Day(context.Background(), "2026-03-14")
	if err != nil {
		t.Fatal(err)
	}
	if day.TotalDurationMS != 3600000 || day.Entries[0].Note != note {
		t.Fatalf("incorrect units or note: %+v", day)
	}
}
