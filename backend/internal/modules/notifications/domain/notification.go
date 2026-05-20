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
		return Template{Type: TypeBookingConfirmed, Title: "Бронь подтверждена", Message: "Оплата прошла, место закреплено за вами."}, true
	case "booking.expired":
		return Template{Type: TypeBookingExpired, Title: "Время брони истекло", Message: "Оплата не была завершена вовремя, место снова доступно."}, true
	case "payment.failed":
		return Template{Type: TypePaymentFailed, Title: "Оплата не прошла", Message: "Платеж отклонен, удержание места снято."}, true
	case "booking.cancelled":
		return Template{Type: TypeBookingCancelled, Title: "Бронь отменена", Message: "Ожидающая оплата бронь отменена, место освобождено."}, true
	default:
		return Template{}, false
	}
}
