package errors

import (
	stderrors "errors"
	"net/http"
)

const (
	CodeValidationError         = "VALIDATION_ERROR"
	CodeUnauthorized            = "UNAUTHORIZED"
	CodeForbidden               = "FORBIDDEN"
	CodeNotFound                = "NOT_FOUND"
	CodeSeatNotAvailable        = "SEAT_NOT_AVAILABLE"
	CodeSeatAlreadyHeld         = "SEAT_ALREADY_HELD"
	CodeSessionNotBookable      = "SESSION_NOT_BOOKABLE"
	CodeBookingNotFound         = "BOOKING_NOT_FOUND"
	CodeBookingExpired          = "BOOKING_EXPIRED"
	CodeBookingNotPending       = "BOOKING_NOT_PENDING"
	CodePaymentAlreadyProcessed = "PAYMENT_ALREADY_PROCESSED"
	CodeIdempotencyConflict     = "IDEMPOTENCY_CONFLICT"
	CodeInternalError           = "INTERNAL_ERROR"
)

type AppError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details"`
	HTTPStatus int            `json:"-"`
	Err        error          `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Details: map[string]any{}, HTTPStatus: status}
}

func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Details: map[string]any{}, HTTPStatus: status, Err: err}
}

func Validation(message string) *AppError {
	return New(CodeValidationError, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

func Conflict(code, message string) *AppError {
	return New(code, message, http.StatusConflict)
}

func Internal(err error) *AppError {
	return Wrap(err, CodeInternalError, "Internal server error", http.StatusInternalServerError)
}

func From(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}

type ErrorResponse struct {
	Error *AppError `json:"error"`
}
