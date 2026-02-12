package cli

import (
	"fmt"
	"time"

	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [workitem name]",
	Short: "Create a new work item",
	Args:  cobra.ExactArgs(1),
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

		svc := service.NewWorkItemService(database)
		created, parent, err := svc.CreateFromInput(personID, args[0], time.Now())
		if err != nil {
			return err
		}

		ui.PrintSuccess(fmt.Sprintf("Created work item: %s", created.Name))
		if parent != nil {
			fmt.Printf("  Under: %s\n", parent.Name)
		}
		return nil
	},
}

