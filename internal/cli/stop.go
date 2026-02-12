package cli

import (
	"fmt"
	"time"

	"github.com/grobmeier/humblebee/internal/duration"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the currently running timer",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		now := time.Now()
		loc := time.Local

		timer := service.NewTimerService(database)
		res, err := timer.Stop(personID, now, loc)
		if err != nil {
			ui.PrintWarning(err.Error())
			fmt.Println("  Use 'humblebee start [work item]' to start tracking time")
			return nil
		}

		name := "Default"
		if res.StoppedEntry.WorkItemID != nil {
			itemsRepo := repo.NewWorkItemRepo(database)
			item, _ := itemsRepo.GetByID(personID, *res.StoppedEntry.WorkItemID)
			if item != nil {
				name = item.Name
			}
		}

		ui.PrintSuccess(fmt.Sprintf("Timer stopped: %s", name))
		fmt.Printf("  Duration: %s\n", duration.FormatSeconds(res.DurationSec))
		fmt.Printf("  Total today: %s\n", duration.FormatSeconds(res.TodayTotal))
		return nil
	},
}

