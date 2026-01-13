package app

import (
	"github.com/hwanchang/tsk/internal/model"
)

// Messages for inter-component communication

// TasksLoadedMsg is sent when tasks are loaded from store
type TasksLoadedMsg struct {
	Tasks []model.Task
}

// ProjectsLoadedMsg is sent when projects are loaded
type ProjectsLoadedMsg struct {
	Projects []model.Project
}

// TagsLoadedMsg is sent when tags are loaded
type TagsLoadedMsg struct {
	Tags []model.Tag
}

// TaskCreatedMsg is sent when a new task is created
type TaskCreatedMsg struct {
	Task *model.Task
}

// TaskUpdatedMsg is sent when a task is updated
type TaskUpdatedMsg struct {
	Task *model.Task
}

// TaskDeletedMsg is sent when a task is deleted
type TaskDeletedMsg struct {
	ID int64
}

// ProjectCreatedMsg is sent when a new project is created
type ProjectCreatedMsg struct {
	Project *model.Project
}

// TagCreatedMsg is sent when a new tag is created
type TagCreatedMsg struct {
	Tag *model.Tag
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// StatusMsg is sent to show a status message
type StatusMsg struct {
	Text    string
	IsError bool
}

// ClearStatusMsg clears the status message
type ClearStatusMsg struct{}

// EnterInputMode starts text input mode
type EnterInputMode struct {
	Prompt string
	Action string // "add", "edit", "search"
}

// SubmitInputMsg is sent when input is submitted
type SubmitInputMsg struct {
	Text   string
	Action string
}

// ChangeViewMsg changes the current view
type ChangeViewMsg struct {
	View ViewType
}
