package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newRecurrenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recurrence",
		Aliases: []string{"rec", "repeat"},
		Short:   "Manage task recurrence",
	}

	cmd.AddCommand(newRecurrenceRmCmd())

	return cmd
}

func newRecurrenceRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <task-id>",
		Aliases: []string{"remove", "clear"},
		Short:   "Remove recurrence from a task",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}

			// Check task exists
			task, err := st.GetTask(taskID)
			if err != nil {
				return err
			}

			// Check recurrence exists
			rec, err := st.GetRecurrence(taskID)
			if err != nil {
				return err
			}
			if rec == nil {
				return fmt.Errorf("task #%d has no recurrence", taskID)
			}

			if err := st.DeleteRecurrence(taskID); err != nil {
				return err
			}

			fmt.Printf("Removed recurrence from task #%d: %s\n", task.ID, task.Title)
			return nil
		},
	}
}
