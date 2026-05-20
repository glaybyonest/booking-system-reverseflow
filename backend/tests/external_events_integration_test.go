//go:build integration

package tests

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"reserveflow/backend/internal/app"
	integrationsapp "reserveflow/backend/internal/modules/integrations/application"
	integrationsrepo "reserveflow/backend/internal/modules/integrations/repository"
)

func TestRepeatedSyncDoesNotDuplicateEventsAndSoftDedupesProviders(t *testing.T) {
	ctx := context.Background()
	kudaGoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "",
			"results": [
				{
					"id": 701,
					"title": "Imported concert",
					"description": "KudaGo description",
					"categories": ["concert"],
					"site_url": "https://kudago.com/msk/event/test-import/",
					"images": [{"image": "https://images.example.com/kudago.jpg"}],
					"is_free": false,
					"place": {
						"id": 9001,
						"title": "Music house",
						"address": "Moscow, embankment 52",
						"coords": {"lat": 55.7343, "lon": 37.6467}
					},
					"dates": [{"start": 1893456000, "end": 1893463200}]
				}
			]
		}`))
	}))
	defer kudaGoServer.Close()

	timepadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"total": 1,
			"values": [
				{
					"id": 8801,
					"name": "Imported concert",
					"description_short": "Timepad description",
					"url": "https://timepad.ru/event/8801/",
					"starts_at": "2030-01-01T18:00:00+03:00",
					"ends_at": "2030-01-01T20:00:00+03:00",
					"location": {
						"name": "Music house",
						"address": "Moscow, embankment 52",
						"city": "Москва",
						"coordinates": {"lat": 55.7343, "lon": 37.6467}
					}
				}
			]
		}`))
	}))
	defer timepadServer.Close()

	databaseURL, cleanupPostgres := startPostgres(t, ctx)
	defer cleanupPostgres()
	applyIntegrationMigrations(t, ctx, databaseURL)

	deps, cleanup := newIntegrationDepsOnly(t, ctx, app.Config{
		AppEnv:                     "test",
		DatabaseURL:                databaseURL,
		RedisAddr:                  "localhost:6379",
		JWTAccessSecret:            "test-access-secret",
		JWTRefreshSecret:           "test-refresh-secret",
		JWTAccessTTL:               15 * time.Minute,
		JWTRefreshTTL:              24 * time.Hour,
		HoldTTL:                    10 * time.Minute,
		SeatmapCacheTTL:            30 * time.Second,
		LogLevel:                   "disabled",
		ExternalSyncCity:           "moscow",
		ExternalSyncKudaGoLocation: "msk",
		ExternalSyncTimepadCity:    "Москва",
		ExternalSyncDaysAhead:      180,
		ExternalSyncLookbackDays:   14,
		ExternalSyncMaxPages:       10,
		ExternalSyncPageSize:       100,
		KudaGoBaseURL:              kudaGoServer.URL,
		TimepadBaseURL:             timepadServer.URL,
	})
	defer cleanup()

	service := integrationsapp.NewService(
		integrationsrepo.NewPostgresRepository(deps.DB, ""),
		deps.Log,
		integrationsapp.Options{
			SyncCity:       "moscow",
			KudaGoLocation: "msk",
			TimepadCity:    "Москва",
			DaysAhead:      180,
			LookbackDays:   14,
			MaxPages:       10,
			PageSize:       100,
			KudaGoBaseURL:  kudaGoServer.URL,
			TimepadBaseURL: timepadServer.URL,
		},
	)

	firstRun, err := service.SyncKudaGo(ctx, "msk", 180, 14)
	require.NoError(t, err)
	require.Equalf(t, "success", firstRun.Status, "kudago error: %s", stringValue(firstRun.ErrorMessage))
	require.Equal(t, 1, firstRun.ImportedCount)

	secondRun, err := service.SyncKudaGo(ctx, "msk", 180, 14)
	require.NoError(t, err)
	require.Equalf(t, "success", secondRun.Status, "kudago error: %s", stringValue(secondRun.ErrorMessage))
	require.Equal(t, 1, secondRun.UpdatedCount)

	timepadRun, err := service.SyncTimepad(ctx, "Москва", 180, 14)
	require.NoError(t, err)
	require.Equalf(t, "success", timepadRun.Status, "timepad error: %s", stringValue(timepadRun.ErrorMessage))
	require.Equal(t, 1, timepadRun.DuplicateCount)

	var eventsCount int
	require.NoError(t, deps.DB.QueryRow(ctx, `SELECT count(*) FROM events`).Scan(&eventsCount))
	require.Equal(t, 1, eventsCount)

	var linksCount int
	require.NoError(t, deps.DB.QueryRow(ctx, `SELECT count(*) FROM event_external_links`).Scan(&linksCount))
	require.Equal(t, 1, linksCount)
}

