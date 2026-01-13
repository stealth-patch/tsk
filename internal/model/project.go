package model

import "time"

type Project struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time

	// Computed stats
	TaskCount int
	DoneCount int
}

func NewProject(name string) *Project {
	return &Project{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

func (p *Project) Progress() float64 {
	if p.TaskCount == 0 {
		return 0
	}
	return float64(p.DoneCount) / float64(p.TaskCount)
}
