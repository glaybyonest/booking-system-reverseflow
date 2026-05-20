package kafka

import "time"

const (
	TopicSeatHeld             = "seat.held"
	TopicBookingCreated       = "booking.created"
	TopicPaymentSucceeded     = "payment.succeeded"
	TopicBookingConfirmed     = "booking.confirmed"
	TopicBookingExpired       = "booking.expired"
	TopicPaymentFailed        = "payment.failed"
	TopicBookingCancelled     = "booking.cancelled"
	TopicExternalEventsSynced = "external_events.synced"
)

var DomainTopics = []string{
	TopicSeatHeld,
	TopicBookingCreated,
	TopicPaymentSucceeded,
	TopicBookingConfirmed,
	TopicBookingExpired,
	TopicPaymentFailed,
	TopicBookingCancelled,
	TopicExternalEventsSynced,
}

var NotificationTopics = []string{
	TopicBookingConfirmed,
	TopicBookingExpired,
	TopicPaymentFailed,
	TopicBookingCancelled,
}

type EventEnvelope struct {
	EventID       string         `json:"eventId"`
	EventType     string         `json:"eventType"`
	AggregateType string         `json:"aggregateType"`
	AggregateID   string         `json:"aggregateId"`
	OccurredAt    time.Time      `json:"occurredAt"`
	Payload       map[string]any `json:"payload"`
}
