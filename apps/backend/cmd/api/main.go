// Command api starts the backend HTTP adapter.
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

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/accounts"
	googleadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/google"
	httpadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/http"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres"
	smtpadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/smtp"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/config"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("la API no pudo iniciarse", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && args[0] == "purge-expired-accounts" {
		return purgeExpiredAccounts()
	}
	if len(args) > 0 {
		return fmt.Errorf("orden no reconocida: %q", args[0])
	}

	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("cargar configuración: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownTelemetry, err := observability.Configure(ctx, appConfig.OTELTracesEndpoint)
	if err != nil {
		return fmt.Errorf("configurar observabilidad: %w", err)
	}
	defer func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := shutdownTelemetry(shutdownContext); err != nil {
			slog.Error("no se pudo cerrar la telemetría", "error", err)
		}
	}()

	pool, err := postgres.NewPool(ctx, appConfig.DatabaseURL, observability.QueryTracer{})
	if err != nil {
		return err
	}
	defer pool.Close()
	mailer, err := smtpadapter.NewMailer(appConfig.SMTPAddr, appConfig.SMTPFrom, appConfig.SMTPUsername, appConfig.SMTPPassword, appConfig.PublicBaseURL)
	if err != nil {
		return fmt.Errorf("configurar correo: %w", err)
	}
	registrationService := registration.NewService(postgres.NewRegistrationRepository(pool), mailer, observability.PasswordVerifier{})
	accountLeagues := postgres.NewAccountLeagueRepository(pool)
	var federatedService *federated.Service
	if len(appConfig.GoogleClientIDs) > 0 {
		service := federated.NewService(postgres.NewFederatedRepository(pool), googleadapter.NewVerifier(appConfig.GoogleClientIDs))
		federatedService = &service
	}

	server := &http.Server{
		Addr:              appConfig.HTTPAddr,
		Handler:           observability.HTTPHandler(httpadapter.NewHandlerWithCookieSecurityAndTrustedProxies(registrationService, federatedService, accountLeagues, leagues.NewService(accountLeagues), appConfig.CORSAllowedOrigins, appConfig.CookieSecure, appConfig.TrustedProxyCIDRs, leagues.NewCreationService(accountLeagues))),
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

func purgeExpiredAccounts() error {
	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("cargar configuración: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pool, err := postgres.NewPool(ctx, appConfig.DatabaseURL, observability.QueryTracer{})
	if err != nil {
		return err
	}
	defer pool.Close()
	deleted, err := accounts.NewService(postgres.NewAccountLeagueRepository(pool)).PurgeExpired(ctx)
	if err != nil {
		return fmt.Errorf("purgar cuentas con baja vencida: %w", err)
	}
	slog.Info("purga de cuentas completada", "deleted_accounts", deleted)
	return nil
}
