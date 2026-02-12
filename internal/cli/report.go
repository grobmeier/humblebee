package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/grobmeier/humblebee/internal/duration"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report [month] [year]",
	Short: "Show a monthly time report grouped by work item",
	Args:  cobra.RangeArgs(0, 2),
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
		month := int(now.Month())
		year := now.Year()

		if len(args) >= 1 {
			m, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("month must be an integer between 1 and 12")
			}
			if m < 1 || m > 12 {
				return fmt.Errorf("invalid month: %d", m)
			}
			month = m
		}
		if len(args) == 2 {
			y, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.New("year must be a 4-digit integer")
			}
			if y < 1900 || y > 2100 {
				return fmt.Errorf("invalid year: %d", y)
			}
			year = y
		}

		svc := service.NewReportService(database)
		rep, err := svc.Monthly(personID, year, time.Month(month), now, time.Local)
		if err != nil {
			return err
		}

		if rep.TotalSec == 0 {
			fmt.Printf("No time entries found for %s %d\n", time.Month(month).String(), year)
			fmt.Println("Use 'humblebee start' to begin tracking time")
			return nil
		}

		fmt.Println(rep.Title)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Work Item", "Duration", ""})
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetAutoWrapText(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT})

		for _, line := range rep.Lines {
			name := line.Name
			if line.Depth > 0 {
				prefix := ""
				for i := 0; i < line.Depth-1; i++ {
					prefix += "  "
				}
				prefix += "  "
				name = prefix + name
			}
			percent := ""
			if line.Percent != nil {
				percent = fmt.Sprintf("(%d%%)", *line.Percent)
			}
			table.Append([]string{name, duration.FormatSeconds(line.Seconds), percent})
		}
		table.Render()

		fmt.Printf("TOTAL  %s\n", duration.FormatSeconds(rep.TotalSec))
		fmt.Printf("\nWorking days: %d\n", rep.WorkingDays)
		fmt.Printf("Average per day: %s\n", duration.FormatSeconds(rep.AvgPerDay))
		return nil
	},
}

