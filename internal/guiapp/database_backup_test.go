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
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func backupFixture(t *testing.T) (*App, *sql.DB, string) {
	t.Helper()
	t.Setenv("HUMBLEBEE_HOME", t.TempDir())
	a := New()
	if err := a.Init("backup@example.test"); err != nil {
		t.Fatal(err)
	}
	database, source, err := a.openDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return a, database, source
}

func TestBackupWALSnapshotAndNormalSwitch(t *testing.T) {
	a, database, source := backupFixture(t)
	for _, query := range []string{
		"PRAGMA journal_mode=WAL", "PRAGMA wal_autocheckpoint=0",
		"CREATE TABLE backup_probe (id INTEGER PRIMARY KEY, value TEXT)",
		"INSERT INTO backup_probe VALUES (1, 'committed in WAL')",
	} {
		if _, err := database.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	walBefore, err := os.ReadFile(source + "-wal")
	if err != nil || len(walBefore) == 0 {
		t.Fatalf("WAL fixture: %v", err)
	}
	mainBefore, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "snapshot.db")
	got, err := a.backupDatabase(source, func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		if !strings.HasPrefix(options.DefaultFilename, "humblebee-") || !strings.HasSuffix(options.DefaultFilename, ".db") {
			t.Fatalf("filename: %q", options.DefaultFilename)
		}
		return target, nil
	})
	if err != nil || got != target {
		t.Fatalf("backup: %q, %v", got, err)
	}
	mainAfter, _ := os.ReadFile(source)
	walAfter, _ := os.ReadFile(source + "-wal")
	if !bytes.Equal(mainBefore, mainAfter) || !bytes.Equal(walBefore, walAfter) {
		t.Fatal("backup changed source main/WAL bytes")
	}
	if selected, _ := a.databasePath(); selected != source {
		t.Fatal("backup switched database")
	}
	if _, err := database.Exec("INSERT INTO backup_probe VALUES (2, 'after snapshot')"); err != nil {
		t.Fatal(err)
	}
	copyDB, err := db.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer copyDB.Close()
	var integrity, value string
	var count int
	if err := copyDB.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("integrity: %s %v", integrity, err)
	}
	if err := copyDB.QueryRow("SELECT value FROM backup_probe WHERE id=1").Scan(&value); err != nil || value != "committed in WAL" {
		t.Fatalf("WAL value: %q %v", value, err)
	}
	if err := copyDB.QueryRow("SELECT count(*) FROM backup_probe").Scan(&count); err != nil || count != 1 {
		t.Fatalf("snapshot count: %d %v", count, err)
	}
	if info, err := a.SwitchDatabase(target); err != nil || !info.Initialized || info.Path != target {
		t.Fatalf("switch snapshot: %#v %v", info, err)
	}
	assertNoBackupStaging(t, filepath.Dir(target))
}

