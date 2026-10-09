package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func incidentCreation(sport tournaments.Sport) tournaments.CreateInput {
	input := tournaments.CreateInput{Name: "Incidents", Sport: sport, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}}}
	if tournaments.SetSport(sport) {
		input.BestOfSets = 3
	}
	if sport == tournaments.SportVolleyball {
		input.BestOfSets = 5
	}
	if sport == tournaments.SportBadminton {
		input.PointsPerGame = 21
	}
	return input
}

func TestIntegrationIncidentKnockoutAcrossSports(t *testing.T) {
	for _, sport := range []tournaments.Sport{tournaments.SportFootball, tournaments.SportBasketball, tournaments.SportHandball, tournaments.SportVolleyball, tournaments.SportTennis, tournaments.SportPadel, tournaments.SportTableTennis, tournaments.SportBadminton} {
		t.Run(string(sport), func(t *testing.T) {
			pool := integrationPool(t)
			ctx := context.Background()
			owner := createVerifiedLocalAccount(ctx, t, pool, "incident@example.test", "incident_owner", "password123")
			service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
			value, err := service.Create(ctx, owner, incidentCreation(sport))
			if err != nil {
				t.Fatal(err)
			}
			value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "single_elimination"})
			if err != nil {
				t.Fatal(err)
			}
			first, second, final := value.Matches[0], value.Matches[1], value.Matches[2]
			incident := &tournaments.MatchIncident{Type: tournaments.ResultRetirement, Side: "home"}
			if tournaments.SetSport(sport) {
				incident.PartialSets = []tournaments.SetScore{{HomeScore: 2, AwayScore: 1}}
			} else {
				incident.PartialHomeScore = intPointer(52)
				incident.PartialAwayScore = intPointer(48)
			}
			value, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Incident: incident})
			if err != nil {
				t.Fatal(err)
			}
			if value.Matches[0].Incident == nil || value.Matches[0].ResultType != tournaments.ResultRetirement || value.Matches[2].HomeTeamID != first.AwayTeamID {
				t.Fatalf("lost incident/winner: %+v", value.Matches)
			}
			value, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Incident: &tournaments.MatchIncident{Type: tournaments.ResultNoShow, Side: "away"}})
			if err != nil {
				t.Fatal(err)
			}
			if value.Matches[2].HomeTeamID != first.HomeTeamID {
				t.Fatal("corrected winner missing")
			}
			// Recording another match must not overwrite the existing incident metadata.
			value, err = service.RecordResult(ctx, owner, value.ID, second.ID, tournaments.MatchResultInput{Incident: &tournaments.MatchIncident{Type: tournaments.ResultNoShow, Side: "home"}})
			if err != nil {
				t.Fatal(err)
			}
			if value.Matches[0].ResultType != tournaments.ResultNoShow || value.Matches[0].Incident.Side != "away" {
				t.Fatal("another result overwrote metadata")
			}
			value, err = service.RecordResult(ctx, owner, value.ID, final.ID, tournaments.MatchResultInput{Incident: &tournaments.MatchIncident{Type: tournaments.ResultNoShow, Side: "away"}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = service.RecordResult(ctx, owner, value.ID, first.ID, tournaments.MatchResultInput{Incident: incident}); !errors.Is(err, tournaments.ErrBracketResultDependency) {
				t.Fatalf("dependency %v", err)
			}
			var count int
			if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1 AND previous_result_type='retirement' AND previous_incident->>'type'='retirement' AND incident->>'type'='no_show'`, first.ID).Scan(&count); err != nil || count != 1 {
				t.Fatalf("history=%d %v", count, err)
			}
			value, err = service.Complete(ctx, owner, value.ID)
			if err != nil || len(value.ChampionTeamIDs) != 1 || value.ChampionTeamIDs[0] != first.HomeTeamID {
				t.Fatalf("champion %+v %v", value.ChampionTeamIDs, err)
			}
		})
	}
}

func TestIntegrationLeagueIncidentCorrectionAndWithdrawal(t *testing.T) {
	for _, sport := range []tournaments.Sport{tournaments.SportFootball, tournaments.SportBasketball, tournaments.SportHandball, tournaments.SportVolleyball} {
		t.Run(string(sport), func(t *testing.T) {
			pool := integrationPool(t)
			ctx := context.Background()
			owner := createVerifiedLocalAccount(ctx, t, pool, "league-incident@example.test", "incident_owner", "password123")
			outsider := createVerifiedLocalAccount(ctx, t, pool, "outsider-incident@example.test", "incident_other", "password123")
			service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
			value, err := service.Create(ctx, owner, incidentCreation(sport))
			if err != nil {
				t.Fatal(err)
			}
			value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
			if err != nil {
				t.Fatal(err)
			}
			match := value.Matches[0]
			input := tournaments.MatchResultInput{Incident: &tournaments.MatchIncident{Type: tournaments.ResultRetirement, Side: "home"}}
			if tournaments.SetSport(sport) {
				input.Incident.PartialSets = []tournaments.SetScore{{HomeScore: 25, AwayScore: 20}, {HomeScore: 7, AwayScore: 4}}
			} else {
				input.Incident.PartialHomeScore = intPointer(52)
				input.Incident.PartialAwayScore = intPointer(48)
			}
			if _, err = service.RecordResult(ctx, outsider, value.ID, match.ID, input); !errors.Is(err, tournaments.ErrMatchResultForbidden) {
				t.Fatalf("permissions %v", err)
			}
			value, err = service.RecordResult(ctx, owner, value.ID, match.ID, input)
			if err != nil {
				t.Fatal(err)
			}
			completed := 0
			for _, m := range value.Matches {
				if m.State == "completed" {
					completed++
				}
			}
			if completed != 1 {
				t.Fatal("incident modified other matches")
			}
			for _, standing := range value.Standings {
				if standing.TeamID == match.HomeTeamID && (standing.Won != 0 || standing.ScoreFor != 0) {
					t.Fatal("partial influenced standings")
				}
			}
			played := tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}
			if sport == tournaments.SportVolleyball {
				played = volleyballResult(true, 1)
			}
			value, err = service.RecordResult(ctx, owner, value.ID, match.ID, played)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range value.Matches {
				if m.ID == match.ID && (m.Incident != nil || m.ResultType != tournaments.ResultPlayed) {
					t.Fatal("incident metadata retained on played correction")
				}
			}
			var count int
			if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1 AND result_type='played' AND incident IS NULL AND previous_incident->>'type'='retirement'`, match.ID).Scan(&count); err != nil || count != 1 {
				t.Fatalf("history %d %v", count, err)
			}
			value, err = service.RecordResult(ctx, owner, value.ID, match.ID, input)
			if err != nil {
				t.Fatal(err)
			}
			value, err = service.WithdrawTeam(ctx, owner, value.ID, match.HomeTeamID)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range value.Matches {
				if m.ID == match.ID && (m.Incident != nil || m.ResultType != tournaments.ResultAdministrative) {
					t.Fatal("withdrawal incident metadata not cleared")
				}
			}
			if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1 AND result_type='administrative' AND previous_incident->>'type'='retirement'`, match.ID).Scan(&count); err != nil || count != 1 {
				t.Fatalf("withdrawal history %d %v", count, err)
			}
			if _, err = service.RecordResult(ctx, owner, value.ID, match.ID, input); !errors.Is(err, tournaments.ErrMatchResultConflict) {
				t.Fatalf("withdrawn result overwritten: %v", err)
			}
		})
	}
}
