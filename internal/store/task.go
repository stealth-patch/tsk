package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hwanchang/tsk/internal/model"
)

func (s *SQLiteStore) CreateTask(t *model.Task) error {
	result, err := s.db.Exec(`
		INSERT INTO tasks (project_id, parent_id, title, description, status, priority, due_date, position)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, t.ProjectID, t.ParentID, t.Title, t.Description, t.Status, t.Priority, t.DueDate, t.Position)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	t.ID = id
	return nil
}

func (s *SQLiteStore) GetTask(id int64) (*model.Task, error) {
	row := s.db.QueryRow(`
		SELECT id, project_id, parent_id, title, description, status, priority, due_date, created_at, completed_at, position
		FROM tasks WHERE id = ?
	`, id)

	t := &model.Task{}
	err := row.Scan(
		&t.ID, &t.ProjectID, &t.ParentID, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.CompletedAt, &t.Position,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}

	// Load tags
	tags, err := s.GetTaskTags(id)
	if err != nil {
		return nil, err
	}
	t.Tags = tags

	return t, nil
}

func (s *SQLiteStore) ListTasks(filter TaskFilter) ([]model.Task, error) {
	query := strings.Builder{}
	args := []interface{}{}

	query.WriteString(`
		SELECT DISTINCT t.id, t.project_id, t.parent_id, t.title, t.description,
		       t.status, t.priority, t.due_date, t.created_at, t.completed_at, t.position
		FROM tasks t
	`)

	// Join for tag filtering
	if len(filter.TagIDs) > 0 {
		query.WriteString(" JOIN task_tags tt ON t.id = tt.task_id")
	}

	query.WriteString(" WHERE 1=1")

	if filter.ProjectID != nil {
		query.WriteString(" AND t.project_id = ?")
		args = append(args, *filter.ProjectID)
	}

	if filter.Status != nil {
		query.WriteString(" AND t.status = ?")
		args = append(args, *filter.Status)
	}

	if filter.ParentID != nil {
		query.WriteString(" AND t.parent_id = ?")
		args = append(args, *filter.ParentID)
	} else {
		// By default, only show top-level tasks
		query.WriteString(" AND t.parent_id IS NULL")
	}

	if filter.HasDueDate != nil {
		if *filter.HasDueDate {
			query.WriteString(" AND t.due_date IS NOT NULL")
		} else {
			query.WriteString(" AND t.due_date IS NULL")
		}
	}

	if filter.Search != "" {
		query.WriteString(" AND (t.title LIKE ? OR t.description LIKE ?)")
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	if len(filter.TagIDs) > 0 {
		placeholders := make([]string, len(filter.TagIDs))
		for i, tagID := range filter.TagIDs {
			placeholders[i] = "?"
			args = append(args, tagID)
		}
		query.WriteString(" AND tt.tag_id IN (" + strings.Join(placeholders, ",") + ")")
	}

	query.WriteString(" ORDER BY t.position ASC, t.created_at DESC")

	if filter.Limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, filter.Limit)
	}

	rows, err := s.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		err := rows.Scan(
			&t.ID, &t.ProjectID, &t.ParentID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.CompletedAt, &t.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}

		// Load tags for each task
		tags, err := s.GetTaskTags(t.ID)
		if err != nil {
			return nil, err
		}
		t.Tags = tags

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *SQLiteStore) UpdateTask(t *model.Task) error {
	_, err := s.db.Exec(`
		UPDATE tasks SET
			project_id = ?, parent_id = ?, title = ?, description = ?,
			status = ?, priority = ?, due_date = ?, completed_at = ?, position = ?
		WHERE id = ?
	`, t.ProjectID, t.ParentID, t.Title, t.Description,
		t.Status, t.Priority, t.DueDate, t.CompletedAt, t.Position, t.ID)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) DeleteTask(id int64) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSubtasks(parentID int64) ([]model.Task, error) {
	return s.ListTasks(TaskFilter{ParentID: &parentID})
}

func (s *SQLiteStore) CompleteTaskWithRecurrence(taskID int64) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	now := time.Now()
	task.Status = model.StatusDone
	task.CompletedAt = &now

	if err := s.UpdateTask(task); err != nil {
		return err
	}

	// Check for recurrence
	rec, err := s.GetRecurrence(taskID)
	if err != nil || rec == nil {
		return nil // No recurrence, done
	}

	// Create next occurrence
	newTask := &model.Task{
		ProjectID:   task.ProjectID,
		ParentID:    task.ParentID,
		Title:       task.Title,
		Description: task.Description,
		Status:      model.StatusTodo,
		Priority:    task.Priority,
		Position:    task.Position,
	}

	nextDue := rec.CalculateNextDue(now)
	newTask.DueDate = &nextDue

	if err := s.CreateTask(newTask); err != nil {
		return err
	}

	// Copy tags
	for _, tag := range task.Tags {
		s.AddTagToTask(newTask.ID, tag.ID)
	}

	// Update recurrence to point to new task
	rec.TaskID = newTask.ID
	rec.NextDue = nextDue
	return s.SetRecurrence(rec)
}
