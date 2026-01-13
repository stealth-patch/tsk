package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "rm <id>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a task",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}

			task, err := st.GetTask(id)
			if err != nil {
				return err
			}

			if !force {
				fmt.Printf("Delete task #%d: %s? [y/N] ", task.ID, task.Title)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			if err := st.DeleteTask(id); err != nil {
				return err
			}

			fmt.Printf("Deleted task #%d: %s\n", task.ID, task.Title)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}