func TestSyncSelectedProvidersDefaultsToKudaGoOnly(t *testing.T) {
	ctx := context.Background()
	kudaGoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "",
			"results": [
				{
					"id": 703,
					"title": "Default sync",
					"description": "KudaGo only",
					"categories": ["concert"],
					"site_url": "https://kudago.com/msk/event/default-kudago/",
					"images": [{"image": "https://images.example.com/kudago-default.jpg"}],
					"is_free": false,
					"place": {
						"id": 9003,
						"title": "Map point",
						"address": "Moscow, Tverskaya 10",
						"coords": {"lat": 55.765, "lon": 37.605}
					},
					"dates": [{"start": 1893628800, "end": 1893636000}]
				}
			]
		}`))
	}))
	defer kudaGoServer.Close()

	databaseURL, cleanupPostgres := startPostgres(t, ctx)
	defer cleanupPostgres()
	applyIntegrationMigrations(t, ctx, databaseURL)

	deps, cleanup := newIntegrationDepsOnly(t, ctx, app.Config{
		AppEnv:                     "test",
		DatabaseURL:                databaseURL,
		RedisAddr:                  "localhost:6379",
		JWTAccessSecret:            "test-access-secret",
		JWTRefreshSecret:           "test-refresh-secret",
		JWTAccessTTL:               15 * time.Minute,
		JWTRefreshTTL:              24 * time.Hour,
		HoldTTL:                    10 * time.Minute,
		SeatmapCacheTTL:            30 * time.Second,
		LogLevel:                   "disabled",
		ExternalSyncCity:           "moscow",
		ExternalSyncKudaGoLocation: "msk",
		ExternalSyncTimepadCity:    "Москва",
		ExternalSyncDaysAhead:      180,
		ExternalSyncLookbackDays:   14,
		ExternalSyncMaxPages:       10,
		ExternalSyncPageSize:       100,
		KudaGoBaseURL:              kudaGoServer.URL,
		TimepadBaseURL:             "http://unused.local",
	})
	defer cleanup()

	service := integrationsapp.NewService(
		integrationsrepo.NewPostgresRepository(deps.DB, ""),
		deps.Log,
		integrationsapp.Options{
			SyncCity:       "moscow",
			KudaGoLocation: "msk",
			TimepadCity:    "Москва",
			DaysAhead:      180,
			LookbackDays:   14,
			MaxPages:       10,
			PageSize:       100,
			KudaGoBaseURL:  kudaGoServer.URL,
			TimepadBaseURL: "http://unused.local",
		},
	)

	runs, err := service.SyncSelectedProviders(ctx, nil, 180, 14)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.Equal(t, "kudago", runs[0].Provider)
	require.Equalf(t, "success", runs[0].Status, "kudago error: %s", stringValue(runs[0].ErrorMessage))
	require.Equal(t, 1, runs[0].ImportedCount)

	var eventsCount int
	require.NoError(t, deps.DB.QueryRow(ctx, `SELECT count(*) FROM events WHERE source = 'kudago'`).Scan(&eventsCount))
	require.Equal(t, 1, eventsCount)
}

func TestAdminSyncMoscowDefaultsToKudaGoOnly(t *testing.T) {
	ctx := context.Background()
	kudaGoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "",
			"results": [
				{
					"id": 705,
					"title": "Admin sync",
					"description": "KudaGo only via admin endpoint",
					"categories": ["concert"],
					"site_url": "https://kudago.com/msk/event/admin-sync-default/",
					"images": [{"image": "https://images.example.com/kudago-admin.jpg"}],
					"is_free": false,
					"place": {
						"id": 9005,
						"title": "Sync point",
						"address": "Moscow, Arbat 20",
						"coords": {"lat": 55.7528, "lon": 37.5929}
					},
					"dates": [{"start": 1893801600, "end": 1893808800}]
				}
			]
		}`))
	}))
	defer kudaGoServer.Close()

	var timepadRequests atomic.Int32
	timepadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timepadRequests.Add(1)
		http.Error(w, "timepad should not be called", http.StatusInternalServerError)
	}))
	defer timepadServer.Close()

	server, deps, cleanup := newConfiguredIntegrationServerWithDeps(t, ctx, func(cfg *app.Config) {
		cfg.KudaGoBaseURL = kudaGoServer.URL
		cfg.TimepadBaseURL = timepadServer.URL
		cfg.ExternalSyncKudaGoLocation = "msk"
		cfg.ExternalSyncTimepadCity = "Москва"
		cfg.ExternalSyncDaysAhead = 180
		cfg.ExternalSyncLookbackDays = 14
		cfg.ExternalSyncMaxPages = 10
		cfg.ExternalSyncPageSize = 100
		cfg.LogLevel = "disabled"
	})
	defer cleanup()

	admin := loginAdmin(t, server.URL)

	var response struct {
		Runs []struct {
			Provider      string `json:"provider"`
			Status        string `json:"status"`
			ImportedCount int    `json:"importedCount"`
		} `json:"runs"`
	}
	doJSON(t, http.MethodPost, server.URL+"/api/v1/admin/integrations/sync/moscow", admin.Tokens.AccessToken, nil, http.StatusOK, &response)

	require.Len(t, response.Runs, 1)
	require.Equal(t, "kudago", response.Runs[0].Provider)
	require.Equalf(t, "success", response.Runs[0].Status, "admin sync should succeed")
	require.Equal(t, 1, response.Runs[0].ImportedCount)
	require.Zero(t, timepadRequests.Load())

	var eventsCount int
	require.NoError(t, deps.DB.QueryRow(ctx, `SELECT count(*) FROM events WHERE source = 'kudago'`).Scan(&eventsCount))
	require.Equal(t, 1, eventsCount)
}

