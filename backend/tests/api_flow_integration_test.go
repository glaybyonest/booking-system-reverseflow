//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"reserveflow/backend/internal/app"
)

func TestAPIAuthCatalogSeatMapFlow(t *testing.T) {
	ctx := context.Background()
	server, deps, cleanup := newIntegrationServerWithDeps(t, ctx)
	defer cleanup()

	fixture := seedBookableKudaGoEvent(t, ctx, deps.DB, "API flow concert")

	var register authResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/auth/register", "", map[string]any{
		"email":    "new-user@example.com",
		"password": "Password123!",
		"name":     "New User",
	}, http.StatusCreated, &register)
	require.Equal(t, "new-user@example.com", register.User.Email)
	require.NotEmpty(t, register.Tokens.AccessToken)
	require.NotEmpty(t, register.Tokens.RefreshToken)

	demo := loginDemo(t, server.URL)

	var me userResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/auth/me", demo.Tokens.AccessToken, nil, http.StatusOK, &me)
	require.Equal(t, "demo@example.com", me.Email)

	var events listResponse[eventResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events", "", nil, http.StatusOK, &events)
	require.Len(t, events.Items, 1)
	require.Equal(t, fixture.EventID, events.Items[0].ID)
	require.Equal(t, "kudago", events.Items[0].Source)
	require.Equal(t, "reserveflow_managed", events.Items[0].BookingMode)

	var sessions listResponse[sessionResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/"+fixture.EventID+"/sessions", "", nil, http.StatusOK, &sessions)
	require.Len(t, sessions.Items, 1)
	require.Equal(t, fixture.SessionID, sessions.Items[0].ID)
	require.True(t, sessions.Items[0].IsBookable)

	var session sessionResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+fixture.SessionID, "", nil, http.StatusOK, &session)
	require.Equal(t, fixture.SessionID, session.ID)
	require.Equal(t, fixture.EventID, session.EventID)

	var seatMap seatMapResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+fixture.SessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Equal(t, fixture.SessionID, seatMap.SessionID)
	require.Equal(t, "react_seat_toolkit", seatMap.Provider)
	require.NotNil(t, seatMap.Layout)
	require.Len(t, seatMap.Seats, len(fixture.Layout.Seats))
	require.Equal(t, "available", seatMap.Seats[0].Status)
}

func TestAPIMultiSeatPaymentSuccessFailureIdempotencyAndOwnership(t *testing.T) {
	ctx := context.Background()
	server, deps, cleanup := newIntegrationServerWithDeps(t, ctx)
	defer cleanup()

	fixture := seedBookableKudaGoEvent(t, ctx, deps.DB, "Payment flow concert")
	demo := loginDemo(t, server.URL)
	seats := availableSeats(t, server.URL, fixture.SessionID)
	require.GreaterOrEqual(t, len(seats), 3)

	firstHold := holdSeats(t, server.URL, demo.Tokens.AccessToken, fixture.SessionID, []string{seats[0].SeatID, seats[1].SeatID})
	require.Len(t, firstHold.Seats, 2)
	require.Positive(t, firstHold.TotalPrice)

	var firstBooking bookingResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/bookings/"+firstHold.BookingID, demo.Tokens.AccessToken, nil, http.StatusOK, &firstBooking)
	require.Equal(t, "pending", firstBooking.Status)
	require.Len(t, firstBooking.Items, 2)

	var success paymentResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/payments", demo.Tokens.AccessToken, map[string]any{
		"bookingId":      firstHold.BookingID,
		"idempotencyKey": "success-key",
		"forceStatus":    "succeeded",
	}, http.StatusOK, &success)
	require.Equal(t, "succeeded", success.Status)

	var replay paymentResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/payments", demo.Tokens.AccessToken, map[string]any{
		"bookingId":      firstHold.BookingID,
		"idempotencyKey": "success-key",
		"forceStatus":    "succeeded",
	}, http.StatusOK, &replay)
	require.Equal(t, success.PaymentID, replay.PaymentID)

	secondHold := holdSeat(t, server.URL, demo.Tokens.AccessToken, fixture.SessionID, seats[2].SeatID)
	var idempotencyErr apiErrorResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/payments", demo.Tokens.AccessToken, map[string]any{
		"bookingId":      secondHold.BookingID,
		"idempotencyKey": "success-key",
		"forceStatus":    "succeeded",
	}, http.StatusConflict, &idempotencyErr)
	require.Equal(t, "IDEMPOTENCY_CONFLICT", idempotencyErr.Error.Code)

	var fetched paymentResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/payments/"+success.PaymentID, demo.Tokens.AccessToken, nil, http.StatusOK, &fetched)
	require.Equal(t, success.PaymentID, fetched.PaymentID)

	var other authResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/auth/register", "", map[string]any{
		"email":    "other-user@example.com",
		"password": "Password123!",
		"name":     "Other User",
	}, http.StatusCreated, &other)
	var forbidden apiErrorResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/payments/"+success.PaymentID, other.Tokens.AccessToken, nil, http.StatusForbidden, &forbidden)
	require.Equal(t, "FORBIDDEN", forbidden.Error.Code)

	var failed paymentResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/payments", demo.Tokens.AccessToken, map[string]any{
		"bookingId":      secondHold.BookingID,
		"idempotencyKey": "failed-key",
		"forceStatus":    "failed",
	}, http.StatusOK, &failed)
	require.Equal(t, "failed", failed.Status)

	var failedBooking bookingResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/bookings/"+secondHold.BookingID, demo.Tokens.AccessToken, nil, http.StatusOK, &failedBooking)
	require.Equal(t, "payment_failed", failedBooking.Status)
	require.Len(t, failedBooking.Items, 1)

	latestSeatMap := seatMap(t, server.URL, fixture.SessionID)
	require.Equal(t, "booked", statusForSeat(t, latestSeatMap, seats[0].SeatID))
	require.Equal(t, "booked", statusForSeat(t, latestSeatMap, seats[1].SeatID))
	require.Equal(t, "available", statusForSeat(t, latestSeatMap, seats[2].SeatID))
}

