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

package db

import (
	"errors"
	"strings"
	"testing"
)

func TestIsBusyErrorRecognizesSQLiteBusyMessages(t *testing.T) {
	tests := []string{
		"database is locked (5) (SQLITE_BUSY)",
		"constraint failed: database table is locked (6) (SQLITE_LOCKED)",
		"SQLITE_BUSY: database is locked",
		"SQLITE_LOCKED: database table is locked",
	}

	for _, message := range tests {
		if !IsBusyError(errors.New(message)) {
			t.Fatalf("expected busy error for %q", message)
		}
	}
}

func TestIsBusyErrorRejectsOtherSQLiteMessages(t *testing.T) {
	err := errors.New("no such table: time_entries")

	if IsBusyError(err) {
		t.Fatalf("expected non-busy error for %q", err.Error())
	}
}

func TestWrapBusyErrorAddsPathAndPreservesCause(t *testing.T) {
	cause := errors.New("database is locked (5) (SQLITE_BUSY)")
	err := WrapBusyError("/tmp/humblebee.db", cause)

	var busy *BusyError
	if !errors.As(err, &busy) {
		t.Fatalf("expected BusyError, got %T", err)
	}
	if busy.Path != "/tmp/humblebee.db" {
		t.Fatalf("expected path to be preserved, got %q", busy.Path)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped error to preserve cause")
	}
	if !strings.Contains(err.Error(), BusyErrorCode) {
		t.Fatalf("expected stable error code in %q", err.Error())
	}
}

func TestBusyRecoveryMessageIncludesSafeGuidance(t *testing.T) {
	message := BusyRecoveryMessage("/tmp/humblebee.db")

	if !strings.Contains(message, "/tmp/humblebee.db") {
		t.Fatalf("expected database path in message: %q", message)
	}
	if strings.Contains(message, "-wal") || strings.Contains(message, "-shm") {
		t.Fatalf("message must not suggest deleting SQLite sidecar files: %q", message)
	}
}
