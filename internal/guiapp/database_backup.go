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

package guiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Selection is persisted process-wide, so this also coordinates separate App instances.
var databaseSelectionMu sync.RWMutex

// BackupDatabase returns an empty path without an error when the dialog is cancelled.
// GUI preferences and news state live outside SQLite and are not included.
func (a *App) BackupDatabase(expectedDatabasePath string) (string, error) {
	return a.backupDatabase(expectedDatabasePath, runtime.SaveFileDialog)
}

func (a *App) backupDatabase(expectedDatabasePath string, dialog func(context.Context, runtime.SaveDialogOptions) (string, error)) (string, error) {
	databaseSelectionMu.RLock()
	defer databaseSelectionMu.RUnlock()
	source, err := a.expectedDatabasePath(expectedDatabasePath)
	if err != nil {
		return "", err
	}
	if err := selectedDatabasePathExists(source); err != nil {
		return "", err
	}
	name := strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	target, err := dialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Back up HumbleBee database",
		DefaultFilename:      name + "-" + time.Now().Format("20060102-150405") + ".db",
		CanCreateDirectories: true,
		Filters:              []runtime.FileFilter{{DisplayName: "SQLite database (*.db)", Pattern: "*.db"}},
	})
	if err != nil || target == "" {
		return "", err
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return snapshotDatabase(ctx, source, target)
}

// Caller holds databaseSelectionMu; nested lock order is selection then settings.
func (a *App) expectedDatabasePath(expected string) (string, error) {
	if expected == "" {
		return "", errors.New("expected database path is required")
	}
	guiSettingsMu.Lock()
	defer guiSettingsMu.Unlock()
	current, err := a.databasePath()
	if err != nil {
		return "", err
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return "", err
	}
	expected, err = filepath.Abs(expected)
	if err != nil {
		return "", err
	}
	if current != expected {
		return "", errors.New("database context changed; refresh and try again")
	}
	return current, nil
}

func snapshotDatabase(ctx context.Context, source, target string) (string, error) {
	source, err := filepath.Abs(source)
	if err != nil {
		return "", err
	}
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if !sourceInfo.Mode().IsRegular() {
		return "", errors.New("source database must be a regular file")
	}
	target, err = normalizeRequiredPath(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(target); err == nil {
		return "", errors.New("backup destination already exists; choose a new filename")
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// Never VACUUM directly into a user-chosen path: SQLite accepts existing empty
	// files. Publish only the completed snapshot, without replacing any destination.
	staging, err := os.MkdirTemp(filepath.Dir(target), ".humblebee-backup-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(staging)
	snapshot := filepath.Join(staging, "snapshot.db")
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(source)}
	q := url.Values{"mode": {"ro"}, "_pragma": {"busy_timeout(5000)"}}
	u.RawQuery = q.Encode()
	database, err := sql.Open("sqlite", u.String())
	if err != nil {
		return "", err
	}
	defer database.Close()
	if _, err := database.ExecContext(ctx, "VACUUM main INTO ?", snapshot); err != nil {
		return "", db.WrapBusyError(source, fmt.Errorf("back up database: %w", err))
	}
	if err := os.Chmod(snapshot, 0o600); err != nil {
		return "", err
	}
	file, err := os.OpenFile(snapshot, os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	err = file.Sync()
	closeErr := file.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := publishBackup(ctx, snapshot, target, os.Link); err != nil {
		return "", fmt.Errorf("publish backup without overwriting destination: %w", err)
	}
	return target, nil
}

func publishBackup(ctx context.Context, snapshot, target string, link func(string, string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := link(snapshot, target)
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrExist) {
		return err
	}
	// Unsupported links have filesystem- and OS-specific errors. O_EXCL is safe
	// for any other link failure; genuine destination errors fail again on open.
	input, err := os.Open(snapshot)
	if err != nil {
		return err
	}
	defer input.Close()
	return copyBackupExclusive(ctx, target, input)
}

func copyBackupExclusive(ctx context.Context, target string, input io.Reader) (err error) {
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	owned, err := output.Stat()
	if err != nil {
		return errors.Join(err, output.Close())
	}
	defer func() {
		err = errors.Join(err, output.Close())
		current, statErr := os.Lstat(target)
		if statErr != nil || !os.SameFile(owned, current) {
			if err == nil {
				err = errors.New("backup destination changed during publication")
			}
			return
		}
		if err != nil {
			// Never remove a replacement file or a symlink installed by another writer.
			err = errors.Join(err, os.Remove(target))
		}
	}()
	if _, err := io.Copy(output, backupContextReader{ctx: ctx, input: input}); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return output.Sync()
}

type backupContextReader struct {
	ctx   context.Context
	input io.Reader
}

func (r backupContextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.input.Read(buffer)
}
