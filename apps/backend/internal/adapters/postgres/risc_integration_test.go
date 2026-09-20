package postgres

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestRISCRevocationIsAtomicAndIdempotent(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	accountID := createVerifiedLocalAccount(t, ctx, pool, "risc@example.test", "risc_person", "correct horse battery staple")
	if _, err := pool.Exec(ctx, `INSERT INTO external_identities (account_id, provider, issuer, subject) VALUES ($1, 'google', $2, $3)`, accountID, federated.GoogleIssuer, "google-subject"); err != nil {
		t.Fatalf("crear identidad Google: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at) VALUES ($1, $2, now() + interval '1 day', now() + interval '1 day')`, accountID, sessionHash("first-session")); err != nil {
		t.Fatalf("crear primera sesión: %v", err)
	}
	repository := NewFederatedRepository(pool)
	event := federated.RISCEvent{ID: "risc-event-id", Issuer: federated.GoogleIssuer, Subject: "google-subject", Type: federated.RISCSessionsRevoked}
	if err := repository.RevokeSessionsForGoogleIdentity(ctx, event); err != nil {
		t.Fatalf("procesar primer evento RISC: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at) VALUES ($1, $2, now() + interval '1 day', now() + interval '1 day')`, accountID, sessionHash("second-session")); err != nil {
		t.Fatalf("crear segunda sesión: %v", err)
	}
	if err := repository.RevokeSessionsForGoogleIdentity(ctx, event); err != nil {
		t.Fatalf("repetir evento RISC: %v", err)
	}
	var active int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE account_id = $1 AND revoked_at IS NULL`, accountID).Scan(&active); err != nil {
		t.Fatalf("contar sesiones activas: %v", err)
	}
	if active != 1 {
		t.Fatalf("sesiones activas = %d, want 1", active)
	}
}

func leagueMatchByID(matches []tournaments.Match, id string) (tournaments.Match, bool) {
	for _, match := range matches {
		if match.ID == id {
			return match, true
		}
	}
	return tournaments.Match{}, false
}

func leagueMatchBetweenTeams(matches []tournaments.Match, firstTeamID, secondTeamID string) (tournaments.Match, bool) {
	for _, match := range matches {
		if (match.HomeTeamID == firstTeamID && match.AwayTeamID == secondTeamID) ||
			(match.HomeTeamID == secondTeamID && match.AwayTeamID == firstTeamID) {
			return match, true
		}
	}
	return tournaments.Match{}, false
}

func recordWin(t *testing.T, ctx context.Context, service tournaments.CreationService, accountID, leagueID string, match tournaments.Match, winnerTeamID string, winningScore, losingScore int) {
	t.Helper()
	homeScore, awayScore := losingScore, winningScore
	if match.HomeTeamID == winnerTeamID {
		homeScore, awayScore = winningScore, losingScore
	}
	if _, err := service.RecordResult(ctx, accountID, leagueID, match.ID, tournaments.MatchResultInput{HomeScore: homeScore, AwayScore: awayScore}); err != nil {
		t.Fatalf("registrar resultado de %s = %v", match.ID, err)
	}
}

// Esta prueba usa una base efímera preparada por el comando de integración.
