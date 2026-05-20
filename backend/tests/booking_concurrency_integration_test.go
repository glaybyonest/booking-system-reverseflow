//go:build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"reserveflow/backend/internal/infrastructure/db"
	bookingsapp "reserveflow/backend/internal/modules/bookings/application"
	bookingsrepo "reserveflow/backend/internal/modules/bookings/repository"
)

func TestConcurrentHoldAllowsOnlyOneWinner(t *testing.T) {
	ctx := context.Background()
	databaseURL, cleanup := startPostgres(t, ctx)
	defer cleanup()

	pool, err := db.NewPostgresPool(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	applyIntegrationMigrations(t, ctx, databaseURL)
	applySQL(t, ctx, databaseURL, filepath.Join("..", "seeds", "dev-users.sql"))

	const userID = "10000000-0000-0000-0000-000000000001"
	fixture := seedBookableKudaGoEvent(t, ctx, pool, "Concurrency concert")
	sessionID := fixture.SessionID
	seatID := fixture.SeatIDs[0]

	service := bookingsapp.NewService(bookingsrepo.NewPostgresRepository(pool), nil, 10*time.Minute, zerolog.Nop())

	var wg sync.WaitGroup
	results := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.HoldSeats(ctx, userID, sessionID, []string{seatID})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			conflicts++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 19, conflicts)

	var pendingCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT count(*) FROM bookings WHERE session_id = $1 AND status = 'pending'
	`, sessionID).Scan(&pendingCount))
	require.Equal(t, 1, pendingCount)

	var status string
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT status FROM session_seats WHERE session_id = $1 AND seat_id = $2
	`, sessionID, seatID).Scan(&status))
	require.Equal(t, "held", status)
}

func startPostgres(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "reserveflow",
			"POSTGRES_USER":     "reserveflow",
			"POSTGRES_PASSWORD": "reserveflow",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)
	cleanup := func() {
		require.NoError(t, container.Terminate(ctx))
	}
	return fmt.Sprintf("postgres://reserveflow:reserveflow@%s:%s/reserveflow?sslmode=disable", host, port.Port()), cleanup
}

func applySQL(t *testing.T, ctx context.Context, databaseURL, path string) {
	t.Helper()
	pool, err := db.NewPostgresPool(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(content))
	require.NoError(t, err)
}
