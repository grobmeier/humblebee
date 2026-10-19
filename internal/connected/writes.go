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
	"errors"
	"slices"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

type Booking struct{ ProjectUUID, TaskUUID, Date, Start, End, Note string }

type WriteResult struct {
	Mutation    Mutation   `json:"mutation"`
	Entry       *Entry     `json:"entry,omitempty"`
	Stopwatch   *Stopwatch `json:"stopwatch,omitempty"`
	StopwatchID int64      `json:"stopwatchId,omitempty"`
	StateToken  string     `json:"stateToken,omitempty"`
	TimeTracked *bool      `json:"timeTracked,omitempty"`
}

func (s *Session) requireScopes(scopes ...string) error {
	for _, scope := range scopes {
		if !slices.Contains(s.Capabilities.Scopes, scope) {
			return Failure("mcp.scope_required", "The app token lacks the required scope for this operation.", 3)
		}
	}
	return nil
}

func (b Booking) validate() error {
	if !validUUID(b.ProjectUUID) || !validUUID(b.TaskUUID) {
		return Failure("client.invalid_booking", "Use explicit project and task UUIDs from remote commands.", 2)
	}
	if err := ValidateDate(b.Date); err != nil {
		return err
	}
	start, err := time.Parse("15:04", b.Start)
	if err != nil || len(b.Start) != 5 {
		return Failure("client.invalid_time", "Use 24-hour HH:mm start and end times.", 2)
	}
	end, err := time.Parse("15:04", b.End)
	if b.End != "24:00" && (err != nil || len(b.End) != 5 || !end.After(start)) {
		return Failure("client.invalid_time", "The end must be after the start on the same day.", 2)
	}
	if !utf8.ValidString(b.Note) || len(utf16.Encode([]rune(b.Note))) > 1000 {
		return Failure("client.invalid_note", "Use a UTF-8 note of at most 1,000 UTF-16 code units.", 2)
	}
	return nil
}

func (s *Session) Book(ctx context.Context, store *Store, booking Booking) (WriteResult, error) {
	if err := booking.validate(); err != nil {
		return WriteResult{}, err
	}
	if err := s.requireScopes("mcp:write:time"); err != nil {
		return WriteResult{}, err
	}
	return s.mutate(ctx, store, "save_time_entry", map[string]any{"projectUuid": booking.ProjectUUID, "taskUuid": booking.TaskUUID,
		"startDate": booking.Date, "startTime": booking.Start, "endTime": booking.End, "description": booking.Note}, func(result WriteResult) bool {
		return result.Entry != nil && result.Entry.valid() && result.Entry.ProjectUUID == booking.ProjectUUID && result.Entry.TaskUUID == booking.TaskUUID && result.Mutation.Resources.TimeUUID == result.Entry.UUID
	})
}

func uncertain(id string) *Error {
	err := Failure("client.uncertain", "The write outcome is uncertain. Do not submit it again with a new ID; use remote operation to recover this mutation.", 6)
	err.MutationID = id
	return err
}

func (s *Session) mutate(ctx context.Context, store *Store, tool string, args map[string]any, valid func(WriteResult) bool) (WriteResult, error) {
	a, err := store.begin(s.Profile, tool)
	if err != nil {
		return WriteResult{}, err
	}
	args["clientMutationId"] = a.MutationID
	var result WriteResult
	err = s.Client.call(ctx, tool, args, &result)
	if err != nil {
		var problem *Error
		if errors.As(err, &problem) && problem.definitive {
			problem.MutationID = a.MutationID
			a.State = "rejected"
			if validMutation(result.Mutation, a.MutationID, tool) && result.Mutation.State == "not_applied" {
				a.State = "not_applied"
				a.Outcome = &result.Mutation
			}
			if store.saveAttempt(a) != nil {
				return result, uncertain(a.MutationID)
			}
			return result, problem
		}
		a.State = "unknown"
		_ = store.saveAttempt(a)
		return result, uncertain(a.MutationID)
	}
	if !validMutation(result.Mutation, a.MutationID, tool) || result.Mutation.State != "applied" || !valid(result) {
		a.State = "unknown"
		_ = store.saveAttempt(a)
		return result, uncertain(a.MutationID)
	}
	a.State = "applied"
	a.Outcome = &result.Mutation
	if store.saveAttempt(a) != nil {
		return result, uncertain(a.MutationID)
	}
	return result, nil
}

func (s *Session) Operation(ctx context.Context, store *Store, id string) (Mutation, error) {
	a, err := store.attempt(s.Profile, id)
	if err != nil {
		return Mutation{}, err
	}
	var response struct {
		Mutation Mutation `json:"mutation"`
	}
	if err := s.Client.call(ctx, "get_mutation_status", map[string]any{"clientMutationId": id}, &response); err != nil {
		return Mutation{}, err
	}
	m := response.Mutation
	if m.ID != id {
		return m, uncertain(id)
	}
	if m.State == "unknown" || m.State == "in_progress" {
		return m, uncertain(id)
	}
	if !validMutation(m, id, m.Operation) || a != nil && a.Operation != m.Operation {
		return m, uncertain(id)
	}
	if a != nil {
		a.State = m.State
		a.Outcome = &m
		if store.saveAttempt(a) != nil {
			return m, uncertain(id)
		}
	}
	if m.State == "not_applied" {
		problem := remoteError(m.ErrorCode)
		problem.MutationID = id
		return m, problem
	}
	return m, nil
}