func TestExpirationReleasesPendingBooking(t *testing.T) {
	ctx := context.Background()
	server, deps, cleanup := newIntegrationServerWithDeps(t, ctx)
	defer cleanup()

	fixture := seedBookableKudaGoEvent(t, ctx, deps.DB, "Expiration flow concert")
	demo := loginDemo(t, server.URL)
	seats := availableSeats(t, server.URL, fixture.SessionID)
	require.GreaterOrEqual(t, len(seats), 2)

	hold := holdSeats(t, server.URL, demo.Tokens.AccessToken, fixture.SessionID, []string{seats[0].SeatID, seats[1].SeatID})

	_, err := deps.DB.Exec(ctx, `
		UPDATE bookings SET expires_at = now() - interval '1 minute' WHERE id = $1
	`, hold.BookingID)
	require.NoError(t, err)
	_, err = deps.DB.Exec(ctx, `
		UPDATE session_seats
		SET hold_expires_at = now() - interval '1 minute'
		WHERE session_id = $1 AND seat_id = ANY($2::uuid[])
	`, fixture.SessionID, []string{seats[0].SeatID, seats[1].SeatID})
	require.NoError(t, err)

	changes, err := deps.BookingsService.ExpireBooking(ctx, 100)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, hold.BookingID, changes[0].BookingID)
	require.ElementsMatch(t, []string{seats[0].SeatID, seats[1].SeatID}, changes[0].SeatIDs)

	var expired bookingResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/bookings/"+hold.BookingID, demo.Tokens.AccessToken, nil, http.StatusOK, &expired)
	require.Equal(t, "expired", expired.Status)

	latestSeatMap := seatMap(t, server.URL, fixture.SessionID)
	require.Equal(t, "available", statusForSeat(t, latestSeatMap, seats[0].SeatID))
	require.Equal(t, "available", statusForSeat(t, latestSeatMap, seats[1].SeatID))
}

func newIntegrationServer(t *testing.T, ctx context.Context) (*httptest.Server, func()) {
	t.Helper()
	server, _, cleanup := newIntegrationServerWithDeps(t, ctx)
	return server, cleanup
}

func newIntegrationServerWithDeps(t *testing.T, ctx context.Context) (*httptest.Server, *app.Dependencies, func()) {
	t.Helper()
	databaseURL, cleanupPostgres := startPostgres(t, ctx)
	applyIntegrationMigrations(t, ctx, databaseURL)
	applySQL(t, ctx, databaseURL, filepath.Join("..", "seeds", "dev-users.sql"))

	redisAddr, cleanupRedis := startRedis(t, ctx)
	cfg := app.Config{
		AppEnv:           "test",
		HTTPPort:         "0",
		DatabaseURL:      databaseURL,
		RedisAddr:        redisAddr,
		RedisDB:          0,
		KafkaBrokers:     []string{"localhost:9092"},
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessTTL:     15 * time.Minute,
		JWTRefreshTTL:    24 * time.Hour,
		HoldTTL:          10 * time.Minute,
		SeatmapCacheTTL:  30 * time.Second,
		LogLevel:         "disabled",
	}
	deps, err := app.NewDependencies(ctx, cfg)
	require.NoError(t, err)
	server := httptest.NewServer(app.NewRouter(deps))
	cleanup := func() {
		server.Close()
		deps.Close()
		cleanupRedis()
		cleanupPostgres()
	}
	return server, deps, cleanup
}

