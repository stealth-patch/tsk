package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hwanchang/tsk/internal/model"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagAddCmd())
	cmd.AddCommand(newTagRmCmd())

	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := st.ListTags()
			if err != nil {
				return err
			}

			if len(tags) == 0 {
				fmt.Println("No tags found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tCOLOR")

			for _, t := range tags {
				fmt.Fprintf(w, "%d\t%s\t%s\n", t.ID, t.Name, t.Color)
			}

			return w.Flush()
		},
	}
}

func newTagAddCmd() *cobra.Command {
	var color string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Check if tag already exists
			existing, err := st.GetTagByName(name)
			if err != nil {
				return err
			}
			if existing != nil {
				return fmt.Errorf("tag already exists: %s", name)
			}

			tag := model.NewTag(name)
			if color != "" {
				tag.Color = color
			}

			if err := st.CreateTag(tag); err != nil {
				return err
			}

			fmt.Printf("Created tag #%d: %s\n", tag.ID, tag.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&color, "color", "c", "", "tag color (hex, e.g., #FF0000)")

	return cmd
}

func newTagRmCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"remove", "delete"},
		Short:   "Delete a tag",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Find the tag
			tag, err := st.GetTagByName(name)
			if err != nil {
				return err
			}
			if tag == nil {
				return fmt.Errorf("tag not found: %s", name)
			}

			// Confirm deletion
			if !force {
				fmt.Printf("Delete tag '%s'? This will remove it from all tasks. [y/N] ", name)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			if err := st.DeleteTag(tag.ID); err != nil {
				return err
			}

			fmt.Printf("Deleted tag: %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}
