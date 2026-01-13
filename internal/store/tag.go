package store

import (
	"database/sql"
	"fmt"

	"github.com/hwanchang/tsk/internal/model"
)

func (s *SQLiteStore) CreateTag(t *model.Tag) error {
	result, err := s.db.Exec(`
		INSERT INTO tags (name, color) VALUES (?, ?)
	`, t.Name, t.Color)
	if err != nil {
		return fmt.Errorf("insert tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	t.ID = id
	return nil
}

func (s *SQLiteStore) GetTag(id int64) (*model.Tag, error) {
	row := s.db.QueryRow("SELECT id, name, color FROM tags WHERE id = ?", id)

	t := &model.Tag{}
	err := row.Scan(&t.ID, &t.Name, &t.Color)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("scan tag: %w", err)
	}
	return t, nil
}

func (s *SQLiteStore) GetTagByName(name string) (*model.Tag, error) {
	row := s.db.QueryRow("SELECT id, name, color FROM tags WHERE name = ?", name)

	t := &model.Tag{}
	err := row.Scan(&t.ID, &t.Name, &t.Color)
	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("scan tag: %w", err)
	}
	return t, nil
}

func (s *SQLiteStore) ListTags() ([]model.Tag, error) {
	rows, err := s.db.Query("SELECT id, name, color FROM tags ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("scan tag row: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *SQLiteStore) DeleteTag(id int64) error {
	// task_tags are automatically deleted via ON DELETE CASCADE
	_, err := s.db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

func (s *SQLiteStore) AddTagToTask(taskID, tagID int64) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)
	`, taskID, tagID)
	if err != nil {
		return fmt.Errorf("add tag to task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) RemoveTagFromTask(taskID, tagID int64) error {
	_, err := s.db.Exec("DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?", taskID, tagID)
	if err != nil {
		return fmt.Errorf("remove tag from task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetTaskTags(taskID int64) ([]model.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN task_tags tt ON t.id = tt.tag_id
		WHERE tt.task_id = ?
		ORDER BY t.name
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("query task tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("scan tag row: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}
