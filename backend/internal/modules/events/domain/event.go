package domain

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCancelled = "cancelled"
	StatusArchived  = "archived"
)

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Category    *string   `json:"category,omitempty"`
	PosterURL   *string   `json:"posterUrl,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SessionSummary struct {
	ID       string    `json:"id"`
	EventID  string    `json:"eventId"`
	HallID   string    `json:"hallId"`
	HallName string    `json:"hallName"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Status   string    `json:"status"`
}
