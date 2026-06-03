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
