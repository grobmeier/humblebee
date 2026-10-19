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
	"os"
	"path/filepath"
	"testing"

	"github.com/grobmeier/humblebee/internal/connected"
)

func TestRemoteStatusUsesProfileWithoutTouchingLocalDatabase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": map[string]any{
			"contractVersion": 1, "accountIdentity": "user:fixture", "timezone": "Europe/Berlin", "scopes": []string{"mcp:read:time"},
			"timestampUnit": "seconds", "durationUnit": "milliseconds", "mutationRecovery": true, "conditionalStopwatches": true, "recoveryWindowSeconds": 604800}}})
	}))
	defer server.Close()
	home := t.TempDir()
	t.Setenv("HUMBLEBEE_HOME", home)
	dbPath := filepath.Join(home, "humblebee.db")
	original := []byte("fixture local database must not be opened or migrated")
	if err := os.WriteFile(dbPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	store := connected.NewStore(filepath.Join(home, "remote"), nil)
	if _, err := store.Connect(context.Background(), "work", server.URL, "fixture-token", true, false, connected.Options{AllowLoopbackHTTP: true}); err != nil {
		t.Fatal(err)
	}
	cmd := NewCommand(Options{Store: store, Token: func() string { return "fixture-token" }})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"status", "--profile", "work", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var output struct {
		Version int    `json:"version"`
		OK      bool   `json:"ok"`
		Profile string `json:"profile"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %s %v", stdout.String(), err)
	}
	if !output.OK || output.Version != 1 || output.Profile != "work" || stderr.Len() != 0 {
		t.Fatalf("unexpected output: %s %s", stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(dbPath)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatal("remote command touched local database")
	}
}