func TestBackupRejectsExistingTargetsAndSourceAliases(t *testing.T) {
	_, _, source := backupFixture(t)
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.db")
	empty := filepath.Join(dir, "empty.db")
	alias := filepath.Join(dir, "alias.db")
	hardlink := filepath.Join(dir, "hardlink.db")
	dangling := filepath.Join(dir, "dangling.db")
	if err := os.WriteFile(existing, []byte("do not overwrite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(source, hardlink); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "missing"), dangling); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(source)
	for _, target := range []string{source, alias, hardlink, existing, empty, dangling, dir} {
		if path, err := snapshotDatabase(context.Background(), source, target); err == nil || path != "" {
			t.Fatalf("accepted %q: %q %v", target, path, err)
		}
	}
	after, _ := os.ReadFile(source)
	if !bytes.Equal(before, after) {
		t.Fatal("source changed")
	}
	if data, _ := os.ReadFile(existing); string(data) != "do not overwrite" {
		t.Fatal("existing file changed")
	}
	if info, err := os.Stat(empty); err != nil || info.Size() != 0 {
		t.Fatal("empty target changed")
	}
	if _, err := os.Lstat(dangling); err != nil {
		t.Fatal("removed unrelated symlink")
	}
	assertNoBackupStaging(t, dir)
}

func TestBackupCancellationAndFailures(t *testing.T) {
	a, _, source := backupFixture(t)
	for _, dialogErr := range []error{nil, errors.New("dialog failed")} {
		path, err := a.backupDatabase(source, func(context.Context, runtime.SaveDialogOptions) (string, error) { return "", dialogErr })
		if path != "" || !errors.Is(err, dialogErr) {
			t.Fatalf("cancel/error: %q %v", path, err)
		}
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "backup.db")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := snapshotDatabase(ctx, source, target); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled snapshot: %v", err)
	}
	corrupt := filepath.Join(dir, "corrupt.db")
	if err := os.WriteFile(corrupt, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{corrupt, filepath.Join(dir, "missing.db"), dir} {
		if _, err := snapshotDatabase(context.Background(), input, target); err == nil {
			t.Fatalf("accepted source %s", input)
		}
	}
	if _, err := snapshotDatabase(context.Background(), source, filepath.Join(dir, "missing", "backup.db")); err == nil {
		t.Fatal("accepted unavailable destination directory")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("partial target remains: %v", err)
	}
	assertNoBackupStaging(t, dir)
}

func TestBackupBusySource(t *testing.T) {
	_, database, source := backupFixture(t)
	conn, err := database.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(context.Background(), "BEGIN EXCLUSIVE"); err != nil {
		t.Fatal(err)
	}
	defer conn.ExecContext(context.Background(), "ROLLBACK")
	dir := t.TempDir()
	if _, err := snapshotDatabase(context.Background(), source, filepath.Join(dir, "busy.db")); err == nil || !db.IsBusyError(err) {
		t.Fatalf("expected busy error: %v", err)
	}
	assertNoBackupStaging(t, dir)
}

func TestBackupRejectsStaleContextBeforeDialog(t *testing.T) {
	a, _, source := backupFixture(t)
	for _, expected := range []string{"", filepath.Join(t.TempDir(), "stale.db")} {
		_, err := a.backupDatabase(expected, func(context.Context, runtime.SaveDialogOptions) (string, error) {
			t.Fatal("dialog opened for stale context")
			return "", nil
		})
		if err == nil {
			t.Fatal("accepted stale context")
		}
	}
	if _, err := a.CreateDatabase(filepath.Join(t.TempDir(), "other.db")); err != nil {
		t.Fatal(err)
	}
	if _, err := a.BackupDatabase(source); err == nil {
		t.Fatal("accepted old database after switch")
	}
}

func TestBackupConcurrentPublicationNeverOverwrites(t *testing.T) {
	_, _, source := backupFixture(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "shared.db")
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := snapshotDatabase(context.Background(), source, target)
			results <- err
		}()
	}
	close(start)
	successes := 0
	for range 2 {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one publisher, got %d", successes)
	}
	if info, err := os.Stat(target); err != nil || info.Size() == 0 {
		t.Fatalf("published backup missing: %v", err)
	}
	assertNoBackupStaging(t, dir)
}

func TestBackupEscapesSQLiteSourceURI(t *testing.T) {
	_, _, source := backupFixture(t)
	dir := t.TempDir()
	escaped := filepath.Join(dir, "source # ? %.db")
	if _, err := snapshotDatabase(context.Background(), source, escaped); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshotDatabase(context.Background(), escaped, filepath.Join(dir, "result.db")); err != nil {
		t.Fatal(err)
	}
}

