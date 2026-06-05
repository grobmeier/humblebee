package cli

import (
	"errors"
	"os"

	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "humblebee",
	Short: "HumbleBee - CLI time tracking that stays out of your way",
}

func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		noColor, _ := cmd.Flags().GetBool("no-color")
		if noColor || os.Getenv("NO_COLOR") != "" {
			ui.DisableColor()
		}
	}

	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")

	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(guiCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		ui.PrintError(cliErrorMessage(err))
		return err
	}
	return nil
}

func cliErrorMessage(err error) string {
	var busy *db.BusyError
	if errors.As(err, &busy) {
		return db.BusyRecoveryMessage(busy.Path)
	}
	return err.Error()
}
