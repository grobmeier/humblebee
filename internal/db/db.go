package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const busyTimeoutMillis = 5000

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(%d)", path, busyTimeoutMillis))
	if err != nil {
		return nil, WrapBusyError(path, err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, WrapBusyError(path, err)
	}
	return db, nil
}
