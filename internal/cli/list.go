package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/hwanchang/tsk/internal/model"
	"github.com/hwanchang/tsk/internal/store"
)

func newListCmd() *cobra.Command {
	var (
		status      string
		projectName string
		tagName     string
		all         bool
		format      string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := store.TaskFilter{}

			// Status filter
			if status != "" {
				s := model.ParseStatus(status)
				filter.Status = &s
			} else if !all {
				// By default, don't show completed tasks
				todo := model.StatusTodo
				doing := model.StatusDoing
				// We'll filter manually since we want both todo and doing
				_ = todo
				_ = doing
			}

			// Project filter
			if projectName != "" {
				projects, err := st.ListProjects()
				if err != nil {
					return err
				}
				for _, p := range projects {
					if strings.EqualFold(p.Name, projectName) {
						filter.ProjectID = &p.ID
						break
					}
				}
			}

			// Tag filter
			if tagName != "" {
				tag, err := st.GetTagByName(tagName)
				if err != nil {
					return err
				}
				if tag != nil {
					filter.TagIDs = []int64{tag.ID}
				}
			}

			tasks, err := st.ListTasks(filter)
			if err != nil {
				return err
			}

			// Filter out done tasks unless -a flag or specific status
			if !all && status == "" {
				var filtered []model.Task
				for _, t := range tasks {
					if t.Status != model.StatusDone {
						filtered = append(filtered, t)
					}
				}
				tasks = filtered
			}

			if format == "json" {
				return printJSON(tasks)
			}

			return printTable(tasks)
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status (todo/doing/done)")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "filter by project")
	cmd.Flags().StringVarP(&tagName, "tag", "t", "", "filter by tag")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all tasks including done")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "output format (table/json)")

	return cmd
}

func printTable(tasks []model.Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tPRIORITY\tTITLE\tDUE\tTAGS")

	for _, t := range tasks {
		status := statusIcon(t.Status)
		priority := t.Priority.Icon()
		due := formatDue(t.DueDate)
		tags := formatTags(t.Tags)

		title := t.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, status, priority, title, due, tags)
	}

	return w.Flush()
}

func printJSON(tasks []model.Task) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(tasks)
}

func statusIcon(s model.Status) string {
	switch s {
	case model.StatusTodo:
		return "[ ]"
	case model.StatusDoing:
		return "[~]"
	case model.StatusDone:
		return "[x]"
	default:
		return "[ ]"
	}
}

func formatDue(due *time.Time) string {
	if due == nil {
		return "-"
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location())

	diff := dueDay.Sub(today).Hours() / 24

	switch {
	case diff < 0:
		return fmt.Sprintf("OVERDUE (%s)", due.Format("Jan 2"))
	case diff == 0:
		return "Today"
	case diff == 1:
		return "Tomorrow"
	case diff <= 7:
		return due.Format("Mon")
	default:
		return due.Format("Jan 2")
	}
}

func formatTags(tags []model.Tag) string {
	if len(tags) == 0 {
		return "-"
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}
