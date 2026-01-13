package store

import (
	"database/sql"
	"fmt"

	"github.com/hwanchang/tsk/internal/model"
)

func (s *SQLiteStore) SetRecurrence(r *model.Recurrence) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO recurrences (task_id, pattern, interval, next_due)
		VALUES (?, ?, ?, ?)
	`, r.TaskID, r.Pattern, r.Interval, r.NextDue)
	if err != nil {
		return fmt.Errorf("set recurrence: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetRecurrence(taskID int64) (*model.Recurrence, error) {
	row := s.db.QueryRow(`
		SELECT id, task_id, pattern, interval, next_due
		FROM recurrences WHERE task_id = ?
	`, taskID)

	r := &model.Recurrence{}
	err := row.Scan(&r.ID, &r.TaskID, &r.Pattern, &r.Interval, &r.NextDue)
	if err == sql.ErrNoRows {
		return nil, nil // No recurrence set
	}
	if err != nil {
		return nil, fmt.Errorf("scan recurrence: %w", err)
	}
	return r, nil
}

func (s *SQLiteStore) DeleteRecurrence(taskID int64) error {
	_, err := s.db.Exec("DELETE FROM recurrences WHERE task_id = ?", taskID)
	if err != nil {
		return fmt.Errorf("delete recurrence: %w", err)
	}
	return nil
}
