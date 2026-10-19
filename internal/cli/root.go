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

package cli

import (
	"errors"
	"os"

	"github.com/grobmeier/humblebee/internal/connected"
	"github.com/grobmeier/humblebee/internal/db"
	"github.com/grobmeier/humblebee/internal/remotecli"
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
	rootCmd.AddCommand(remotecli.NewCommand(remotecli.Options{}))

	command, err := rootCmd.ExecuteC()
	if err != nil {
		if remotecli.IsRemote(command) {
			return remotecli.ReportError(command, err)
		}
		ui.PrintError(cliErrorMessage(err))
		return err
	}
	return nil
}

// ExitCode keeps legacy CLI failures at 1 and preserves remote scripting codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var remote *connected.Error
	if errors.As(err, &remote) {
		return remote.Exit
	}
	return 1
}

func cliErrorMessage(err error) string {
	var busy *db.BusyError
	if errors.As(err, &busy) {
		return db.BusyRecoveryMessage(busy.Path)
	}
	return err.Error()
}
