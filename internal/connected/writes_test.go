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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBookingRecordsRecoveryIDBeforeDispatchWithoutNotePayload(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir, nil)
	note := "\u2022 Gr\u00fc\u00dfe\n\u2502 private note"
	var mutationID string
	server := toolServer(t, func(tool string, args map[string]any) any {
		if tool != "save_time_entry" {
			t.Errorf("unexpected write: %s", tool)
		}
		mutationID, _ = args["clientMutationId"].(string)
		if !validUUID(mutationID) || args["description"] != note || args["complete"] == true {
			t.Error("invalid booking intent")
		}
		data, err := os.ReadFile(filepath.Join(dir, "operations", mutationID+".json"))
		if err != nil {
			t.Errorf("mutation not durable before dispatch: %v", err)
		}
		if strings.Contains(string(data), "private note") {
			t.Error("payload persisted")
		}
		return map[string]any{"entry": map[string]any{"uuid": fixtureEntry, "projectUuid": fixtureProject, "taskUuid": fixtureTask,
			"startAtTs": 1773478800, "endAtTs": 1773482400, "durationMs": 3600000, "description": note},
			"mutation": map[string]any{"id": mutationID, "operation": tool, "state": "applied", "expiresAtTs": 1774000000,
				"resources": map[string]any{"timeUuid": fixtureEntry}}}
	})
	defer server.Close()
	client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	session := &Session{Profile: Profile{Name: "work", Origin: server.URL, AccountIdentity: "user:fixture", Timezone: "Europe/Berlin"},
		Capabilities: Capabilities{Scopes: []string{"mcp:write:time"}}, Client: client}
	result, err := session.Book(context.Background(), store, Booking{ProjectUUID: fixtureProject, TaskUUID: fixtureTask,
		Date: "2026-03-14", Start: "09:00", End: "10:00", Note: note})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mutation.ID != mutationID || result.Entry.UUID != fixtureEntry {
		t.Fatalf("wrong write result: %+v", result)
	}
	data, err := os.ReadFile(filepath.Join(dir, "operations", mutationID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var attempt map[string]any
	if json.Unmarshal(data, &attempt) != nil || attempt["state"] != "applied" || strings.Contains(string(data), "private note") {
		t.Fatalf("unsafe recovery metadata: %s", data)
	}
}

func TestLostBookingResponseCanBeRecoveredAfterRestartWithoutAnotherWrite(t *testing.T) {
	dir := t.TempDir()
	writes := 0
	var id string
	committed := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if request.Params.Name == "save_time_entry" {
			writes++
			id, _ = request.Params.Arguments["clientMutationId"].(string)
			close(committed)
			connection, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = connection.Close()
			return
		}
		if request.Params.Name != "get_mutation_status" || request.Params.Arguments["clientMutationId"] != id {
			t.Error("recovery did not use original mutation ID")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": map[string]any{
			"mutation": map[string]any{"id": id, "operation": "save_time_entry", "state": "applied", "expiresAtTs": 1774000000, "resources": map[string]any{"timeUuid": fixtureEntry}}}}})
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	session := &Session{Profile: Profile{Name: "work", Origin: server.URL, AccountIdentity: "user:fixture"}, Client: client, Capabilities: Capabilities{Scopes: []string{"mcp:write:time"}}}
	_, err := session.Book(context.Background(), NewStore(dir, nil), Booking{ProjectUUID: fixtureProject, TaskUUID: fixtureTask, Date: "2026-03-14", Start: "09:00", End: "10:00"})
	<-committed
	var uncertainError *Error
	if !errors.As(err, &uncertainError) || uncertainError.Exit != 6 || uncertainError.MutationID != id {
		t.Fatalf("missing recovery guidance: %v", err)
	}
	client, _ = NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	session.Client = client
	mutation, err := session.Operation(context.Background(), NewStore(dir, nil), id)
	if err != nil || mutation.State != "applied" || writes != 1 {
		t.Fatalf("unsafe recovery: %+v %v writes=%d", mutation, err, writes)
	}
}

func TestCancelledBeforeDispatchDoesNotCreateAWriteAttempt(t *testing.T) {
	dir := t.TempDir()
	client, _ := NewClient("https://example.invalid", "fixture-token", Options{})
	session := &Session{Client: client, Capabilities: Capabilities{Scopes: []string{"mcp:write:time"}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := session.Book(ctx, NewStore(dir, nil), Booking{ProjectUUID: fixtureProject, TaskUUID: fixtureTask, Date: "2026-03-14", Start: "09:00", End: "10:00"})
	if ExitCode(err) != 5 {
		t.Fatalf("pre-dispatch cancellation should be a known unsent request: %v", err)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 0 {
		t.Fatalf("unsent request created recovery state: %v %v", files, err)
	}
}
