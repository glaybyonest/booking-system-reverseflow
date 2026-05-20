package domain

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCancelled = "cancelled"
	StatusArchived  = "archived"
)

const (
	SourceManual       = "manual"
	SourceKudaGo       = "kudago"
	SourceTimepad      = "timepad"
	SourceYandexAfisha = "yandex_afisha"
)

const (
	BookingModeReserveFlowManaged = "reserveflow_managed"
	BookingModeExternalLinkOnly   = "external_link_only"
	BookingModeGeneralAdmission   = "general_admission"
)

type ListQuery struct {
	City        string
	Source      string
	Category    string
	From        *time.Time
	To          *time.Time
	BookingMode string
	OnlyActual  bool
	Limit       int
	Offset      int
}

type Event struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Description     *string          `json:"description,omitempty"`
	LongDescription *string          `json:"longDescription,omitempty"`
	Category        *string          `json:"category,omitempty"`
	PosterURL       *string          `json:"posterUrl,omitempty"`
	Status          string           `json:"status"`
	Source          string           `json:"source"`
	ExternalSource  *string          `json:"externalSource,omitempty"`
	SourceURL       *string          `json:"sourceUrl,omitempty"`
	BookingMode     string           `json:"bookingMode"`
	StartsAt        *time.Time       `json:"startsAt,omitempty"`
	EndsAt          *time.Time       `json:"endsAt,omitempty"`
	AgeRestriction  *string          `json:"ageRestriction,omitempty"`
	PriceMin        *float64         `json:"priceMin,omitempty"`
	PriceMax        *float64         `json:"priceMax,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	RatingCount     *int             `json:"ratingCount,omitempty"`
	IsImported      bool             `json:"isImported"`
	Venue           *VenueSummary    `json:"venue,omitempty"`
	ExternalLinks   []ExternalLink   `json:"externalLinks,omitempty"`
	Sessions        []SessionSummary `json:"sessions,omitempty"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

type VenueSummary struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	City          string   `json:"city"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	MetroStations []string `json:"metroStations,omitempty"`
	VenueTypeCode *string  `json:"venueTypeCode,omitempty"`
	VenueTypeName *string  `json:"venueTypeName,omitempty"`
}

type ExternalLink struct {
	ID             string    `json:"id"`
	ExternalSource string    `json:"externalSource"`
	ExternalID     string    `json:"externalId"`
	SourceURL      *string   `json:"sourceUrl,omitempty"`
	ImportedAt     time.Time `json:"importedAt"`
}

type SessionSummary struct {
	ID             string     `json:"id"`
	EventID        string     `json:"eventId"`
	HallID         *string    `json:"hallId,omitempty"`
	HallName       *string    `json:"hallName,omitempty"`
	StartsAt       *time.Time `json:"startsAt,omitempty"`
	EndsAt         *time.Time `json:"endsAt,omitempty"`
	Status         string     `json:"status"`
	IsBookable     bool       `json:"isBookable"`
	ExternalSource *string    `json:"externalSource,omitempty"`
	SourceURL      *string    `json:"sourceUrl,omitempty"`
}

type MapEvent struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Category    *string      `json:"category,omitempty"`
	PosterURL   *string      `json:"posterUrl,omitempty"`
	Source      string       `json:"source"`
	BookingMode string       `json:"bookingMode"`
	StartsAt    *time.Time   `json:"startsAt,omitempty"`
	EndsAt      *time.Time   `json:"endsAt,omitempty"`
	PriceMin    *float64     `json:"priceMin,omitempty"`
	Venue       VenueSummary `json:"venue"`
}
