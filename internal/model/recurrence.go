package model

import "time"

type RecurrencePattern string

const (
	Daily   RecurrencePattern = "daily"
	Weekly  RecurrencePattern = "weekly"
	Monthly RecurrencePattern = "monthly"
	Yearly  RecurrencePattern = "yearly"
)

type Recurrence struct {
	ID       int64
	TaskID   int64
	Pattern  RecurrencePattern
	Interval int
	NextDue  time.Time
}

func NewRecurrence(taskID int64, pattern RecurrencePattern, interval int) *Recurrence {
	if interval < 1 {
		interval = 1
	}
	return &Recurrence{
		TaskID:   taskID,
		Pattern:  pattern,
		Interval: interval,
		NextDue:  time.Now(),
	}
}

func (r *Recurrence) CalculateNextDue(from time.Time) time.Time {
	switch r.Pattern {
	case Daily:
		return from.AddDate(0, 0, r.Interval)
	case Weekly:
		return from.AddDate(0, 0, 7*r.Interval)
	case Monthly:
		return from.AddDate(0, r.Interval, 0)
	case Yearly:
		return from.AddDate(r.Interval, 0, 0)
	}
	return from
}

func (r *Recurrence) PatternString() string {
	if r.Interval == 1 {
		switch r.Pattern {
		case Daily:
			return "daily"
		case Weekly:
			return "weekly"
		case Monthly:
			return "monthly"
		case Yearly:
			return "yearly"
		}
	}

	unit := ""
	switch r.Pattern {
	case Daily:
		unit = "days"
	case Weekly:
		unit = "weeks"
	case Monthly:
		unit = "months"
	case Yearly:
		unit = "years"
	}
	return "every " + string(rune('0'+r.Interval)) + " " + unit
}

func ParseRecurrencePattern(s string) RecurrencePattern {
	switch s {
	case "daily", "d":
		return Daily
	case "weekly", "w":
		return Weekly
	case "monthly", "m":
		return Monthly
	case "yearly", "y":
		return Yearly
	default:
		return Daily
	}
}