func TestImportedKudaGoEventRemainsHiddenUntilLayoutIsAttached(t *testing.T) {
	ctx := context.Background()
	kudaGoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "",
			"results": [
				{
					"id": 704,
					"title": "Hidden until layout",
					"description": "Event becomes public only after layout",
					"categories": ["concert"],
					"site_url": "https://kudago.com/msk/event/map-kudago/",
					"images": [{"image": "https://images.example.com/kudago-map.jpg"}],
					"is_free": false,
					"place": {
						"id": 9004,
						"title": "Yandex stage",
						"address": "Moscow, New Arbat 15",
						"coords": {"lat": 55.7522, "lon": 37.5925}
					},
					"dates": [{"start": 1893715200, "end": 1893722400}]
				}
			]
		}`))
	}))
	defer kudaGoServer.Close()

	server, deps, cleanup := newConfiguredIntegrationServerWithDeps(t, ctx, func(cfg *app.Config) {
		cfg.KudaGoBaseURL = kudaGoServer.URL
		cfg.TimepadBaseURL = "http://unused.local"
		cfg.ExternalSyncKudaGoLocation = "msk"
		cfg.ExternalSyncTimepadCity = "Москва"
		cfg.ExternalSyncDaysAhead = 180
		cfg.ExternalSyncLookbackDays = 14
		cfg.ExternalSyncMaxPages = 10
		cfg.ExternalSyncPageSize = 100
		cfg.LogLevel = "disabled"
	})
	defer cleanup()

	run, err := deps.IntegrationsService.SyncKudaGo(ctx, "msk", 180, 14)
	require.NoError(t, err)
	require.Equalf(t, "success", run.Status, "kudago error: %s", stringValue(run.ErrorMessage))

	var importedEventID, sessionID, hallID string
	require.NoError(t, deps.DB.QueryRow(ctx, `
		SELECT e.id, s.id, s.hall_id
		FROM events e
		JOIN sessions s ON s.event_id = e.id
		WHERE e.source = 'kudago'
		LIMIT 1
	`).Scan(&importedEventID, &sessionID, &hallID))

	var events listResponse[eventCatalogResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events", "", nil, http.StatusOK, &events)
	require.Empty(t, events.Items)

	var mapResponse struct {
		Events []eventMapResponse `json:"events"`
	}
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/map", "", nil, http.StatusOK, &mapResponse)
	require.Empty(t, mapResponse.Events)

	var detail eventDetailResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/"+importedEventID, "", nil, http.StatusOK, &detail)
	require.Equal(t, "external_link_only", detail.BookingMode)
	require.Equal(t, "kudago", detail.Source)
	require.Len(t, detail.Sessions, 1)
	require.False(t, detail.Sessions[0].IsBookable)
	require.Equal(t, hallID, detail.Sessions[0].HallID)

	admin := loginAdmin(t, server.URL)

	var initialState sessionLayoutStateResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &initialState)
	require.Equal(t, "none", initialState.LayoutSource)
	require.Nil(t, initialState.EffectiveLayout)

	layout := sampleStoredLayout(2, 3)
	var mutation layoutMutationResponse
	doJSON(t, http.MethodPut, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, map[string]any{
		"layout": layout,
	}, http.StatusOK, &mutation)
	require.True(t, mutation.IsBookable)
	require.True(t, mutation.VisibleInCatalog)
	require.Equal(t, "reserveflow_managed", mutation.BookingMode)
	require.Equal(t, 6, mutation.MaterializedSeats)

	doJSON(t, http.MethodGet, server.URL+"/api/v1/events", "", nil, http.StatusOK, &events)
	require.Len(t, events.Items, 1)
	require.Equal(t, importedEventID, events.Items[0].ID)

	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/map", "", nil, http.StatusOK, &mapResponse)
	require.Len(t, mapResponse.Events, 1)
	require.Equal(t, importedEventID, mapResponse.Events[0].ID)

	var seatMap seatMapResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+sessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Equal(t, "react_seat_toolkit", seatMap.Provider)
	require.NotNil(t, seatMap.Layout)
	require.Len(t, seatMap.Seats, 6)
}

func TestHallFallbackLayoutAndSessionOverride(t *testing.T) {
	ctx := context.Background()
	kudaGoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "",
			"results": [
				{
					"id": 706,
					"title": "Fallback layout concert",
					"description": "Hall fallback and session override",
					"categories": ["concert"],
					"site_url": "https://kudago.com/msk/event/fallback-layout/",
					"images": [{"image": "https://images.example.com/fallback-layout.jpg"}],
					"is_free": false,
					"place": {
						"id": 9006,
						"title": "Layout hall",
						"address": "Moscow, Layout street 6",
						"coords": {"lat": 55.761, "lon": 37.61}
					},
					"dates": [{"start": 1893888000, "end": 1893895200}]
				}
			]
		}`))
	}))
	defer kudaGoServer.Close()

	server, deps, cleanup := newConfiguredIntegrationServerWithDeps(t, ctx, func(cfg *app.Config) {
		cfg.KudaGoBaseURL = kudaGoServer.URL
		cfg.TimepadBaseURL = "http://unused.local"
		cfg.ExternalSyncKudaGoLocation = "msk"
		cfg.ExternalSyncTimepadCity = "Москва"
		cfg.ExternalSyncDaysAhead = 180
		cfg.ExternalSyncLookbackDays = 14
		cfg.ExternalSyncMaxPages = 10
		cfg.ExternalSyncPageSize = 100
		cfg.LogLevel = "disabled"
	})
	defer cleanup()

	run, err := deps.IntegrationsService.SyncKudaGo(ctx, "msk", 180, 14)
	require.NoError(t, err)
	require.Equalf(t, "success", run.Status, "kudago error: %s", stringValue(run.ErrorMessage))

	var eventID, sessionID, hallID string
	require.NoError(t, deps.DB.QueryRow(ctx, `
		SELECT e.id, s.id, s.hall_id
		FROM events e
		JOIN sessions s ON s.event_id = e.id
		WHERE e.source = 'kudago'
		LIMIT 1
	`).Scan(&eventID, &sessionID, &hallID))

	admin := loginAdmin(t, server.URL)
	hallLayout := sampleStoredLayout(2, 2)
	var hallMutation layoutMutationResponse
	doJSON(t, http.MethodPut, server.URL+"/api/v1/admin/halls/"+hallID+"/layout", admin.Tokens.AccessToken, map[string]any{
		"layout": hallLayout,
	}, http.StatusOK, &hallMutation)
	require.True(t, hallMutation.IsBookable)
	require.True(t, hallMutation.VisibleInCatalog)
	require.Equal(t, 4, hallMutation.MaterializedSeats)

	var hallState hallLayoutStateResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/admin/halls/"+hallID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &hallState)
	require.NotNil(t, hallState.Layout)
	require.Len(t, hallState.Sessions, 1)
	require.Equal(t, sessionID, hallState.Sessions[0].ID)

	var fallbackState sessionLayoutStateResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &fallbackState)
	require.Equal(t, "hall", fallbackState.LayoutSource)
	require.Nil(t, fallbackState.Layout)
	require.NotNil(t, fallbackState.FallbackLayout)
	require.NotNil(t, fallbackState.EffectiveLayout)

	var seatMap seatMapResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+sessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Len(t, seatMap.Seats, 4)

	sessionLayout := sampleStoredLayout(2, 3)
	var sessionMutation layoutMutationResponse
	doJSON(t, http.MethodPut, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, map[string]any{
		"layout": sessionLayout,
	}, http.StatusOK, &sessionMutation)
	require.Equal(t, eventID, sessionMutation.EventID)
	require.Equal(t, 6, sessionMutation.MaterializedSeats)

	var overrideState sessionLayoutStateResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &overrideState)
	require.Equal(t, "session", overrideState.LayoutSource)
	require.NotNil(t, overrideState.Layout)
	require.NotNil(t, overrideState.FallbackLayout)
	require.NotNil(t, overrideState.EffectiveLayout)

	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+sessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Len(t, seatMap.Seats, 6)

	var deleteMutation layoutMutationResponse
	doJSON(t, http.MethodDelete, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &deleteMutation)
	require.True(t, deleteMutation.IsBookable)
	require.Equal(t, 4, deleteMutation.MaterializedSeats)

	var restoredState sessionLayoutStateResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/admin/sessions/"+sessionID+"/layout", admin.Tokens.AccessToken, nil, http.StatusOK, &restoredState)
	require.Equal(t, "hall", restoredState.LayoutSource)
	require.Nil(t, restoredState.Layout)
	require.NotNil(t, restoredState.FallbackLayout)

	doJSON(t, http.MethodGet, server.URL+"/api/v1/sessions/"+sessionID+"/seats", "", nil, http.StatusOK, &seatMap)
	require.Len(t, seatMap.Seats, 4)
}

