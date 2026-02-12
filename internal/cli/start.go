package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [workitem name]",
	Short: "Start a timer",
	Args:  cobra.RangeArgs(0, 1),
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
		running, err := timer.Running(personID)
		if err != nil {
			return err
		}

		var targetName = "Default"
		var workItemID *int64 = nil

		if len(args) == 1 {
			input := strings.TrimSpace(args[0])
			if input != "" && !strings.EqualFold(input, "Default") {
				workSvc := service.NewWorkItemService(database)
				res, err := workSvc.ResolveByInput(personID, input)
				if err != nil && res.Item == nil && len(res.Candidates) > 0 {
					ui.PrintError(err.Error())
					fmt.Println("Matches:")
					for _, c := range res.Candidates {
						if c.Path != nil {
							fmt.Printf("  - %s (%s)\n", c.Name, *c.Path)
						} else {
							fmt.Printf("  - %s\n", c.Name)
						}
					}
					return nil
				}
				if err != nil {
					return err
				}
				targetName = res.Item.Name
				// Special-case: user might explicitly target the "Default" work item.
				if strings.EqualFold(targetName, "Default") && res.Item.ParentID == nil {
					workItemID = nil
				} else {
					workItemID = &res.Item.ID
				}
			}
		}

		if running != nil {
			currentName := "Default"
			if running.WorkItemID != nil {
				itemsRepo := repo.NewWorkItemRepo(database)
				item, _ := itemsRepo.GetByID(personID, *running.WorkItemID)
				if item != nil {
					currentName = item.Name
				}
			}
			elapsed := now.UTC().Unix() - running.StartTime
			ui.PrintWarning(fmt.Sprintf("Timer already running: %s (started %s ago)", currentName, serviceFormatElapsed(elapsed)))
			ok, err := ui.Confirm("Stop current timer and start new one?", true)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			stopped, err := timer.Stop(personID, now, loc)
			if err != nil {
				return err
			}
			ui.PrintSuccess(fmt.Sprintf("Stopped: %s (%s)", currentName, serviceFormatElapsed(stopped.DurationSec)))
		}

		if _, err := timer.Start(service.StartParams{
			PersonID:   personID,
			WorkItemID: workItemID,
			Now:        now,
		}); err != nil {
			return err
		}

		fmt.Printf("⏱  Timer started: %s\n", targetName)
		fmt.Printf("   Started at: %s\n", now.In(loc).Format("15:04"))
		return nil
	},
}

func serviceFormatElapsed(seconds int64) string {
	// Use a simple humanization compatible with the spec examples.
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
}

