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
	"math"
)

func (s *Session) Start(ctx context.Context, store *Store, task string) (WriteResult, error) {
	if !validUUID(task) {
		return WriteResult{}, Failure("client.invalid_task", "Use a task UUID from remote tasks.", 2)
	}
	if err := s.requireScopes("mcp:read:stopwatch", "mcp:write:stopwatch", "mcp:write:time"); err != nil {
		return WriteResult{}, err
	}
	state, err := s.Client.Stopwatches(ctx)
	if err != nil {
		return WriteResult{}, err
	}
	var existing int64
	for _, watch := range state.Stopwatches {
		if watch.Running() && watch.TaskUUID != task {
			return WriteResult{}, Failure("stopwatch.another_running", "Stop/book or discard the running stopwatch before starting another task.", 4)
		}
		if watch.TaskUUID == task {
			if existing != 0 {
				return WriteResult{}, Failure("client.ambiguous_timer", "Multiple stopwatches exist for this task. Resolve them in the web application first.", 4)
			}
			existing = watch.ID
		}
	}
	tool := "create_stopwatch"
	args := map[string]any{"taskUuid": task, "activate": true, "expectedStateToken": state.StateToken}
	if existing != 0 {
		tool = "activate_stopwatch"
		args = map[string]any{"stopwatchId": existing, "expectedStateToken": state.StateToken}
	}
	return s.mutate(ctx, store, tool, args, func(result WriteResult) bool {
		return result.StateToken != "" && result.Stopwatch != nil && result.Stopwatch.Running() && result.Stopwatch.TaskUUID == task &&
			result.Stopwatch.ID > 0 && result.Mutation.Resources.StopwatchID == result.Stopwatch.ID && (existing == 0 || existing == result.Stopwatch.ID)
	})
}

func (s *Session) Stop(ctx context.Context, store *Store, id int64) (WriteResult, error) {
	if err := s.requireScopes("mcp:read:stopwatch", "mcp:write:stopwatch", "mcp:write:time"); err != nil {
		return WriteResult{}, err
	}
	state, err := s.Client.Stopwatches(ctx)
	if err != nil {
		return WriteResult{}, err
	}
	if id == 0 {
		for _, watch := range state.Stopwatches {
			if watch.Running() {
				if id != 0 {
					return WriteResult{}, Failure("client.ambiguous_timer", "More than one stopwatch is running. Choose its exact --id.", 4)
				}
				id = watch.ID
			}
		}
	}
	if !observedTimer(state, id, true) {
		return WriteResult{}, Failure("stopwatch.not_running", "The selected stopwatch is not running. Refresh remote stopwatches.", 4)
	}
	return s.mutate(ctx, store, "stop_stopwatch", map[string]any{"stopwatchId": id, "expectedStateToken": state.StateToken, "trackTime": true}, func(result WriteResult) bool {
		return result.StateToken != "" && bookedStopwatch(result.Stopwatch) && result.Stopwatch.ID == id &&
			result.Mutation.Resources.StopwatchID == id && result.TimeTracked != nil && *result.TimeTracked
	})
}

// A booked interval may be cleared so the same stopwatch is ready to start again.
func bookedStopwatch(watch *Stopwatch) bool {
	if watch == nil {
		return false
	}
	if watch.StartTime == nil {
		return watch.EndTime == nil && watch.DurationMS == 0
	}
	return watch.EndTime != nil && *watch.EndTime >= *watch.StartTime
}

func (s *Session) Discard(ctx context.Context, store *Store, id int64) (WriteResult, error) {
	if err := s.requireScopes("mcp:read:stopwatch", "mcp:write:stopwatch"); err != nil {
		return WriteResult{}, err
	}
	state, err := s.Client.Stopwatches(ctx)
	if err != nil {
		return WriteResult{}, err
	}
	if !observedTimer(state, id, false) {
		return WriteResult{}, Failure("stopwatch.not_found", "The selected stopwatch does not exist. Refresh remote stopwatches.", 4)
	}
	return s.mutate(ctx, store, "delete_stopwatch", map[string]any{"stopwatchId": id, "expectedStateToken": state.StateToken}, func(result WriteResult) bool {
		return result.StateToken != "" && result.StopwatchID == id && result.Mutation.Resources.StopwatchID == id
	})
}

func observedTimer(state StopwatchState, id int64, running bool) bool {
	if id < 1 || id > math.MaxInt32 {
		return false
	}
	for _, watch := range state.Stopwatches {
		if watch.ID == id {
			return !running || watch.Running()
		}
	}
	return false
}