func newConfiguredIntegrationServerWithDeps(t *testing.T, ctx context.Context, mutate func(*app.Config)) (*httptest.Server, *app.Dependencies, func()) {
	t.Helper()
	databaseURL, cleanupPostgres := startPostgres(t, ctx)
	applyIntegrationMigrations(t, ctx, databaseURL)
	applySQL(t, ctx, databaseURL, filepath.Join("..", "seeds", "dev-users.sql"))

	redisAddr, cleanupRedis := startRedis(t, ctx)
	cfg := app.Config{
		AppEnv:                     "test",
		HTTPPort:                   "0",
		DatabaseURL:                databaseURL,
		RedisAddr:                  redisAddr,
		RedisDB:                    0,
		KafkaBrokers:               []string{"localhost:9092"},
		JWTAccessSecret:            "test-access-secret",
		JWTRefreshSecret:           "test-refresh-secret",
		JWTAccessTTL:               15 * time.Minute,
		JWTRefreshTTL:              24 * time.Hour,
		HoldTTL:                    10 * time.Minute,
		SeatmapCacheTTL:            30 * time.Second,
		LogLevel:                   "disabled",
		ExternalSyncCity:           "moscow",
		ExternalSyncKudaGoLocation: "msk",
		ExternalSyncTimepadCity:    "Москва",
		ExternalSyncDaysAhead:      180,
		ExternalSyncLookbackDays:   14,
		ExternalSyncMaxPages:       10,
		ExternalSyncPageSize:       100,
	}
	if mutate != nil {
		mutate(&cfg)
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

func newIntegrationDepsOnly(t *testing.T, ctx context.Context, cfg app.Config) (*app.Dependencies, func()) {
	t.Helper()
	redisAddr, cleanupRedis := startRedis(t, ctx)
	cfg.RedisAddr = redisAddr
	cfg.RedisDB = 0
	cfg.KafkaBrokers = []string{"localhost:9092"}
	deps, err := app.NewDependencies(ctx, cfg)
	require.NoError(t, err)
	cleanup := func() {
		deps.Close()
		cleanupRedis()
	}
	return deps, cleanup
}

func loginAdmin(t *testing.T, baseURL string) authResponse {
	t.Helper()
	var login authResponse
	doJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]any{
		"email":    "admin@example.com",
		"password": "Password123!",
	}, http.StatusOK, &login)
	return login
}

