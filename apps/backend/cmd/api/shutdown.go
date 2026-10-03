package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// shutdownHTTP preserves active requests until the deadline. Close cancels
// remaining HTTP requests on timeout before the caller releases dependencies.
func shutdownHTTP(ctx context.Context, server *http.Server) error {
	slog.Info("drenaje HTTP iniciado")
	if err := server.Shutdown(ctx); err != nil {
		slog.Warn("drenaje HTTP agotado; cerrando conexiones restantes")
		return errors.Join(err, server.Close())
	}
	slog.Info("drenaje HTTP completado")
	return nil
}
