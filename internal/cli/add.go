package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hwanchang/tsk/internal/model"
)

func newAddCmd() *cobra.Command {
	var (
		projectName string
		tagNames    []string
		priority    string
		dueDate     string
		repeat      string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Long:  `Add a new task with optional project, tags, priority, and due date.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")

			task := model.NewTask(title)

			// Set project
			if projectName != "" {
				projects, err := st.ListProjects()
				if err != nil {
					return err
				}
				for _, p := range projects {
					if strings.EqualFold(p.Name, projectName) {
						task.ProjectID = &p.ID
						break
					}
				}
				if task.ProjectID == nil {
					return fmt.Errorf("project not found: %s", projectName)
				}
			}

			// Set priority
			if priority != "" {
				task.Priority = model.ParsePriority(priority)
			}

			// Set due date
			if dueDate != "" {
				due, err := parseDate(dueDate)
				if err != nil {
					return fmt.Errorf("invalid date: %w", err)
				}
				task.DueDate = &due
			}

			// Create task
			if err := st.CreateTask(task); err != nil {
				return err
			}

			// Add tags
			for _, tagName := range tagNames {
				tag, err := st.GetTagByName(tagName)
				if err != nil {
					return err
				}
				if tag == nil {
					// Create tag if not exists
					tag = model.NewTag(tagName)
					if err := st.CreateTag(tag); err != nil {
						return err
					}
				}
				if err := st.AddTagToTask(task.ID, tag.ID); err != nil {
					return err
				}
			}

			// Set recurrence
			if repeat != "" {
				pattern, interval := parseRepeat(repeat)
				rec := model.NewRecurrence(task.ID, pattern, interval)
				// Set next due based on task's due date or today
				if task.DueDate != nil {
					rec.NextDue = rec.CalculateNextDue(*task.DueDate)
				} else {
					rec.NextDue = rec.CalculateNextDue(time.Now())
				}
				if err := st.SetRecurrence(rec); err != nil {
					return err
				}
			}

			fmt.Printf("Created task #%d: %s\n", task.ID, task.Title)
			if repeat != "" {
				fmt.Printf("  Recurrence: %s\n", repeat)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&projectName, "project", "p", "", "project name")
	cmd.Flags().StringSliceVarP(&tagNames, "tag", "t", nil, "tags (can be repeated)")
	cmd.Flags().StringVar(&priority, "priority", "", "priority (low/medium/high)")
	cmd.Flags().StringVarP(&dueDate, "due", "d", "", "due date (today/tomorrow/YYYY-MM-DD)")
	cmd.Flags().StringVarP(&repeat, "repeat", "r", "", "recurrence pattern (daily/weekly/monthly/yearly or daily:2 for every 2 days)")

	return cmd
}

func parseDate(s string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	switch strings.ToLower(s) {
	case "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	case "next week", "nextweek":
		return today.AddDate(0, 0, 7), nil
	}

	// Try parsing as relative days (e.g., "3d" for 3 days)
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil {
			return today.AddDate(0, 0, days), nil
		}
	}

	// Try parsing as date
	formats := []string{
		"2006-01-02",
		"01-02",
		"01/02",
		"Jan 2",
		"January 2",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// If year not specified, use current year
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location())
				// If date has passed, use next year
				if t.Before(now) {
					t = t.AddDate(1, 0, 0)
				}
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

// parseRepeat parses a recurrence pattern string like "daily" or "daily:2"
func parseRepeat(s string) (model.RecurrencePattern, int) {
	parts := strings.Split(s, ":")
	pattern := model.ParseRecurrencePattern(parts[0])
	interval := 1

	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
			interval = n
		}
	}

	return pattern, interval
}
