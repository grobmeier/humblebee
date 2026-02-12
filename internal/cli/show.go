package cli

import (
	"fmt"

	"github.com/grobmeier/humblebee/internal/repo"
	"github.com/grobmeier/humblebee/internal/service"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "List all work items",
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

		itemsRepo := repo.NewWorkItemRepo(database)
		items, err := itemsRepo.ListActive(personID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("No work items found.")
			fmt.Println(`Use 'humblebee add "Work Item Name"' to create one.`)
			return nil
		}

		fmt.Println("Work Items:")
		roots := service.BuildTree(items)
		for _, r := range roots {
			printNode(r, 0)
		}
		return nil
	},
}

func printNode(n *service.TreeNode, indent int) {
	prefix := ""
	if indent > 0 {
		for i := 0; i < indent-1; i++ {
			prefix += "  "
		}
		prefix += "└─ "
	}
	fmt.Printf("  %s[%d] %s\n", prefix, n.Item.ID, n.Item.Name)
	for _, c := range n.Children {
		printNode(c, indent+1)
	}
}

