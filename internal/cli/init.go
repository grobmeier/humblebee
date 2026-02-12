package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/grobmeier/humblebee/internal/validator"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize HumbleBee on this machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, dbPath, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		initialized, err := db.IsInitialized(database)
		if err != nil {
			return err
		}
		if initialized {
			people := repo.NewPersonRepo(database)
			p, _ := people.GetDefault()
			ui.PrintWarning("HumbleBee is already initialized")
			fmt.Printf("  Database: %s\n", dbPath)
			if p != nil {
				fmt.Printf("  User: %s\n", p.Email)
			}
			fmt.Printf("  Use 'humblebee show' to list work items\n")
			return nil
		}

		if err := db.Migrate(database); err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		email = strings.TrimSpace(email)
		if email == "" {
			v := func(s string) error { return validator.ValidateEmail(s) }
			email, err = ui.PromptString("Enter your email address:", v)
			if err != nil {
				return err
			}
		} else {
			if err := validator.ValidateEmail(email); err != nil {
				return err
			}
		}

		initialItem, _ := cmd.Flags().GetString("workitem")
		initialItem = strings.TrimSpace(initialItem)
		if initialItem == "" {
			initialItem, _ = ui.PromptString("Create an initial work item? (optional, press Enter to skip):", nil)
		}

		svc := service.NewInitService(database)
		person, created, err := svc.Init(service.InitParams{
			Email:           email,
			InitialWorkItem: initialItem,
			Now:             time.Now(),
		})
		if err != nil {
			return err
		}

		ui.PrintSuccess("HumbleBee initialized!")
		fmt.Printf("  Database: %s\n", dbPath)
		fmt.Printf("  User: %s\n", person.Email)
		fmt.Printf("  Work items created:\n")
		for _, it := range created {
			fmt.Printf("    - %s\n", it.Name)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().String("email", "", "Email to initialize with (non-interactive)")
	initCmd.Flags().String("workitem", "", "Initial work item name to create (non-interactive)")
}

