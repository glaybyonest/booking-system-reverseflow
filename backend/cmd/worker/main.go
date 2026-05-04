package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"reserveflow/backend/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := app.LoadConfig()
	if err != nil {
		panic(err)
	}
	deps, err := app.NewDependencies(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer deps.Close()

	if err := app.RunWorker(ctx, deps); err != nil && !errors.Is(err, context.Canceled) {
		deps.Log.Fatal().Err(err).Msg("backend-worker stopped")
	}
}
