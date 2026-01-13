package model

type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return true
	}
	return false
}

func ParseStatus(s string) Status {
	switch s {
	case "todo":
		return StatusTodo
	case "doing":
		return StatusDoing
	case "done":
		return StatusDone
	default:
		return StatusTodo
	}
}
