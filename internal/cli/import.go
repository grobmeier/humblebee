package cli

import (
	"fmt"

	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [timeandbill-export.json]",
	Short: "Import a Time & Bill export",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		assumeYes, _ := cmd.Flags().GetBool("yes")
		updateExisting, _ := cmd.Flags().GetBool("update-existing")

		database, _, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := requireInitialized(database); err != nil {
			return err
		}
		personID, err := defaultPersonID(database)
		if err != nil {
			return err
		}

		importer := service.NewTimeAndBillImportService(database)
		importer.SetConfirm(func(message string) (bool, error) {
			return ui.Confirm(message, true)
		})

		options := service.TimeAndBillImportOptions{
			DryRun:         dryRun,
			AssumeYes:      assumeYes,
			UpdateExisting: updateExisting,
		}

		if !dryRun && !assumeYes {
			ok, err := ui.Confirm("Import this Time & Bill export into HumbleBee?", true)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintError("Import cancelled.")
				return nil
			}
		}

		summary, err := importer.ImportFile(personID, args[0], options)
		if err != nil {
			return err
		}

		printImportSummary(summary, dryRun)
		if summary.AlreadyImported && !updateExisting {
			ui.PrintError("This Time & Bill export was already imported. Use --update-existing to re-apply matching time-entry updates.")
			return nil
		}
		if dryRun {
			ui.PrintSuccess("Dry run completed.")
		} else {
			ui.PrintSuccess("Import completed.")
		}
		return nil
	},
}

func init() {
	importCmd.Flags().Bool("dry-run", false, "Preview the import without writing data")
	importCmd.Flags().Bool("yes", false, "Skip the final import confirmation")
	importCmd.Flags().Bool("update-existing", false, "Update matching imported time entries")
}

func printImportSummary(summary service.TimeAndBillImportSummary, dryRun bool) {
	mode := "Import"
	if dryRun {
		mode = "Dry run"
	}
	fmt.Printf("%s summary for export %s\n", mode, summary.ExportUUID)
	fmt.Printf("  Projects: created %d, mapped %d, skipped %d\n", summary.ProjectsCreated, summary.ProjectsMapped, summary.ProjectsSkipped)
	fmt.Printf("  Tasks: created %d, mapped %d, skipped %d\n", summary.TasksCreated, summary.TasksMapped, summary.TasksSkipped)
	fmt.Printf("  Time entries: created %d, updated %d, skipped %d\n", summary.TimeEntriesCreated, summary.TimeEntriesUpdated, summary.TimeEntriesSkipped)
	if summary.NeedsConfirmation > 0 {
		fmt.Printf("  Needs confirmation: %d project name match(es)\n", summary.NeedsConfirmation)
	}
}
