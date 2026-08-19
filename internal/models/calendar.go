package models

import "time"

type CalendarEvent struct {
	UID         string    `json:"uid"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Title       string    `json:"title"`
	LessonType  string    `json:"lesson_type"`
	Teachers    []string  `json:"teachers"`
	Auditoriums []string  `json:"auditoriums"`
	Sequence    int64     `json:"sequence"`
	Cancelled   bool      `json:"cancelled"`
}

type GroupCalendar struct {
	Group     string          `json:"group"`
	Source    string          `json:"-"`
	UpdatedAt time.Time       `json:"updated_at"`
	Events    []CalendarEvent `json:"events"`
}
