package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}

			task, err := st.GetTask(id)
			if err != nil {
				return err
			}

			// Check if task has recurrence
			rec, _ := st.GetRecurrence(id)
			if rec != nil {
				// Complete with recurrence handling
				if err := st.CompleteTaskWithRecurrence(id); err != nil {
					return err
				}
				fmt.Printf("Completed task #%d: %s (next occurrence created)\n", task.ID, task.Title)
			} else {
				task.MarkDone()
				if err := st.UpdateTask(task); err != nil {
					return err
				}
				fmt.Printf("Completed task #%d: %s\n", task.ID, task.Title)
			}

			return nil
		},
	}

	return cmd
}

func newDoingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doing <id>",
		Short: "Mark a task as in progress",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task id: %s", args[0])
			}

			task, err := st.GetTask(id)
			if err != nil {
				return err
			}

			task.MarkDoing()
			if err := st.UpdateTask(task); err != nil {
				return err
			}

			fmt.Printf("Started task #%d: %s\n", task.ID, task.Title)
			return nil
		},
	}

	return cmd
}
