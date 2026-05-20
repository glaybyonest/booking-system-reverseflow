package domain

import "time"

const (
	StatusScheduled = "scheduled"
	StatusCancelled = "cancelled"
	StatusFinished  = "finished"
)

type Session struct {
	ID             string     `json:"id"`
	EventID        string     `json:"eventId"`
	Event          EventRef   `json:"event"`
	HallID         *string    `json:"hallId,omitempty"`
	Hall           *HallRef   `json:"hall,omitempty"`
	StartsAt       *time.Time `json:"startsAt,omitempty"`
	EndsAt         *time.Time `json:"endsAt,omitempty"`
	Status         string     `json:"status"`
	IsBookable     bool       `json:"isBookable"`
	ExternalSource *string    `json:"externalSource,omitempty"`
	ExternalID     *string    `json:"externalId,omitempty"`
	SourceURL      *string    `json:"sourceUrl,omitempty"`
}

type EventRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type HallRef struct {
	ID    *string `json:"id,omitempty"`
	Name  string  `json:"name"`
	Venue string  `json:"venue"`
}
