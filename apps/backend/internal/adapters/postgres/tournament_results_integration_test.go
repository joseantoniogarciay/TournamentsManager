package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationWithdrawTournamentTeamAppliesUniformResultsAndKeepsHistory(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "withdraw-organizer@example.test", "withdraworganizer", "correct password")
	outsiderID := createVerifiedLocalAccount(t, ctx, pool, "withdraw-outsider@example.test", "withdrawoutsider", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, organizerID, tournaments.CreateInput{Name: "Liga con baja", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	started, err := service.Start(ctx, organizerID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar liga = %v", err)
	}
	withdrawn := started.Teams[0]
	match, found := leagueMatchBetweenTeams(started.Matches, withdrawn.ID, started.Teams[1].ID)
	if !found {
		t.Fatal("faltaba partido del equipo retirado")
	}
	if _, err := service.RecordResult(ctx, organizerID, created.ID, match.ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}); err != nil {
		t.Fatalf("registrar resultado previo = %v", err)
	}
	if _, err := service.WithdrawTeam(ctx, outsiderID, created.ID, withdrawn.ID); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("baja ajena = %v, se esperaba prohibida", err)
	}
	updated, err := service.WithdrawTeam(ctx, organizerID, created.ID, withdrawn.ID)
	if err != nil {
		t.Fatalf("retirar equipo = %v", err)
	}
	for _, updatedMatch := range updated.Matches {
		if updatedMatch.HomeTeamID != withdrawn.ID && updatedMatch.AwayTeamID != withdrawn.ID {
			continue
		}
		if updatedMatch.State != "completed" || updatedMatch.ResultType != tournaments.ResultAdministrative || updatedMatch.HomeScore == nil || updatedMatch.AwayScore == nil {
			t.Fatalf("partido retirado sin completar = %#v", updatedMatch)
		}
		if updatedMatch.HomeTeamID == withdrawn.ID && (*updatedMatch.HomeScore != 0 || *updatedMatch.AwayScore != 3) {
			t.Fatalf("marcador local retirado = %d-%d", *updatedMatch.HomeScore, *updatedMatch.AwayScore)
		}
		if updatedMatch.AwayTeamID == withdrawn.ID && (*updatedMatch.HomeScore != 3 || *updatedMatch.AwayScore != 0) {
			t.Fatalf("marcador visitante retirado = %d-%d", *updatedMatch.HomeScore, *updatedMatch.AwayScore)
		}
	}
	if !updated.Teams[0].Withdrawn {
		t.Fatalf("equipo retirado = %#v, se esperaba marcado", updated.Teams[0])
	}
	var historyCount, playedCount, administrativeCount int
	if err := pool.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE result_type = 'played'), count(*) FILTER (WHERE result_type = 'administrative') FROM match_result_changes WHERE match_id = $1`, match.ID).Scan(&historyCount, &playedCount, &administrativeCount); err != nil || historyCount != 2 || playedCount != 1 || administrativeCount != 1 {
		t.Fatalf("historial del resultado sustituido = total %d, jugado %d, administrativo %d, %v", historyCount, playedCount, administrativeCount, err)
	}
	if _, err := service.WithdrawTeam(ctx, organizerID, created.ID, withdrawn.ID); !errors.Is(err, tournaments.ErrTournamentWithdrawalConflict) {
		t.Fatalf("segunda baja = %v, se esperaba conflicto", err)
	}
}

func TestIntegrationBasketballRejectsTiesAndAppliesTwentyZeroWithdrawal(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "basketball-organizer@example.test", "basketballorganizer", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, organizerID, tournaments.CreateInput{
		Name: "Liga de baloncesto", Sport: tournaments.SportBasketball,
		Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}},
	})
	if err != nil || created.Sport != tournaments.SportBasketball {
		t.Fatalf("crear baloncesto = %#v, %v", created, err)
	}
	started, err := service.Start(ctx, organizerID, created.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar baloncesto = %v", err)
	}
	withdrawn := started.Teams[0]
	match, found := leagueMatchBetweenTeams(started.Matches, withdrawn.ID, started.Teams[1].ID)
	if !found {
		t.Fatal("faltaba partido de baloncesto")
	}
	if _, err := service.RecordResult(ctx, organizerID, created.ID, match.ID, tournaments.MatchResultInput{HomeScore: 80, AwayScore: 80}); !errors.Is(err, tournaments.ErrInvalidTournamentInput) {
		t.Fatalf("empate de baloncesto = %v, se esperaba validación", err)
	}
	if _, err := service.RecordResult(ctx, organizerID, created.ID, match.ID, tournaments.MatchResultInput{HomeScore: 84, AwayScore: 76}); err != nil {
		t.Fatalf("resultado de baloncesto = %v", err)
	}
	updated, err := service.WithdrawTeam(ctx, organizerID, created.ID, withdrawn.ID)
	if err != nil {
		t.Fatalf("retirar equipo de baloncesto = %v", err)
	}
	for _, updatedMatch := range updated.Matches {
		if updatedMatch.HomeTeamID != withdrawn.ID && updatedMatch.AwayTeamID != withdrawn.ID {
			continue
		}
		if updatedMatch.ResultType != tournaments.ResultAdministrative || updatedMatch.HomeScore == nil || updatedMatch.AwayScore == nil {
			t.Fatalf("resultado administrativo = %#v", updatedMatch)
		}
		if updatedMatch.HomeTeamID == withdrawn.ID && (*updatedMatch.HomeScore != 0 || *updatedMatch.AwayScore != 20) {
			t.Fatalf("retirada local = %d-%d", *updatedMatch.HomeScore, *updatedMatch.AwayScore)
		}
		if updatedMatch.AwayTeamID == withdrawn.ID && (*updatedMatch.HomeScore != 20 || *updatedMatch.AwayScore != 0) {
			t.Fatalf("retirada visitante = %d-%d", *updatedMatch.HomeScore, *updatedMatch.AwayScore)
		}
	}
	for _, standing := range updated.Standings {
		if standing.TeamID == withdrawn.ID && standing.Points != 0 {
			t.Fatalf("puntos del retirado = %d, se esperaban 0", standing.Points)
		}
	}
}

func TestIntegrationHandballUsesTwoOneZeroAndTenZeroWithdrawal(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "handball-organizer@example.test", "handballorganizer", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, organizerID, tournaments.CreateInput{
		Name: "Liga de balonmano", Sport: tournaments.SportHandball,
		Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}},
	})
	if err != nil || created.Sport != tournaments.SportHandball {
		t.Fatalf("crear balonmano = %#v, %v", created, err)
	}
	started, err := service.Start(ctx, organizerID, created.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar balonmano = %v", err)
	}
	first, found := leagueMatchBetweenTeams(started.Matches, started.Teams[0].ID, started.Teams[1].ID)
	if !found {
		t.Fatal("faltaba primer partido de balonmano")
	}
	if _, err := service.RecordResult(ctx, organizerID, created.ID, first.ID, tournaments.MatchResultInput{HomeScore: 24, AwayScore: 24}); err != nil {
		t.Fatalf("empate de balonmano = %v", err)
	}
	second, found := leagueMatchBetweenTeams(started.Matches, started.Teams[0].ID, started.Teams[2].ID)
	if !found {
		t.Fatal("faltaba segundo partido de balonmano")
	}
	winningResult := tournaments.MatchResultInput{HomeScore: 25, AwayScore: 24}
	if second.AwayTeamID == started.Teams[0].ID {
		winningResult = tournaments.MatchResultInput{HomeScore: 24, AwayScore: 25}
	}
	updated, err := service.RecordResult(ctx, organizerID, created.ID, second.ID, winningResult)
	if err != nil {
		t.Fatalf("victoria de balonmano = %v", err)
	}
	for _, standing := range updated.Standings {
		if standing.TeamID == started.Teams[0].ID && standing.Points != 3 {
			t.Fatalf("puntos de balonmano = %d, se esperaban 3", standing.Points)
		}
	}
	withdrawn := started.Teams[1]
	updated, err = service.WithdrawTeam(ctx, organizerID, created.ID, withdrawn.ID)
	if err != nil {
		t.Fatalf("retirar equipo de balonmano = %v", err)
	}
	for _, match := range updated.Matches {
		if match.HomeTeamID != withdrawn.ID && match.AwayTeamID != withdrawn.ID {
			continue
		}
		if match.ResultType != tournaments.ResultAdministrative || match.HomeScore == nil || match.AwayScore == nil {
			t.Fatalf("resultado administrativo de balonmano = %#v", match)
		}
		if match.HomeTeamID == withdrawn.ID && (*match.HomeScore != 0 || *match.AwayScore != 10) {
			t.Fatalf("retirada local de balonmano = %d-%d", *match.HomeScore, *match.AwayScore)
		}
		if match.AwayTeamID == withdrawn.ID && (*match.HomeScore != 10 || *match.AwayScore != 0) {
			t.Fatalf("retirada visitante de balonmano = %d-%d", *match.HomeScore, *match.AwayScore)
		}
	}
}
