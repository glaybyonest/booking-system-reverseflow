package domain

import "time"

const (
	TypeBookingConfirmed = "booking_confirmed"
	TypeBookingExpired   = "booking_expired"
	TypePaymentFailed    = "payment_failed"
	TypeBookingCancelled = "booking_cancelled"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

type Template struct {
	Type    string
	Title   string
	Message string
}

func TemplateForEvent(eventType string) (Template, bool) {
	switch eventType {
	case "booking.confirmed":
		return Template{Type: TypeBookingConfirmed, Title: "Booking confirmed", Message: "Your booking has been confirmed."}, true
	case "booking.expired":
		return Template{Type: TypeBookingExpired, Title: "Booking expired", Message: "Your booking expired because payment was not completed in time."}, true
	case "payment.failed":
		return Template{Type: TypePaymentFailed, Title: "Payment failed", Message: "Payment failed and the seat was released."}, true
	case "booking.cancelled":
		return Template{Type: TypeBookingCancelled, Title: "Booking cancelled", Message: "Your pending booking was cancelled."}, true
	default:
		return Template{}, false
	}
}
