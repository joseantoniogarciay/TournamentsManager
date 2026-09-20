package postgres

import (
	"context"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type integrationMailer struct{ passwordResetToken string }

type integrationGoogleVerifier struct{ identity federated.Identity }

func (v *integrationGoogleVerifier) Verify(context.Context, string) (federated.Identity, error) {
	return v.identity, nil
}

func (*integrationMailer) SendVerification(context.Context, string, registration.Locale, string) error {
	return nil
}

func (m *integrationMailer) SendPasswordReset(_ context.Context, _ string, _ registration.Locale, token string) error {
	m.passwordResetToken = token
	return nil
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TM_INTEGRATION_DATABASE_URL")
	if databaseURL == "" || os.Getenv("TM_RUN_INTEGRATION") != "1" {
		t.Skip("la integración PostgreSQL no está activada")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("conectar PostgreSQL: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(context.Background(), "TRUNCATE accounts CASCADE"); err != nil {
		t.Fatalf("limpiar base: %v", err)
	}
	return pool
}

func createVerifiedLocalAccount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email, username, password string) string {
	t.Helper()
	passwordHash, err := registration.HashPassword(password)
	if err != nil {
		t.Fatalf("crear hash de contraseña: %v", err)
	}
	var accountID string
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, verified_at) VALUES ($1, 'es', 'verified', $2, now()) RETURNING id::text`, email, username).Scan(&accountID); err != nil {
		t.Fatalf("crear cuenta: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO local_credentials (account_id, password_hash) VALUES ($1, $2)`, accountID, passwordHash); err != nil {
		t.Fatalf("crear credencial: %v", err)
	}
	return accountID
}

func sessionHash(token string) []byte {
	hash := sha256.Sum256([]byte("session:" + token))
	return hash[:]
}
