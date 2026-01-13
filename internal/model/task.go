package model

import "time"

type Task struct {
	ID          int64
	ProjectID   *int64
	ParentID    *int64
	Title       string
	Description string
	Status      Status
	Priority    Priority
	DueDate     *time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
	Position    int

	// Relations (populated on join)
	Tags       []Tag
	Subtasks   []Task
	Recurrence *Recurrence
}

func NewTask(title string) *Task {
	return &Task{
		Title:     title,
		Status:    StatusTodo,
		Priority:  PriorityNone,
		CreatedAt: time.Now(),
	}
}

func (t *Task) MarkDone() {
	now := time.Now()
	t.Status = StatusDone
	t.CompletedAt = &now
}

func (t *Task) MarkDoing() {
	t.Status = StatusDoing
	t.CompletedAt = nil
}

func (t *Task) MarkTodo() {
	t.Status = StatusTodo
	t.CompletedAt = nil
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

func (t *Task) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now()
	return t.DueDate.Year() == now.Year() &&
		t.DueDate.YearDay() == now.YearDay()
}
