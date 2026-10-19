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
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Resources struct {
	TimeUUID    string `json:"timeUuid,omitempty"`
	StopwatchID int64  `json:"stopwatchId,omitempty"`
	ProjectUUID string `json:"projectUuid,omitempty"`
}

type Mutation struct {
	ID          string    `json:"id"`
	Operation   string    `json:"operation,omitempty"`
	State       string    `json:"state"`
	ExpiresAtTS int64     `json:"expiresAtTs,omitempty"`
	Resources   Resources `json:"resources"`
	ErrorCode   string    `json:"errorCode,omitempty"`
	Reason      string    `json:"reason,omitempty"`
}

type Attempt struct {
	Version         int       `json:"version"`
	MutationID      string    `json:"mutationId"`
	Profile         string    `json:"profile"`
	Origin          string    `json:"origin"`
	AccountIdentity string    `json:"accountIdentity"`
	Operation       string    `json:"operation"`
	CreatedAt       time.Time `json:"createdAt"`
	State           string    `json:"state"`
	Outcome         *Mutation `json:"outcome,omitempty"`
}

func (s *Store) begin(p Profile, operation string) (*Attempt, error) {
	if err := os.MkdirAll(filepath.Join(s.dir, "operations"), 0700); err != nil {
		return nil, Failure("client.operation_write", "Cannot save recovery metadata; no write was sent.", 5)
	}
	a := &Attempt{Version: 1, MutationID: uuid.NewString(), Profile: p.Name, Origin: p.Origin, AccountIdentity: p.AccountIdentity,
		Operation: operation, CreatedAt: time.Now().UTC(), State: "pending"}
	return a, s.saveAttempt(a)
}

func (s *Store) saveAttempt(a *Attempt) error {
	return writePrivateJSON(filepath.Join(s.dir, "operations", a.MutationID+".json"), a)
}

func (s *Store) attempt(p Profile, id string) (*Attempt, error) {
	if !validUUID(id) {
		return nil, Failure("client.invalid_mutation", "Use a canonical mutation UUID from the original command.", 2)
	}
	data, err := os.ReadFile(filepath.Join(s.dir, "operations", id+".json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, Failure("client.operation_read", "Cannot read recovery metadata.", 5)
	}
	var a Attempt
	if len(data) > 65536 || json.Unmarshal(data, &a) != nil || a.Version != 1 || a.MutationID != id {
		return nil, Failure("client.operation_invalid", "Recovery metadata is invalid.", 5)
	}
	if a.Origin != p.Origin || a.AccountIdentity != p.AccountIdentity {
		return nil, Failure("client.operation_account", "This mutation belongs to another server/account binding.", 3)
	}
	return &a, nil
}

func validMutation(m Mutation, id, operation string) bool {
	if m.ID != id || m.Operation != operation || m.ExpiresAtTS <= 0 {
		return false
	}
	if m.State != "applied" && m.State != "not_applied" {
		return false
	}
	if m.Resources.TimeUUID != "" && !validUUID(m.Resources.TimeUUID) || m.Resources.ProjectUUID != "" && !validUUID(m.Resources.ProjectUUID) || m.Resources.StopwatchID < 0 {
		return false
	}
	return m.ErrorCode == "" || safeCode(m.ErrorCode)
}
