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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCapabilitiesUsesAuthenticatedMCPAndExplicitUnits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mcp" || r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Errorf("unexpected authenticated request: %s", r.URL.Path)
		}
		var request struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
			Params struct {
				Name string `json:"name"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		if request.Method != "tools/call" || request.Params.Name != "get_client_capabilities" {
			t.Errorf("unexpected MCP operation: %+v", request)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID,
			"result": map[string]any{"isError": false, "structuredContent": capabilityFixture()}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	caps, err := client.Capabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if caps.AccountIdentity != "user:fixture" || caps.Timezone != "Europe/Berlin" || caps.ContractVersion != 1 {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestUnsafeOriginsAreRejected(t *testing.T) {
	for _, origin := range []string{"http://example.org", "https://user:secret@example.org", "https://example.org/api", "https://example.org?token=secret", "https://example.org#secret", "file:///tmp/token"} {
		t.Run(origin, func(t *testing.T) {
			if _, err := NewClient(origin, "fixture-token", Options{AllowLoopbackHTTP: true}); ExitCode(err) != 2 {
				t.Fatalf("unsafe origin accepted: %v", err)
			}
		})
	}
	if _, err := NewClient("http://127.0.0.1", "fixture-token", Options{}); ExitCode(err) != 2 {
		t.Fatal("implicit HTTP accepted")
	}
}

func TestRedirectNeverForwardsCredentials(t *testing.T) {
	forwarded := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded++ }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
	if _, err := client.Capabilities(context.Background()); err == nil {
		t.Fatal("redirect succeeded")
	}
	if forwarded != 0 {
		t.Fatal("redirect reached another origin")
	}
}

func TestMalformedAndErrorResponsesFailClosedWithoutSecrets(t *testing.T) {
	for name, body := range map[string]string{
		"HTML":               "<html>fixture-token</html>",
		"RPC error":          `{"jsonrpc":"2.0","id":1,"error":{"message":"fixture-token"}}`,
		"missing result":     `{"jsonrpc":"2.0","id":1}`,
		"missing error flag": `{"jsonrpc":"2.0","id":1,"result":{"structuredContent":{}}}`,
		"wrong ID":           `{"jsonrpc":"2.0","id":9,"result":{"isError":false,"structuredContent":{}}}`,
		"oversized":          strings.Repeat("x", maxResponseBytes+1),
		"MCP error":          `{"jsonrpc":"2.0","id":1,"result":{"isError":true,"structuredContent":{"error":{"code":"time.conflict","message":"fixture-token"}}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = fmt.Fprint(w, body) }))
			defer server.Close()
			client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
			_, err := client.Capabilities(context.Background())
			if err == nil || strings.Contains(err.Error(), "fixture-token") {
				t.Fatalf("unsafe result: %v", err)
			}
			if name == "MCP error" && ExitCode(err) != 4 {
				t.Fatalf("wrong conflict classification: %v", err)
			}
		})
	}
}

func TestIncompleteSafetyCapabilitiesAreIncompatible(t *testing.T) {
	for _, field := range []string{"contractVersion", "accountIdentity", "timezone", "scopes", "mutationRecovery", "recoveryWindowSeconds", "conditionalStopwatches", "timestampUnit", "durationUnit"} {
		t.Run(field, func(t *testing.T) {
			caps := capabilityFixture()
			delete(caps, field)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": map[string]any{"isError": false, "structuredContent": caps}})
			}))
			defer server.Close()
			client, _ := NewClient(server.URL, "fixture-token", Options{AllowLoopbackHTTP: true})
			if _, err := client.Capabilities(context.Background()); ExitCode(err) != 7 {
				t.Fatalf("missing %s accepted: %v", field, err)
			}
		})
	}
}

func capabilityFixture() map[string]any {
	return map[string]any{"contractVersion": 1, "accountIdentity": "user:fixture", "timezone": "Europe/Berlin",
		"scopes":        []string{"mcp:read:projects", "mcp:read:time", "mcp:read:stopwatch", "mcp:write:time", "mcp:write:stopwatch"},
		"timestampUnit": "seconds", "durationUnit": "milliseconds", "mutationRecovery": true,
		"recoveryWindowSeconds": 604800, "conditionalStopwatches": true}
}
