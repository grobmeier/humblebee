package cli

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/timeutil"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check HumbleBee installation and data health",
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")
		tzNameOverride, _ := cmd.Flags().GetString("tz-name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		database, dbPath, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		fmt.Println("HumbleBee Doctor")
		fmt.Printf("Database: %s\n", dbPath)
		if home := os.Getenv("HUMBLEBEE_HOME"); home != "" {
			fmt.Printf("HUMBLEBEE_HOME: %s\n", home)
		}

		initialized, err := db.IsInitialized(database)
		if err != nil {
			return err
		}
		if !initialized {
			ui.PrintWarning("Not initialized")
			fmt.Println("Run: humblebee init")
			return nil
		}
		ui.PrintSuccess("Initialized")

		schemaVersion, err := getSchemaVersion(database)
		if err != nil {
			return err
		}
		fmt.Printf("Schema version: %s\n", schemaVersion)

		hasTZCols, err := hasTimeEntryTZColumns(database)
		if err != nil {
			return err
		}
		if hasTZCols {
			ui.PrintSuccess("time_entries timezone columns present")
		} else {
			ui.PrintWarning("time_entries timezone columns missing")
		}

		personID, err := defaultPersonID(database)
		if err != nil {
			return err
		}
		fmt.Printf("Default person id: %d\n", personID)

		entriesRepo := repo.NewTimeEntryRepo(database)
		running, err := entriesRepo.FindRunning(personID)
		if err != nil {
			return err
		}
		if running != nil {
			ui.PrintWarning("Timer is running (stop it before deleting/backfilling)")
		} else {
			ui.PrintSuccess("No running timer")
		}

		missing, err := countMissingTZ(database, personID)
		if err != nil {
			return err
		}
		fmt.Printf("Entries missing timezone info: %d\n", missing)

		if missing > 0 && !fix {
			ui.PrintWarning("Some entries predate timezone tracking")
			fmt.Println("Run: humblebee doctor --fix")
		}

		if !fix {
			return nil
		}
		if dryRun {
			ui.PrintWarning("--dry-run enabled; no changes will be made")
		}
		if running != nil {
			return errors.New("stop the running timer before running doctor --fix")
		}

		// Migrate again (idempotent) to ensure columns exist.
		if !dryRun {
			if err := db.Migrate(database); err != nil {
				return err
			}
		}

		loc := time.Local
		if tzNameOverride != "" {
			override, err := time.LoadLocation(tzNameOverride)
			if err != nil {
				return fmt.Errorf("invalid --tz-name: %w", err)
			}
			loc = override
		}

		if missing == 0 {
			ui.PrintSuccess("Nothing to fix")
			return nil
		}

		if !dryRun {
			n, err := backfillTZ(database, personID, loc)
			if err != nil {
				return err
			}
			ui.PrintSuccess(fmt.Sprintf("Backfilled timezone info for %d entries", n))
		} else {
			fmt.Printf("Would backfill timezone info for %d entries\n", missing)
		}

		return nil
	},
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Attempt safe fixes (backfill timezone info for older entries)")
	doctorCmd.Flags().String("tz-name", "", "Timezone name to use for backfill (IANA, e.g. America/New_York)")
	doctorCmd.Flags().Bool("dry-run", false, "Show what would change without modifying the database")
}

func getSchemaVersion(dbConn *sql.DB) (string, error) {
	var v string
	if err := dbConn.QueryRow(`SELECT value FROM config WHERE key='schema_version'`).Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

func hasTimeEntryTZColumns(dbConn *sql.DB) (bool, error) {
	rows, err := dbConn.Query(`PRAGMA table_info(time_entries);`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return cols["tz_name"] && cols["tz_offset_minutes"], nil
}

func countMissingTZ(dbConn *sql.DB, personID int64) (int64, error) {
	var n int64
	if err := dbConn.QueryRow(`
		SELECT COUNT(*)
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND tz_name = ''
		  AND tz_offset_minutes = 0
	`, personID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func backfillTZ(dbConn *sql.DB, personID int64, loc *time.Location) (int64, error) {
	rows, err := dbConn.Query(`
		SELECT id, start_time, end_time
		FROM time_entries
		WHERE person_id = ?
		  AND end_time IS NOT NULL
		  AND tz_name = ''
		  AND tz_offset_minutes = 0
	`, personID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type row struct {
		id        int64
		startTime int64
		endTime   int64
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.startTime, &r.endTime); err != nil {
			return 0, err
		}
		pending = append(pending, r)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var updated int64
	for _, r := range pending {
		t := time.Unix(r.startTime, 0).In(loc)
		_, offsetSec := t.Zone()
		offsetMin := offsetSec / 60
		tzName := loc.String()
		// If the OS returns "Local", ensure we still have a useful location later by keeping offset.
		fallbackLoc := timeutil.LocationForEntry(tzName, offsetMin, time.Local)
		tzName = fallbackLoc.String()

		if _, err := tx.Exec(`
			UPDATE time_entries
			SET tz_name = ?, tz_offset_minutes = ?, updated_at = strftime('%s','now')
			WHERE id = ? AND person_id = ?
		`, tzName, offsetMin, r.id, personID); err != nil {
			return 0, err
		}
		updated++
	}

	return updated, tx.Commit()
}

