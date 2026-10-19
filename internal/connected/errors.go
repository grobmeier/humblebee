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

import "errors"

// Error is a safe diagnostic. Raw HTTP errors and server text may contain secrets.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Exit       int    `json:"-"`
	MutationID string `json:"mutationId,omitempty"`
	definitive bool
}

func (e *Error) Error() string { return e.Message }

func Failure(code, message string, exit int) *Error {
	return &Error{Code: code, Message: message, Exit: exit}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var problem *Error
	if errors.As(err, &problem) {
		return problem.Exit
	}
	return 5
}

func remoteError(code string) *Error {
	exit := 5
	switch {
	case code == "mcp.scope_required", code == "auth.subscription_required", code == "time.subscription_required":
		exit = 3
	case code == "client.mutation_expired", code == "client.mutation_mismatch":
		exit = 4
	case code == "time.conflict", code == "time.booking_closed", code == "stopwatch.stale_state", code == "stopwatch.another_running":
		exit = 4
	}
	return Failure(code, "Time & Bill rejected the operation. Check permissions and refresh the remote state.", exit)
}
