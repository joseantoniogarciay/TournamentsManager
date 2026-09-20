package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationTournamentCompletionPersistsCoChampions(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "completion@example.test", "completion", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, accountID, tournaments.CreateInput{Name: "Liga empate", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	started, err := service.Start(ctx, accountID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar liga = %v", err)
	}
	if _, err := service.Complete(ctx, accountID, created.ID); !errors.Is(err, tournaments.ErrTournamentCompletionConflict) {
		t.Fatalf("finalizar con pendiente = %v, se esperaba %v", err, tournaments.ErrTournamentCompletionConflict)
	}
	pending, err := service.GetPublic(ctx, created.ID)
	if err != nil || pending.State != "in_progress" || len(pending.ChampionTeamIDs) != 0 {
		t.Fatalf("cierre rechazado dejó la liga = %#v, %v; se esperaba in_progress sin campeones", pending, err)
	}
	if _, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 1}); err != nil {
		t.Fatalf("registrar empate = %v", err)
	}
	completed, err := service.Complete(ctx, accountID, created.ID)
	if err != nil {
		t.Fatalf("finalizar liga = %v", err)
	}
	if completed.State != "completed" || len(completed.ChampionTeamIDs) != 2 {
		t.Fatalf("liga finalizada = %#v; se esperaban dos co-campeones", completed)
	}
	var persistedChampions int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournament_champions WHERE tournament_id = $1`, created.ID).Scan(&persistedChampions); err != nil || persistedChampions != 2 {
		t.Fatalf("co-campeones persistidos = %d, %v; se esperaban dos", persistedChampions, err)
	}
	if _, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}); !errors.Is(err, tournaments.ErrMatchResultConflict) {
		t.Fatalf("corregir liga finalizada = %v, se esperaba %v", err, tournaments.ErrMatchResultConflict)
	}
}

func TestIntegrationConcurrentTournamentCompletionAllowsOneTransition(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "concurrent-completion@example.test", "concurrent_completion", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, accountID, tournaments.CreateInput{Name: "Liga cierre simultáneo", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	started, err := service.Start(ctx, accountID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar liga = %v", err)
	}
	if _, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}); err != nil {
		t.Fatalf("registrar resultado = %v", err)
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Go(func() {
			<-start
			_, err := service.Complete(ctx, accountID, created.ID)
			errs <- err
		})
	}
	close(start)
	group.Wait()
	close(errs)

	successes, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, tournaments.ErrTournamentCompletionConflict):
			conflicts++
		default:
			t.Fatalf("finalización simultánea = %v; se esperaba éxito o conflicto", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("finalizaciones simultáneas: éxitos/conflictos = %d/%d; se esperaba 1/1", successes, conflicts)
	}
	var state string
	var champions int
	if err := pool.QueryRow(ctx, `SELECT state FROM tournaments WHERE id = $1`, created.ID).Scan(&state); err != nil {
		t.Fatalf("consultar estado final = %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournament_champions WHERE tournament_id = $1`, created.ID).Scan(&champions); err != nil {
		t.Fatalf("contar campeones finales = %v", err)
	}
	if state != "completed" || champions != 1 {
		t.Fatalf("estado/campeones tras cierre simultáneo = %q/%d; se esperaba completed/1", state, champions)
	}
}

func TestIntegrationTournamentStandingsReadPersistedResults(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "standings@example.test", "standings", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, accountID, tournaments.CreateInput{Name: "Liga clasificación", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	started, err := service.Start(ctx, accountID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar liga = %v", err)
	}
	azules, rojos, verdes := started.Teams[0], started.Teams[1], started.Teams[2]
	azulesRojos, found := leagueMatchBetweenTeams(started.Matches, azules.ID, rojos.ID)
	if !found {
		t.Fatal("no se encontró el partido Azules-Rojos")
	}
	azulesVerdes, found := leagueMatchBetweenTeams(started.Matches, azules.ID, verdes.ID)
	if !found {
		t.Fatal("no se encontró el partido Azules-Verdes")
	}
	rojosVerdes, found := leagueMatchBetweenTeams(started.Matches, rojos.ID, verdes.ID)
	if !found {
		t.Fatal("no se encontró el partido Rojos-Verdes")
	}
	recordWin(t, ctx, service, accountID, created.ID, azulesRojos, azules.ID, 2, 0)
	recordWin(t, ctx, service, accountID, created.ID, azulesVerdes, verdes.ID, 1, 0)
	recordWin(t, ctx, service, accountID, created.ID, rojosVerdes, rojos.ID, 3, 0)

	league, err := service.GetPublic(ctx, created.ID)
	if err != nil {
		t.Fatalf("consultar clasificación = %v", err)
	}
	if len(league.Standings) != 3 || league.Standings[0].TeamID != rojos.ID || league.Standings[0].Position != 1 || league.Standings[1].TeamID != azules.ID || league.Standings[1].Position != 2 || league.Standings[2].TeamID != verdes.ID || league.Standings[2].Position != 3 {
		t.Fatalf("clasificación = %#v; se esperaba Rojos, Azules, Verdes", league.Standings)
	}
}
