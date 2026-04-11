package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/grobmeier/humblebee/internal/duration"
	"github.com/grobmeier/humblebee/internal/model"
	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/timeutil"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Short:   "Delete data (time entries)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior for now: delete time entries.
		return deleteTimeInteractive(cmd)
	},
}

var deleteTimeCmd = &cobra.Command{
	Use:   "time",
	Short: "Interactively delete time entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteTimeInteractive(cmd)
	},
}

func init() {
	deleteCmd.AddCommand(deleteTimeCmd)
}

type workItemChoice struct {
	label string
	id    *int64 // nil means Default entries (workitem_id IS NULL)
}

type entryChoice struct {
	label string
	id    int64
}

func deleteTimeInteractive(cmd *cobra.Command) error {
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

	itemsRepo := repo.NewWorkItemRepo(database)
	workItems, err := itemsRepo.ListActive(personID)
	if err != nil {
		return err
	}

	selected, err := chooseWorkItem(workItems)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil
	}

	loc := time.Local
	currentDay := time.Now().In(loc)
	// normalize to midnight
	currentDay = time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 0, 0, 0, 0, loc)

	entriesRepo := repo.NewTimeEntryRepo(database)

	for {
		dayStart := currentDay
		dayEnd := currentDay.AddDate(0, 0, 1)

		entries, err := entriesRepo.ListOverlappingForWorkItem(personID, selected.id, dayStart.UTC().Unix(), dayEnd.UTC().Unix())
		if err != nil {
			return err
		}

		header := fmt.Sprintf("%s — %s", strings.TrimSpace(selected.label), currentDay.Format("2006-01-02"))
		fmt.Println(header)

		type action string
		const (
			actPrev   action = "← Previous day"
			actNext   action = "→ Next day"
			actProj   action = "Change work item"
			actExit   action = "Exit"
			actPrefix action = ""
		)

		var options []string
		options = append(options, string(actPrev))
		today := time.Now().In(loc)
		todayMid := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
		if currentDay.Before(todayMid) {
			options = append(options, string(actNext))
		}
		options = append(options, string(actProj))
		options = append(options, string(actExit))

		var entryOptions []entryChoice
		for _, e := range entries {
			if e.EndTime == nil {
				continue
			}
			overlap := timeutil.OverlapSeconds(e.StartTime, *e.EndTime, timeutil.Window{Start: dayStart, End: dayEnd})
			if overlap <= 0 {
				continue
			}
			startLocal := time.Unix(e.StartTime, 0).In(loc).Format("15:04")
			endLocal := time.Unix(*e.EndTime, 0).In(loc).Format("15:04")
			desc := ""
			if e.Description != nil && strings.TrimSpace(*e.Description) != "" {
				desc = " — " + strings.TrimSpace(*e.Description)
			}
			label := fmt.Sprintf("[#%d] %s-%s (%s)%s", e.ID, startLocal, endLocal, duration.FormatSeconds(overlap), desc)
			entryOptions = append(entryOptions, entryChoice{label: label, id: e.ID})
		}
		if len(entryOptions) == 0 {
			options = append(options, "No entries for this day")
		} else {
			options = append(options, "Select an entry to delete:")
			for _, eo := range entryOptions {
				options = append(options, "  "+eo.label)
			}
		}

		var picked string
		if err := survey.AskOne(&survey.Select{
			Message: "Choose an action:",
			Options: options,
		}, &picked); err != nil {
			if errors.Is(err, terminal.InterruptErr) {
				return nil
			}
			return err
		}

		switch action(strings.TrimSpace(picked)) {
		case actPrev:
			currentDay = currentDay.AddDate(0, 0, -1)
			continue
		case actNext:
			currentDay = currentDay.AddDate(0, 0, 1)
			continue
		case actProj:
			next, err := chooseWorkItem(workItems)
			if err != nil {
				return err
			}
			if next == nil {
				return nil
			}
			selected = next
			continue
		case actExit:
			return nil
		default:
			// entry selection lines are prefixed with two spaces
			picked = strings.TrimSpace(picked)
			var entryID int64 = -1
			for _, eo := range entryOptions {
				if eo.label == picked {
					entryID = eo.id
					break
				}
			}
			if entryID < 0 {
				// no-op (e.g., "No entries..." informational rows)
				continue
			}

			ok, err := ui.Confirm(fmt.Sprintf("Delete time entry #%d? This cannot be undone.", entryID), true)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if err := entriesRepo.DeleteByID(personID, entryID); err != nil {
				return err
			}
			ui.PrintSuccess(fmt.Sprintf("Deleted time entry #%d", entryID))
			continue
		}
	}
}

func chooseWorkItem(workItems []model.WorkItem) (*workItemChoice, error) {
	choices := []workItemChoice{{label: "Default", id: nil}}
	tree := service.BuildTree(workItems)
	var flatten func(n *service.TreeNode, depth int)
	flatten = func(n *service.TreeNode, depth int) {
		// Skip the "Default" row in the workitems table; Default entries are represented by NULL workitem_id.
		if !(strings.EqualFold(n.Item.Name, "Default") && n.Item.ParentID == nil) {
			id := n.Item.ID
			prefix := ""
			if depth > 0 {
				prefix = strings.Repeat("  ", depth) + "└─ "
			}
			choices = append(choices, workItemChoice{label: prefix + n.Item.Name, id: &id})
		}
		for _, c := range n.Children {
			flatten(c, depth+1)
		}
	}
	for _, r := range tree {
		flatten(r, 0)
	}

	labels := make([]string, 0, len(choices))
	for _, c := range choices {
		labels = append(labels, c.label)
	}

	var chosenLabel string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a work item:",
		Options: labels,
	}, &chosenLabel); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			return nil, nil
		}
		return nil, err
	}

	for i := range choices {
		if choices[i].label == chosenLabel {
			return &choices[i], nil
		}
	}
	return nil, errors.New("no work item selected")
}
