package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationBadmintonProfilesAndHistory(t *testing.T) {
	for _, points := range []int{15, 21} {
		t.Run(fmt.Sprint(points), func(t *testing.T) {
			pool := integrationPool(t)
			ctx := context.Background()
			owner := createVerifiedLocalAccount(ctx, t, pool, "badminton@example.test", "badminton_owner", "correct password")
			service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
			value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Badminton", Sport: tournaments.SportBadminton, BestOfSets: 3, PointsPerGame: points, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}}})
			if err != nil {
				t.Fatal(err)
			}
			read, err := service.GetPublic(ctx, value.ID)
			if err != nil || read.PointsPerGame != points || read.BestOfSets != 3 {
				t.Fatalf("configuration read=%+v %v", read, err)
			}
			for _, format := range []string{"league", "league_then_single_elimination"} {
				if _, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: format, RoundRobinLegs: 1}); !errors.Is(err, tournaments.ErrInvalidTournamentInput) {
					t.Fatalf("format=%s: %v", format, err)
				}
			}
			value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "single_elimination"})
			if err != nil {
				t.Fatal(err)
			}
			first, second, final := value.Matches[0], value.Matches[1], value.Matches[2]
			scoreCap := 30
			if points == 15 {
				scoreCap = 21
			}
			sets := []tournaments.SetScore{{HomeScore: points, AwayScore: points - 2}, {HomeScore: scoreCap - 1, AwayScore: scoreCap}, {HomeScore: scoreCap, AwayScore: scoreCap - 1}}
			value, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Sets: sets})
			if err != nil || value.PointsPerGame != points || len(value.Matches[0].Sets) != 3 || value.Matches[2].HomeTeamID != first.HomeTeamID {
				t.Fatalf("record=%+v %v", value, err)
			}
			invalid := []tournaments.SetScore{{HomeScore: scoreCap + 1, AwayScore: scoreCap - 1}, {HomeScore: points, AwayScore: 0}}
			if _, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Sets: invalid}); !errors.Is(err, tournaments.ErrInvalidBracketResult) {
				t.Fatalf("above cap accepted: %v", err)
			}
			corrected := []tournaments.SetScore{{HomeScore: 0, AwayScore: points}, {HomeScore: points - 2, AwayScore: points}}
			value, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Sets: corrected})
			if err != nil || len(value.Matches[0].Sets) != 2 || value.Matches[2].HomeTeamID != first.AwayTeamID {
				t.Fatalf("correction=%+v %v", value, err)
			}
			var history int
			if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_change_sets s JOIN match_result_changes c ON c.id=s.change_id WHERE c.match_id=$1`, first.ID).Scan(&history); err != nil || history != 5 {
				t.Fatalf("history=%d %v", history, err)
			}
			if _, err = service.RecordResult(ctx, owner, value.ID, second.ID, tournaments.MatchResultInput{Sets: corrected}); err != nil {
				t.Fatal(err)
			}
			if _, err = service.RecordResult(ctx, owner, value.ID, final.ID, tournaments.MatchResultInput{Sets: corrected}); err != nil {
				t.Fatal(err)
			}
			if _, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Sets: sets}); !errors.Is(err, tournaments.ErrBracketResultDependency) {
				t.Fatalf("dependency=%v", err)
			}
			value, err = service.Complete(ctx, owner, value.ID)
			if err != nil || value.PointsPerGame != points || len(value.ChampionTeamIDs) != 1 {
				t.Fatalf("completion=%+v %v", value, err)
			}
			for _, query := range []string{
				`INSERT INTO tournaments (organizer_account_id,name,sport,best_of_sets,published_at) VALUES ($1,'Missing profile','badminton',3,now())`,
				`INSERT INTO tournaments (organizer_account_id,name,sport,best_of_sets,points_per_game,published_at) VALUES ($1,'Invalid profile','badminton',3,11,now())`,
				`INSERT INTO tournaments (organizer_account_id,name,sport,best_of_sets,points_per_game,published_at) VALUES ($1,'Foreign profile','tennis',3,21,now())`,
			} {
				if _, err = pool.Exec(ctx, query, owner); err == nil {
					t.Fatal("SQL accepted invalid configuration")
				}
			}
		})
	}
}