func TestBackupSerializesDatabaseSwitches(t *testing.T) {
	for _, action := range []string{"switch", "create", "default"} {
		t.Run(action, func(t *testing.T) {
			a, _, source := backupFixture(t)
			dir := t.TempDir()
			other := filepath.Join(dir, "other.db")
			if err := os.WriteFile(other, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			backupDone, switchDone := make(chan error, 1), make(chan error, 1)
			go func() {
				_, err := a.backupDatabase(source, func(context.Context, runtime.SaveDialogOptions) (string, error) {
					close(entered)
					<-release
					return filepath.Join(dir, "backup.db"), nil
				})
				backupDone <- err
			}()
			<-entered
			go func() {
				var err error
				switch action {
				case "switch":
					_, err = a.SwitchDatabase(other)
				case "create":
					_, err = a.CreateDatabase(filepath.Join(dir, "new.db"))
				case "default":
					_, err = a.UseDefaultDatabase()
				}
				switchDone <- err
			}()
			select {
			case err := <-switchDone:
				close(release)
				<-backupDone
				t.Fatalf("switch completed during backup: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			if selected, _ := a.databasePath(); selected != source {
				t.Fatal("selection changed during backup")
			}
			close(release)
			if err := <-backupDone; err != nil {
				t.Fatal(err)
			}
			if err := <-switchDone; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func assertNoBackupStaging(t *testing.T, dir string) {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(dir, ".humblebee-backup-*"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("staging leftovers: %v %v", entries, err)
	}
}

func TestBackupPublishFallback(t *testing.T) {
	_, _, source := backupFixture(t)
	dir := t.TempDir()
	snapshot := filepath.Join(dir, "snapshot.db")
	if _, err := snapshotDatabase(context.Background(), source, snapshot); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "portable.db")
	unsupported := func(string, string) error { return errors.ErrUnsupported }
	if err := publishBackup(context.Background(), snapshot, target, unsupported); err != nil {
		t.Fatal(err)
	}
	want, _ := os.ReadFile(snapshot)
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("copy mismatch: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions: %v", info.Mode())
	}
	inputInfo, err := os.Stat(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(info, inputInfo) {
		t.Fatal("fallback unexpectedly linked files")
	}
	if err := publishBackup(context.Background(), snapshot, target, unsupported); !errors.Is(err, os.ErrExist) {
		t.Fatalf("existing destination accepted: %v", err)
	}
	// Simulate another publisher winning between the failed link and O_EXCL.
	raceTarget := filepath.Join(dir, "raced.db")
	err = publishBackup(context.Background(), snapshot, raceTarget, func(string, string) error {
		if err := os.WriteFile(raceTarget, []byte("other publisher"), 0o600); err != nil {
			t.Fatal(err)
		}
		return errors.ErrUnsupported
	})
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("race accepted: %v", err)
	}
	if got, _ := os.ReadFile(raceTarget); string(got) != "other publisher" {
		t.Fatal("race winner changed")
	}
}

type backupTestReader func([]byte) (int, error)

func (read backupTestReader) Read(p []byte) (int, error) { return read(p) }

func TestBackupFallbackFailureCleanupOwnership(t *testing.T) {
	for _, replacement := range []string{"none", "file", "symlink"} {
		t.Run(replacement, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "backup.db")
			failure := errors.New("synthetic copy failure")
			reader := backupTestReader(func(p []byte) (int, error) {
				if replacement != "none" {
					// Keep the original inode alive so inode reuse cannot hide replacement.
					if err := os.Rename(target, filepath.Join(dir, "original")); err != nil {
						t.Fatal(err)
					}
					if replacement == "file" {
						if err := os.WriteFile(target, []byte("replacement"), 0o600); err != nil {
							t.Fatal(err)
						}
					} else {
						if err := os.Symlink(filepath.Join(dir, "original"), target); err != nil {
							t.Fatal(err)
						}
					}
				}
				return copy(p, "partial snapshot"), failure
			})
			if err := copyBackupExclusive(context.Background(), target, reader); !errors.Is(err, failure) {
				t.Fatalf("copy error: %v", err)
			}
			info, err := os.Lstat(target)
			if replacement == "none" {
				if !os.IsNotExist(err) {
					t.Fatalf("owned partial file remains: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("replacement removed: %v", err)
				}
				if replacement == "file" {
					if got, _ := os.ReadFile(target); string(got) != "replacement" {
						t.Fatal("replacement changed")
					}
				} else if info.Mode()&os.ModeSymlink == 0 {
					t.Fatal("replacement symlink changed")
				}
			}
		})
	}
}

func TestBackupFallbackCancellationCleansPartialFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	target := filepath.Join(t.TempDir(), "cancelled.db")
	reader := backupTestReader(func(p []byte) (int, error) {
		cancel()
		return copy(p, "partial"), io.EOF
	})
	if err := copyBackupExclusive(ctx, target, reader); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("partial file remains: %v", err)
	}
}
