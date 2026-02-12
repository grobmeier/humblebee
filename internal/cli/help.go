package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Show available commands",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := cmd.Root()
		if len(args) == 1 {
			target, _, err := root.Find([]string{args[0]})
			if err != nil || target == nil || target == root {
				return fmt.Errorf("unknown help topic: %s", args[0])
			}
			return target.Help()
		}

		type row struct {
			use   string
			short string
		}
		var rows []row
		for _, c := range root.Commands() {
			if c.Hidden || c.Name() == "help" {
				continue
			}
			if c.Short == "" {
				continue
			}
			rows = append(rows, row{use: c.Name(), short: c.Short})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].use < rows[j].use })

		fmt.Println("Available commands:")
		for _, r := range rows {
			fmt.Printf("  %-10s %s\n", r.use, strings.TrimSpace(r.short))
		}
		fmt.Println()
		fmt.Printf("Version: %s (%s)\n", buildInfo.Version, buildInfo.Commit)
		fmt.Println("Use 'humblebee help [command]' or 'humblebee [command] --help' for more information.")
		return nil
	},
}
