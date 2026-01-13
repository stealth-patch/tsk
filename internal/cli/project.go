package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hwanchang/tsk/internal/model"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj", "p"},
		Short:   "Manage projects",
	}

	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectAddCmd())
	cmd.AddCommand(newProjectRmCmd())

	return cmd
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := st.ListProjects()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tTASKS\tPROGRESS")

			for _, p := range projects {
				progress := ""
				if p.TaskCount > 0 {
					pct := float64(p.DoneCount) / float64(p.TaskCount) * 100
					progress = fmt.Sprintf("%d/%d (%.0f%%)", p.DoneCount, p.TaskCount, pct)
				} else {
					progress = "0/0"
				}
				fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", p.ID, p.Name, p.TaskCount, progress)
			}

			return w.Flush()
		},
	}
}

func newProjectAddCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new project",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.Join(args, " ")

			project := model.NewProject(name)
			project.Description = description

			if err := st.CreateProject(project); err != nil {
				return err
			}

			fmt.Printf("Created project #%d: %s\n", project.ID, project.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "project description")

	return cmd
}

func newProjectRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"remove", "delete"},
		Short:   "Delete a project (tasks will be moved to Inbox)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			projects, err := st.ListProjects()
			if err != nil {
				return err
			}

			var projectID int64
			for _, p := range projects {
				if strings.EqualFold(p.Name, name) {
					projectID = p.ID
					break
				}
			}

			if projectID == 0 {
				return fmt.Errorf("project not found: %s", name)
			}

			if err := st.DeleteProject(projectID); err != nil {
				return err
			}

			fmt.Printf("Deleted project: %s (tasks moved to Inbox)\n", name)
			return nil
		},
	}
}
