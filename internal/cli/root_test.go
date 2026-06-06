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
	"errors"
	"strings"
	"testing"

	"github.com/grobmeier/humblebee/internal/db"
)

func TestCLIErrorMessageFormatsBusyError(t *testing.T) {
	err := &db.BusyError{
		Path: "/tmp/humblebee.db",
		Err:  errors.New("database is locked (5) (SQLITE_BUSY)"),
	}

	message := cliErrorMessage(err)

	if !strings.Contains(message, "currently in use") {
		t.Fatalf("expected recovery guidance, got %q", message)
	}
	if !strings.Contains(message, "/tmp/humblebee.db") {
		t.Fatalf("expected database path, got %q", message)
	}
	if strings.Contains(message, db.BusyErrorCode) {
		t.Fatalf("expected user-facing message without internal error code, got %q", message)
	}
}

func TestCLIErrorMessageKeepsRegularErrors(t *testing.T) {
	err := errors.New("not initialized")

	if got := cliErrorMessage(err); got != "not initialized" {
		t.Fatalf("expected regular error text, got %q", got)
	}
}
