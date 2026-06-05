package db

import (
	"errors"
	"fmt"
	"strings"
)

const BusyErrorCode = "HUMBLEBEE_DATABASE_BUSY"

type BusyError struct {
	Path string
	Err  error
}

func (e *BusyError) Error() string {
	return fmt.Sprintf("%s\nDatabase: %s\nDetails: %v", BusyErrorCode, e.Path, e.Err)
}

func (e *BusyError) Unwrap() error {
	return e.Err
}

func WrapBusyError(path string, err error) error {
	if err == nil {
		return nil
	}
	var busy *BusyError
	if errors.As(err, &busy) {
		return err
	}
	if !IsBusyError(err) {
		return err
	}
	return &BusyError{Path: path, Err: err}
}

func IsBusyError(err error) bool {
	if err == nil {
		return false
	}
	var busy *BusyError
	if errors.As(err, &busy) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "sqlite_busy") ||
		strings.Contains(message, "sqlite_locked") ||
		strings.Contains(message, "database is locked") ||
		strings.Contains(message, "database table is locked")
}

func BusyRecoveryMessage(path string) string {
	return fmt.Sprintf("HumbleBee could not access the local database because it is currently in use.\n\nDatabase: %s\n\nClose other HumbleBee windows or wait for other commands, backups, or sync tools to finish, then retry.", path)
}