func startRedis(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)
	cleanup := func() {
		require.NoError(t, container.Terminate(ctx))
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), cleanup
}

func loginDemo(t *testing.T, baseURL string) authResponse {
	t.Helper()
	var login authResponse
	doJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]any{
		"email":    "demo@example.com",
		"password": "Password123!",
	}, http.StatusOK, &login)
	return login
}

func availableSeats(t *testing.T, baseURL, sessionID string) []seatResponse {
	t.Helper()
	seatMap := seatMap(t, baseURL, sessionID)
	seats := make([]seatResponse, 0)
	for _, seat := range seatMap.Seats {
		if seat.Status == "available" {
			seats = append(seats, seat)
		}
	}
	return seats
}

func seatMap(t *testing.T, baseURL, sessionID string) seatMapResponse {
	t.Helper()
	var seatMap seatMapResponse
	doJSON(t, http.MethodGet, baseURL+"/api/v1/sessions/"+sessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	return seatMap
}

func holdSeat(t *testing.T, baseURL, token, sessionID, seatID string) holdResponse {
	t.Helper()
	return holdSeats(t, baseURL, token, sessionID, []string{seatID})
}

func holdSeats(t *testing.T, baseURL, token, sessionID string, seatIDs []string) holdResponse {
	t.Helper()
	var hold holdResponse
	doJSON(t, http.MethodPost, baseURL+"/api/v1/bookings/hold", token, map[string]any{
		"sessionId": sessionID,
		"seatIds":   seatIDs,
	}, http.StatusCreated, &hold)
	require.Equal(t, "pending", hold.Status)
	return hold
}

func statusForSeat(t *testing.T, seatMap seatMapResponse, seatID string) string {
	t.Helper()
	for _, seat := range seatMap.Seats {
		if seat.SeatID == seatID {
			return seat.Status
		}
	}
	t.Fatalf("seat %s not found", seatID)
	return ""
}

func doJSON(t *testing.T, method, url, token string, payload any, expectedStatus int, out any) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equalf(t, expectedStatus, resp.StatusCode, "response body: %s", string(responseBody))
	if out != nil && len(responseBody) > 0 {
		require.NoError(t, json.Unmarshal(responseBody, out), "response body: %s", string(responseBody))
	}
}

type authResponse struct {
	User   userResponse `json:"user"`
	Tokens struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"tokens"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type listResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type eventResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	BookingMode string `json:"bookingMode"`
}

type sessionResponse struct {
	ID         string `json:"id"`
	EventID    string `json:"eventId"`
	HallID     string `json:"hallId"`
	Status     string `json:"status"`
	IsBookable bool   `json:"isBookable"`
}

type seatMapResponse struct {
	SessionID string              `json:"sessionId"`
	Provider  string              `json:"provider"`
	Layout    *seatLayoutResponse `json:"layout"`
	Seats     []seatResponse      `json:"seats"`
}

type seatLayoutResponse struct {
	Version int `json:"version"`
}

type seatResponse struct {
	SeatID string  `json:"seatId"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}

type holdResponse struct {
	BookingID  string             `json:"bookingId"`
	Status     string             `json:"status"`
	TotalPrice float64            `json:"totalPrice"`
	Seats      []heldSeatResponse `json:"seats"`
}

type heldSeatResponse struct {
	SeatID string `json:"seatId"`
	Row    string `json:"row"`
	Number int    `json:"number"`
}

type paymentResponse struct {
	PaymentID string `json:"paymentId"`
	BookingID string `json:"bookingId"`
	Status    string `json:"status"`
}

type bookingResponse struct {
	ID         string                `json:"id"`
	Status     string                `json:"status"`
	TotalPrice float64               `json:"totalPrice"`
	Items      []bookingItemResponse `json:"items"`
}

type bookingItemResponse struct {
	ID     string  `json:"id"`
	SeatID string  `json:"seatId"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
