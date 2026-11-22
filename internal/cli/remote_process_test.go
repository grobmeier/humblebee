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

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/grobmeier/humblebee/internal/connected"
)

func TestRemoteProcessHelper(t *testing.T) {
	if os.Getenv("HUMBLEBEE_CLI_TEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"humblebee"}, os.Args[i+1:]...)
			break
		}
	}
	os.Exit(ExitCode(Execute()))
}

func TestRemoteWriteRecoversAcrossCLIProcessesWithoutReplay(t *testing.T) {
	const project = "11111111-1111-4111-8111-111111111111"
	const task = "22222222-2222-4222-8222-222222222222"
	const entry = "33333333-3333-4333-8333-333333333333"
	var mu sync.Mutex
	writes := 0
	var mutationID string
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
		switch request.Params.Name {
		case "get_client_capabilities":
			body = map[string]any{"contractVersion": 1, "accountIdentity": "user:fixture", "timezone": "Europe/Berlin", "scopes": []string{"mcp:write:time"}, "timestampUnit": "seconds", "durationUnit": "milliseconds", "mutationRecovery": true, "conditionalStopwatches": true, "recoveryWindowSeconds": 604800}
		case "save_time_entry":
			mu.Lock()
			writes++
			mutationID, _ = request.Params.Arguments["clientMutationId"].(string)
			mu.Unlock()
			connection, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = connection.Close()
			return
		case "get_mutation_status":
			mu.Lock()
			id := mutationID
			mu.Unlock()
			if request.Params.Arguments["clientMutationId"] != id {
				t.Error("recovery changed mutation ID")
			}
			body = map[string]any{"mutation": map[string]any{"id": id, "operation": "save_time_entry", "state": "applied", "expiresAtTs": 1774000000, "resources": map[string]any{"timeUuid": entry}}}
		default:
			t.Errorf("unexpected remote tool: %s", request.Params.Name)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": body}})
	}))
	defer server.Close()
	home := t.TempDir()
	store := connected.NewStore(filepath.Join(home, "remote"), nil)
	if _, err := store.Connect(context.Background(), "work", server.URL, "fixture-token", true, false, connected.Options{AllowLoopbackHTTP: true}); err != nil {
		t.Fatal(err)
	}
	database := []byte("local fixture must remain unchanged")
	if err := os.WriteFile(filepath.Join(home, "humblebee.db"), database, 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) ([]byte, int) {
		cmd := exec.CommandContext(t.Context(), os.Args[0], append([]string{"-test.run=^TestRemoteProcessHelper$", "--", "remote"}, args...)...)
		cmd.Env = append(os.Environ(), "HUMBLEBEE_CLI_TEST_HELPER=1", "HUMBLEBEE_HOME="+home, "HUMBLEBEE_TAB_TOKEN=fixture-token")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if stderr.Len() != 0 {
			t.Fatalf("unexpected stderr: %s", stderr.String())
		}
		if err == nil {
			return output, 0
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		return output, exit.ExitCode()
	}
	output, exit := run("book", "--profile", "work", "--project", project, "--task", task, "--date", "2026-03-14", "--start", "09:00", "--end", "10:00", "--json")
	var lost struct {
		OK    bool `json:"ok"`
		Error struct {
			MutationID string `json:"mutationId"`
		} `json:"error"`
	}
	if json.Unmarshal(output, &lost) != nil || exit != 6 || lost.OK || lost.Error.MutationID == "" {
		t.Fatalf("missing uncertainty: %d %s", exit, output)
	}
	output, exit = run("operation", "--profile", "work", "--json", lost.Error.MutationID)
	var recovered struct {
		OK   bool               `json:"ok"`
		Data connected.Mutation `json:"data"`
	}
	if json.Unmarshal(output, &recovered) != nil || exit != 0 || !recovered.OK || recovered.Data.Resources.TimeUUID != entry {
		t.Fatalf("recovery failed: %d %s", exit, output)
	}
	mu.Lock()
	count := writes
	mu.Unlock()
	if count != 1 {
		t.Fatalf("write replayed %d times", count)
	}
	current, err := os.ReadFile(filepath.Join(home, "humblebee.db"))
	if err != nil || !bytes.Equal(current, database) {
		t.Fatal("remote write/recovery changed local database")
	}
}

func TestRemoteProcessErrorsRemainJSONAndNeverInitializeLocalDatabase(t *testing.T) {
	for _, args := range [][]string{
		{"remote", "status", "--profile", "missing", "--json"},
		{"remote", "status", "--bad", "--json"},
		{"remote", "not-a-command", "--json"},
	} {
		t.Run(args[1]+args[2], func(t *testing.T) {
			home := t.TempDir()
			cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestRemoteProcessHelper$", "--"}, args...)...)
			cmd.Env = append(os.Environ(), "HUMBLEBEE_CLI_TEST_HELPER=1", "HUMBLEBEE_HOME="+home, "HUMBLEBEE_TAB_TOKEN=")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 2 {
				t.Fatalf("unexpected exit %v: %s %s", err, output, stderr.String())
			}
			var result struct {
				Version int  `json:"version"`
				OK      bool `json:"ok"`
			}
			if json.Unmarshal(output, &result) != nil || result.Version != 1 || result.OK || stderr.Len() != 0 {
				t.Fatalf("invalid error output: %s %s", output, stderr.String())
			}
			if _, err := os.Stat(filepath.Join(home, "humblebee.db")); !os.IsNotExist(err) {
				t.Fatal("remote error created local database")
			}
		})
	}
}
