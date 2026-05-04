package app

import (
	"context"
	"net/http"
	"time"
)

type API struct {
	server *http.Server
	deps   *Dependencies
}

func NewAPI(deps *Dependencies) *API {
	return &API{
		deps: deps,
		server: &http.Server{
			Addr:              ":" + deps.Config.HTTPPort,
			Handler:           NewRouter(deps),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (a *API) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.deps.Log.Info().Str("addr", a.server.Addr).Msg("backend-api listening")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
