// Command league-preview-renderer injects sharing metadata into the static web
// shell for one canonical public league URL. Expo continues to export the rest
// of the application statically.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leaguepreview"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("el renderer de previews no pudo iniciarse", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config := leaguepreview.Config{
		WebRoot:       os.Getenv("LEAGUE_PREVIEW_WEB_ROOT"),
		PublicBaseURL: os.Getenv("LEAGUE_PREVIEW_PUBLIC_BASE_URL"),
		PublicHost:    os.Getenv("LEAGUE_PREVIEW_PUBLIC_HOST"),
		APIBaseURL:    os.Getenv("LEAGUE_PREVIEW_API_BASE_URL"),
		APIHost:       os.Getenv("LEAGUE_PREVIEW_API_HOST"),
	}
	handler, err := leaguepreview.NewHandler(config, nil)
	if err != nil {
		return fmt.Errorf("configurar renderer de previews: %w", err)
	}
	address := os.Getenv("LEAGUE_PREVIEW_LISTEN_ADDRESS")
	if address == "" {
		return errors.New("LEAGUE_PREVIEW_LISTEN_ADDRESS es obligatorio")
	}
	server := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	slog.Info("renderer de previews iniciado", "address", address)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("escuchar renderer de previews: %w", err)
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownContext)
	}
}
