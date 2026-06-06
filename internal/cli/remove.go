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
	"fmt"

	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/grobmeier/humblebee/internal/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [workitem name]",
	Short: "Archive a work item (and its subtree)",
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
		res, err := svc.ResolveByInput(personID, args[0])
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

		ok, err := ui.Confirm("This will archive the work item and all descendants; time entries will remain. Continue?", true)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		itemsRepo := repo.NewWorkItemRepo(database)
		if err := itemsRepo.ArchiveSubtree(personID, res.Item.ID); err != nil {
			return err
		}
		ui.PrintSuccess(fmt.Sprintf("Archived work item: %s", res.Item.Name))
		return nil
	},
}

