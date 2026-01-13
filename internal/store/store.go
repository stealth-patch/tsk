package store

import (
	"github.com/hwanchang/tsk/internal/db"
	"github.com/hwanchang/tsk/internal/model"
)

type TaskFilter struct {
	ProjectID  *int64
	Status     *model.Status
	ParentID   *int64
	TagIDs     []int64
	HasDueDate *bool
	Search     string
	Limit      int
}

type Store interface {
	// Tasks
	CreateTask(t *model.Task) error
	GetTask(id int64) (*model.Task, error)
	ListTasks(filter TaskFilter) ([]model.Task, error)
	UpdateTask(t *model.Task) error
	DeleteTask(id int64) error
	GetSubtasks(parentID int64) ([]model.Task, error)

	// Projects
	CreateProject(p *model.Project) error
	GetProject(id int64) (*model.Project, error)
	ListProjects() ([]model.Project, error)
	DeleteProject(id int64) error

	// Tags
	CreateTag(t *model.Tag) error
	GetTag(id int64) (*model.Tag, error)
	GetTagByName(name string) (*model.Tag, error)
	ListTags() ([]model.Tag, error)
	AddTagToTask(taskID, tagID int64) error
	RemoveTagFromTask(taskID, tagID int64) error
	GetTaskTags(taskID int64) ([]model.Tag, error)

	// Recurrence
	SetRecurrence(r *model.Recurrence) error
	GetRecurrence(taskID int64) (*model.Recurrence, error)
	DeleteRecurrence(taskID int64) error

	// Close
	Close() error
}

type SQLiteStore struct {
	db *db.DB
}

func New(database *db.DB) *SQLiteStore {
	return &SQLiteStore{db: database}
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