type eventCatalogResponse struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	BookingMode string `json:"bookingMode"`
}

type eventMapResponse struct {
	ID    string                `json:"id"`
	Venue *eventMapVenueSummary `json:"venue"`
}

type eventMapVenueSummary struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type eventDetailResponse struct {
	ID          string            `json:"id"`
	Source      string            `json:"source"`
	BookingMode string            `json:"bookingMode"`
	Sessions    []sessionResponse `json:"sessions"`
}

type sessionLayoutStateResponse struct {
	SessionID       string              `json:"sessionId"`
	LayoutSource    string              `json:"layoutSource"`
	Layout          *seatLayoutResponse `json:"layout"`
	FallbackLayout  *seatLayoutResponse `json:"fallbackLayout"`
	EffectiveLayout *seatLayoutResponse `json:"effectiveLayout"`
}

type hallLayoutStateResponse struct {
	HallID   string                  `json:"hallId"`
	Layout   *seatLayoutResponse     `json:"layout"`
	Sessions []adminSessionReference `json:"sessions"`
}

type adminSessionReference struct {
	ID string `json:"id"`
}

type layoutMutationResponse struct {
	SessionID         *string `json:"sessionId"`
	HallID            string  `json:"hallId"`
	EventID           string  `json:"eventId"`
	EventTitle        string  `json:"eventTitle"`
	BookingMode       string  `json:"bookingMode"`
	IsBookable        bool    `json:"isBookable"`
	VisibleInCatalog  bool    `json:"visibleInCatalog"`
	MaterializedSeats int     `json:"materializedSeats"`
}

func mustQueryInt(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var value int
	require.NoError(t, db.QueryRowContext(ctx, query, args...).Scan(&value))
	return value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
