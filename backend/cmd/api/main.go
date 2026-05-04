package main

import (
	"context"
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

	if err := app.NewAPI(deps).Run(ctx); err != nil {
		deps.Log.Fatal().Err(err).Msg("backend-api stopped")
	}
}
