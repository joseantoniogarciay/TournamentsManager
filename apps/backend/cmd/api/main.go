// Command api inicia el adaptador HTTP del backend.
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

	googleadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/google"
	httpadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/http"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres"
	smtpadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/smtp"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/config"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("la API no pudo iniciarse", "error", err)
		os.Exit(1)
	}
}

func run() error {
	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("cargar configuración: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, appConfig.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	mailer, err := smtpadapter.NewMailer(appConfig.SMTPAddr, appConfig.SMTPFrom, appConfig.PublicBaseURL)
	if err != nil {
		return fmt.Errorf("configurar correo: %w", err)
	}
	registrationService := registration.NewService(postgres.NewRegistrationRepository(pool), mailer)
	accountLeagues := postgres.NewAccountLeagueRepository(pool)
	var federatedService *federated.Service
	if len(appConfig.GoogleClientIDs) > 0 {
		service := federated.NewService(postgres.NewFederatedRepository(pool), googleadapter.NewVerifier(appConfig.GoogleClientIDs))
		federatedService = &service
	}

	server := &http.Server{
		Addr:              appConfig.HTTPAddr,
		Handler:           httpadapter.NewHandler(registrationService, federatedService, accountLeagues, leagues.NewService(accountLeagues), appConfig.CORSAllowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	slog.Info("API iniciada", "address", appConfig.HTTPAddr)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("escuchar HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("detener HTTP: %w", err)
		}
		return nil
	}
}
