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

type fixtureCredentials map[string]string

func (f fixtureCredentials) Get(ref string) (string, error) {
	value, ok := f[ref]
	if !ok {
		return "", errors.New("fixture credential not found")
	}
	return value, nil
}
func (f fixtureCredentials) Set(ref, token string) error { f[ref] = token; return nil }
func (f fixtureCredentials) Delete(ref string) error     { delete(f, ref); return nil }

func capabilityServer(t *testing.T, account *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		caps := capabilityFixture()
		caps["accountIdentity"] = *account
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": caps}})
	}))
}

func TestProfilesKeepTokensOutOfFilesAndRejectAccountChanges(t *testing.T) {
	account := "user:fixture"
	server := capabilityServer(t, &account)
	defer server.Close()
	dir := t.TempDir()
	secrets := fixtureCredentials{}
	store := NewStore(dir, secrets)
	profile, err := store.Connect(context.Background(), "work", server.URL, "private-fixture-token", false, false, Options{AllowLoopbackHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 1 || profile.CredentialRef == "" {
		t.Fatal("missing bound credential")
	}
	data, err := os.ReadFile(filepath.Join(dir, "profiles.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-fixture-token") {
		t.Fatal("token persisted in plaintext")
	}
	if _, err := store.Open(context.Background(), "work", "", Options{}); err != nil {
		t.Fatal(err)
	}
	account = "user:different"
	if _, err := store.Open(context.Background(), "work", "", Options{}); ExitCode(err) != 3 {
		t.Fatalf("different account accepted: %v", err)
	}
	if err := store.Disconnect("work"); err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 0 {
		t.Fatal("credential retained after disconnect")
	}
}

func TestEditedProfileCannotForwardStoredCredentialToAnotherOrigin(t *testing.T) {
	account := "user:fixture"
	original := capabilityServer(t, &account)
	defer original.Close()
	forwarded := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded++
		caps := capabilityFixture()
		var request struct {
			ID int `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"isError": false, "structuredContent": caps}})
	}))
	defer target.Close()
	dir := t.TempDir()
	store := NewStore(dir, fixtureCredentials{})
	if _, err := store.Connect(context.Background(), "work", original.URL, "fixture-token", false, false, Options{AllowLoopbackHTTP: true}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "profiles.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file map[string]any
	if json.Unmarshal(data, &file) != nil {
		t.Fatal("invalid profile file")
	}
	file["profiles"].(map[string]any)["work"].(map[string]any)["origin"] = target.URL
	data, _ = json.Marshal(file)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = store.Open(context.Background(), "work", "", Options{})
	if ExitCode(err) != 3 || forwarded != 0 {
		t.Fatalf("stored credential crossed origin binding: err=%v forwarded=%d", err, forwarded)
	}
}

func TestEphemeralProfilesRequireTokenAndExplicitMultiProfileSelection(t *testing.T) {
	account := "user:fixture"
	server := capabilityServer(t, &account)
	defer server.Close()
	store := NewStore(t.TempDir(), nil)
	options := Options{AllowLoopbackHTTP: true}
	for _, name := range []string{"work", "personal"} {
		if _, err := store.Connect(context.Background(), name, server.URL, "fixture-token", true, false, options); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.Select(""); ExitCode(err) != 2 {
		t.Fatal("ambiguous profile selection accepted")
	}
	if _, err := store.Open(context.Background(), "work", "", Options{}); ExitCode(err) != 3 {
		t.Fatal("missing ephemeral token accepted")
	}
	if _, err := store.Open(context.Background(), "work", "fixture-token", Options{}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Connect(context.Background(), "work", server.URL, "fixture-token", true, false, options); ExitCode(err) != 2 {
		t.Fatal("implicit replacement accepted")
	}
}

func TestUnavailableVaultDoesNotPersistPlaintextOrProfile(t *testing.T) {
	account := "user:fixture"
	server := capabilityServer(t, &account)
	defer server.Close()
	dir := t.TempDir()
	store := NewStore(dir, nil)
	if _, err := store.Connect(context.Background(), "work", server.URL, "private-fixture-token", false, false, Options{AllowLoopbackHTTP: true}); ExitCode(err) != 3 {
		t.Fatal("missing vault accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "profiles.json")); !os.IsNotExist(err) {
		t.Fatal("profile saved despite vault failure")
	}
}
