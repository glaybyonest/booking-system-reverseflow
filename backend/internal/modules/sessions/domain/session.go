package domain

import "time"

const (
	StatusScheduled = "scheduled"
	StatusCancelled = "cancelled"
	StatusFinished  = "finished"
)

type Session struct {
	ID       string    `json:"id"`
	EventID  string    `json:"eventId"`
	Event    EventRef  `json:"event"`
	HallID   string    `json:"hallId"`
	Hall     HallRef   `json:"hall"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Status   string    `json:"status"`
}

type EventRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type HallRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Venue string `json:"venue"`
}
