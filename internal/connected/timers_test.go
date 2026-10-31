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
	"testing"
)

func TestStopAndDiscardTargetObservedTimerWithDifferentBookingIntent(t *testing.T) {
	for _, operation := range []string{"stop", "stop-cleared", "discard"} {
		t.Run(operation, func(t *testing.T) {
			calls := 0
			server := toolServer(t, func(tool string, args map[string]any) any {
				watch := map[string]any{"id": 7, "projectUuid": fixtureProject, "taskUuid": fixtureTask, "projectName": "Client", "title": "Review", "startTime": 1773478800, "duration": 3600000}
				if tool == "list_stopwatches" {
					return map[string]any{"stateToken": "observed", "stopwatches": []any{watch}}
				}
				calls++
				if args["expectedStateToken"] != "observed" || args["stopwatchId"] != float64(7) {
					t.Error("write did not target observed state")
				}
				if operation != "discard" && (tool != "stop_stopwatch" || args["trackTime"] != true) {
					t.Error("stop did not explicitly book")
				}
				if operation == "discard" && (tool != "delete_stopwatch" || args["trackTime"] != nil) {
					t.Error("discard could book time")
				}
				watch["endTime"] = 1773482400
				if operation == "stop-cleared" {
					delete(watch, "startTime")
					delete(watch, "endTime")
					watch["duration"] = 0
				}
				body := map[string]any{"stateToken": "new-state", "mutation": map[string]any{"id": args["clientMutationId"], "operation": tool, "state": "applied", "expiresAtTs": 1774000000, "resources": map[string]any{"stopwatchId": 7}}}
				if operation != "discard" {
					body["stopwatch"] = watch
					body["timeTracked"] = true
				} else {
					body["stopwatchId"] = 7
				}
				return body
			})
			defer server.Close()
			client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
			session := &Session{Profile: Profile{Name: "work", Origin: server.URL, AccountIdentity: "user:fixture"}, Client: client,
				Capabilities: Capabilities{Scopes: []string{"mcp:read:stopwatch", "mcp:write:stopwatch", "mcp:write:time"}}}
			store := NewStore(t.TempDir(), nil)
			var err error
			if operation != "discard" {
				_, err = session.Stop(context.Background(), store, 0)
			} else {
				_, err = session.Discard(context.Background(), store, 7)
			}
			if err != nil || calls != 1 {
				t.Fatalf("%s: %v calls=%d", operation, err, calls)
			}
		})
	}
}

func TestStartDoesNotResetSameTaskOrSwitchAnotherRunningTask(t *testing.T) {
	for _, scenario := range []string{"new", "same", "different"} {
		t.Run(scenario, func(t *testing.T) {
			writes := 0
			server := toolServer(t, func(tool string, args map[string]any) any {
				watch := map[string]any{"id": 7, "projectUuid": fixtureProject, "taskUuid": fixtureTask, "projectName": "Client", "title": "Review", "startTime": 1773478800, "duration": 3600000}
				if tool == "list_stopwatches" {
					rows := []any{}
					if scenario != "new" {
						if scenario == "different" {
							watch["taskUuid"] = fixtureEntry
						}
						rows = append(rows, watch)
					}
					return map[string]any{"stateToken": "observed", "stopwatches": rows}
				}
				writes++
				if args["expectedStateToken"] != "observed" || args["startAtTs"] != nil {
					t.Error("unsafe start precondition")
				}
				if scenario == "new" && (tool != "create_stopwatch" || args["activate"] != true || args["taskUuid"] != fixtureTask) {
					t.Error("incorrect new start")
				}
				if scenario == "same" && (tool != "activate_stopwatch" || args["stopwatchId"] != float64(7)) {
					t.Error("did not reuse existing stopwatch")
				}
				return map[string]any{"stateToken": "observed", "stopwatch": watch, "mutation": map[string]any{"id": args["clientMutationId"], "operation": tool, "state": "applied", "expiresAtTs": 1774000000, "resources": map[string]any{"stopwatchId": 7}}}
			})
			defer server.Close()
			client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
			session := &Session{Profile: Profile{Name: "work", Origin: server.URL, AccountIdentity: "user:fixture"}, Client: client,
				Capabilities: Capabilities{Scopes: []string{"mcp:read:stopwatch", "mcp:write:stopwatch", "mcp:write:time"}}}
			result, err := session.Start(context.Background(), NewStore(t.TempDir(), nil), fixtureTask)
			if scenario == "different" {
				if ExitCode(err) != 4 || writes != 0 {
					t.Fatalf("implicit task switch: %v writes=%d", err, writes)
				}
				return
			}
			if err != nil || writes != 1 || *result.Stopwatch.StartTime != 1773478800 {
				t.Fatalf("timer reset or failed: %+v %v", result, err)
			}
		})
	}
}
