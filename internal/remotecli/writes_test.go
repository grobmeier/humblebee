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

package remotecli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grobmeier/humblebee/internal/connected"
)

func TestRemoteBookUsesExplicitIDsAndVersionedJSON(t *testing.T) {
	const project = "11111111-1111-4111-8111-111111111111"
	const task = "22222222-2222-4222-8222-222222222222"
	const entry = "33333333-3333-4333-8333-333333333333"
	const note = "\u2022 Review\nSecond line \u202e\x1b[31m"
	writes := 0
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
		var body any
		if request.Params.Name == "get_client_capabilities" {
			body = map[string]any{"contractVersion": 1, "accountIdentity": "user:fixture", "timezone": "Europe/Berlin", "scopes": []string{"mcp:write:time"}, "timestampUnit": "seconds", "durationUnit": "milliseconds", "mutationRecovery": true, "conditionalStopwatches": true, "recoveryWindowSeconds": 604800}
		} else {
			writes++
			args := request.Params.Arguments
			if request.Params.Name != "save_time_entry" || args["projectUuid"] != project || args["taskUuid"] != task || args["complete"] != nil {
				t.Error("unexpected booking intent")
			}
			body = map[string]any{"entry": map[string]any{"uuid": entry, "projectUuid": project, "taskUuid": task, "startAtTs": 1773475200, "endAtTs": 1773478800, "durationMs": 3600000, "description": note},
				"mutation": map[string]any{"id": args["clientMutationId"], "operation": "save_time_entry", "state": "applied", "expiresAtTs": 1774000000, "resources": map[string]any{"timeUuid": entry}}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": body}})
	}))
	defer server.Close()
	store := connected.NewStore(filepath.Join(t.TempDir(), "remote"), nil)
	if _, err := store.Connect(context.Background(), "work", server.URL, "fixture-token", true, false, connected.Options{AllowLoopbackHTTP: true}); err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand(Options{Store: store, Token: func() string { return "fixture-token" }})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"book", "--profile", "work", "--project", project, "--task", task, "--date", "2026-03-14", "--start", "09:00", "--end", "10:00", "--note", note, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result struct {
		OK   bool                  `json:"ok"`
		Data connected.WriteResult `json:"data"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil || !result.OK || result.Data.Entry.UUID != entry || writes != 1 {
		t.Fatalf("incorrect booking output: %s %v writes=%d", output.String(), err, writes)
	}
	if result.Data.Entry.Note != note || strings.ContainsAny(output.String(), "\u202e\x1b") {
		t.Fatalf("JSON lost note content or emitted terminal controls: %q", output.String())
	}
}
