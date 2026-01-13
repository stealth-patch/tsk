package store

import (
	"database/sql"
	"fmt"

	"github.com/hwanchang/tsk/internal/model"
)

func (s *SQLiteStore) CreateProject(p *model.Project) error {
	result, err := s.db.Exec(`
		INSERT INTO projects (name, description) VALUES (?, ?)
	`, p.Name, p.Description)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	p.ID = id
	return nil
}

func (s *SQLiteStore) GetProject(id int64) (*model.Project, error) {
	row := s.db.QueryRow(`
		SELECT p.id, p.name, p.description, p.created_at,
			   COUNT(t.id) as task_count,
			   SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) as done_count
		FROM projects p
		LEFT JOIN tasks t ON p.id = t.project_id AND t.parent_id IS NULL
		WHERE p.id = ?
		GROUP BY p.id
	`, id)

	p := &model.Project{}
	var doneCount sql.NullInt64
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.TaskCount, &doneCount)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("scan project: %w", err)
	}
	if doneCount.Valid {
		p.DoneCount = int(doneCount.Int64)
	}
	return p, nil
}

func (s *SQLiteStore) ListProjects() ([]model.Project, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.name, p.description, p.created_at,
			   COUNT(t.id) as task_count,
			   SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) as done_count
		FROM projects p
		LEFT JOIN tasks t ON p.id = t.project_id AND t.parent_id IS NULL
		GROUP BY p.id
		ORDER BY p.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var doneCount sql.NullInt64
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.TaskCount, &doneCount)
		if err != nil {
			return nil, fmt.Errorf("scan project row: %w", err)
		}
		if doneCount.Valid {
			p.DoneCount = int(doneCount.Int64)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *SQLiteStore) DeleteProject(id int64) error {
	// Don't allow deleting the default Inbox project
	if id == 1 {
		return fmt.Errorf("cannot delete default project")
	}

	// Move tasks to Inbox before deleting
	_, err := s.db.Exec("UPDATE tasks SET project_id = 1 WHERE project_id = ?", id)
	if err != nil {
		return fmt.Errorf("move tasks to inbox: %w", err)
	}

	_, err = s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
