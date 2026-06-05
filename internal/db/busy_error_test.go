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
