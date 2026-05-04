package middleware

import (
	"encoding/json"
	"net/http"

	apperrors "reserveflow/backend/internal/infrastructure/errors"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, err error) {
	appErr := apperrors.From(err)
	JSON(w, appErr.HTTPStatus, apperrors.ErrorResponse{Error: appErr})
}
