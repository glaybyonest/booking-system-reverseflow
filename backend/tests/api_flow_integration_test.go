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

const (
	seedEventID   = "40000000-0000-0000-0000-000000000001"
	seedSessionID = "50000000-0000-0000-0000-000000000001"
)

func TestAPIAuthCatalogSeatMapFlow(t *testing.T) {
	ctx := context.Background()
	server, cleanup := newIntegrationServer(t, ctx)
	defer cleanup()

	var register authResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/auth/register", "", map[string]any{
		"email":    "new-user@example.com",
		"password": "Password123!",
		"name":     "New User",
	}, http.StatusCreated, &register)
	require.Equal(t, "new-user@example.com", register.User.Email)
	require.NotEmpty(t, register.Tokens.AccessToken)
	require.NotEmpty(t, register.Tokens.RefreshToken)

	var login authResponse
	doJSON(t, http.MethodPost, server.URL+"/api/v1/auth/login", "", map[string]any{
		"email":    "demo@example.com",
		"password": "Password123!",
	}, http.StatusOK, &login)
	require.Equal(t, "demo@example.com", login.User.Email)
	require.NotEmpty(t, login.Tokens.AccessToken)

	var me userResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/auth/me", login.Tokens.AccessToken, nil, http.StatusOK, &me)
	require.Equal(t, "demo@example.com", me.Email)

	var events listResponse[eventResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events", "", nil, http.StatusOK, &events)
	require.Len(t, events.Items, 3)

	var sessions listResponse[sessionResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/"+seedEventID+"/sessions", "", nil, http.StatusOK, &sessions)
	require.NotEmpty(t, sessions.Items)

	var session sessionResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+seedSessionID, "", nil, http.StatusOK, &session)
	require.Equal(t, seedSessionID, session.ID)

	var seatMap seatMapResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+seedSessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Equal(t, seedSessionID, seatMap.SessionID)
	require.Len(t, seatMap.Seats, 40)
	require.Equal(t, "available", seatMap.Seats[0].Status)
}

func TestAPIPaymentSuccessFailureIdempotencyAndOwnership(t *testing.T) {
	ctx := context.Background()
	server, cleanup := newIntegrationServer(t, ctx)
	defer cleanup()

	demo := loginDemo(t, server.URL)
	seats := availableSeats(t, server.URL, seedSessionID)
	require.GreaterOrEqual(t, len(seats), 2)

	firstHold := holdSeat(t, server.URL, demo.Tokens.AccessToken, seedSessionID, seats[0].SeatID)
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

	secondHold := holdSeat(t, server.URL, demo.Tokens.AccessToken, seedSessionID, seats[1].SeatID)
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

	seatMap := seatMap(t, server.URL, seedSessionID)
	require.Equal(t, "available", statusForSeat(t, seatMap, seats[1].SeatID))
	require.Equal(t, "booked", statusForSeat(t, seatMap, seats[0].SeatID))
}

func TestExpirationReleasesPendingBooking(t *testing.T) {
	ctx := context.Background()
	server, deps, cleanup := newIntegrationServerWithDeps(t, ctx)
	defer cleanup()

	demo := loginDemo(t, server.URL)
	seats := availableSeats(t, server.URL, seedSessionID)
	hold := holdSeat(t, server.URL, demo.Tokens.AccessToken, seedSessionID, seats[0].SeatID)

	_, err := deps.DB.Exec(ctx, `
		UPDATE bookings SET expires_at = now() - interval '1 minute' WHERE id = $1
	`, hold.BookingID)
	require.NoError(t, err)
	_, err = deps.DB.Exec(ctx, `
		UPDATE session_seats SET hold_expires_at = now() - interval '1 minute' WHERE session_id = $1 AND seat_id = $2
	`, seedSessionID, seats[0].SeatID)
	require.NoError(t, err)

	changes, err := deps.BookingsService.ExpireBooking(ctx, 100)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, hold.BookingID, changes[0].BookingID)

	var expired bookingResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/bookings/"+hold.BookingID, demo.Tokens.AccessToken, nil, http.StatusOK, &expired)
	require.Equal(t, "expired", expired.Status)

	seatMap := seatMap(t, server.URL, seedSessionID)
	require.Equal(t, "available", statusForSeat(t, seatMap, seats[0].SeatID))
}

func newIntegrationServer(t *testing.T, ctx context.Context) (*httptest.Server, func()) {
	t.Helper()
	server, _, cleanup := newIntegrationServerWithDeps(t, ctx)
	return server, cleanup
}

func newIntegrationServerWithDeps(t *testing.T, ctx context.Context) (*httptest.Server, *app.Dependencies, func()) {
	t.Helper()
	databaseURL, cleanupPostgres := startPostgres(t, ctx)
	applySQL(t, ctx, databaseURL, filepath.Join("..", "migrations", "000001_init.up.sql"))
	applySQL(t, ctx, databaseURL, filepath.Join("..", "migrations", "000002_seed.up.sql"))

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
	var hold holdResponse
	doJSON(t, http.MethodPost, baseURL+"/api/v1/bookings/hold", token, map[string]any{
		"sessionId": sessionID,
		"seatId":    seatID,
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
}

type eventResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type sessionResponse struct {
	ID      string `json:"id"`
	EventID string `json:"eventId"`
	HallID  string `json:"hallId"`
	Status  string `json:"status"`
}

type seatMapResponse struct {
	SessionID string         `json:"sessionId"`
	Seats     []seatResponse `json:"seats"`
}

type seatResponse struct {
	SeatID string `json:"seatId"`
	Row    string `json:"row"`
	Number int    `json:"number"`
	Status string `json:"status"`
}

type holdResponse struct {
	BookingID string `json:"bookingId"`
	Status    string `json:"status"`
}

type paymentResponse struct {
	PaymentID string `json:"paymentId"`
	BookingID string `json:"bookingId"`
	Status    string `json:"status"`
}

type bookingResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
