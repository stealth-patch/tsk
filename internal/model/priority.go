package model

type Priority int

const (
	PriorityNone Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	return [...]string{"", "Low", "Medium", "High"}[p]
}

func (p Priority) Icon() string {
	return [...]string{"", "↓", "→", "↑"}[p]
}

func (p Priority) IsValid() bool {
	return p >= PriorityNone && p <= PriorityHigh
}

func ParsePriority(s string) Priority {
	switch s {
	case "low", "l", "1":
		return PriorityLow
	case "medium", "med", "m", "2":
		return PriorityMedium
	case "high", "h", "3":
		return PriorityHigh
	default:
		return PriorityNone
	}
}
