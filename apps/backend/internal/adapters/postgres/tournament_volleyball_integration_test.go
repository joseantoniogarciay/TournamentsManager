package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func volleyballResult(homeWins bool, losingSets int) tournaments.MatchResultInput {
	sets := []tournaments.SetScore{}
	for i := 0; i < losingSets; i++ {
		sets = append(sets, tournaments.SetScore{HomeScore: 20, AwayScore: 25})
	}
	for i := 0; i < 3; i++ {
		target := 25
		if len(sets) == 4 {
			target = 15
		}
		sets = append(sets, tournaments.SetScore{HomeScore: target, AwayScore: 0})
	}
	if !homeWins {
		for i := range sets {
			sets[i].HomeScore, sets[i].AwayScore = sets[i].AwayScore, sets[i].HomeScore
		}
	}
	return tournaments.MatchResultInput{Sets: sets}
}

func TestIntegrationVolleyballLeagueCorrectionAndWithdrawalHistory(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "volley@example.test", "volley_owner", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Volleyball", Sport: tournaments.SportVolleyball, BestOfSets: 5, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}}})
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Matches) != 6 {
		t.Fatal("two-leg fixture missing")
	}
	match := value.Matches[0]
	if _, err = service.RecordResult(ctx, owner, value.ID, match.ID, tournaments.MatchResultInput{HomeScore: 3, AwayScore: 0}); !errors.Is(err, tournaments.ErrInvalidTournamentInput) {
		t.Fatalf("aggregate only accepted: %v", err)
	}
	value, err = service.RecordResult(ctx, owner, value.ID, match.ID, volleyballResult(true, 2))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range value.Matches {
		if m.ID == match.ID && (len(m.Sets) != 5 || *m.HomeScore != 3 || *m.AwayScore != 2) {
			t.Fatalf("five sets=%#v", m)
		}
	}
	value, err = service.RecordResult(ctx, owner, value.ID, match.ID, volleyballResult(false, 0))
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.WithdrawTeam(ctx, owner, value.ID, match.AwayTeamID)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range value.Matches {
		if m.HomeTeamID != match.AwayTeamID && m.AwayTeamID != match.AwayTeamID {
			continue
		}
		expected := tournaments.AdministrativeVolleyballSets(m.HomeTeamID != match.AwayTeamID)
		if m.ResultType != tournaments.ResultAdministrative || len(m.Sets) != 3 {
			t.Fatalf("administrative match=%#v", m)
		}
		for i := range expected {
			if m.Sets[i] != expected[i] {
				t.Fatalf("admin set=%#v", m.Sets)
			}
		}
	}
	if _, err = service.RecordResult(ctx, owner, value.ID, match.ID, volleyballResult(true, 0)); !errors.Is(err, tournaments.ErrMatchResultConflict) {
		t.Fatalf("admin overwritten: %v", err)
	}
	var sets, played, admin int
	err = pool.QueryRow(ctx, "SELECT count(*),count(*) FILTER (WHERE c.result_type='played'),count(*) FILTER (WHERE c.result_type='administrative') FROM match_result_change_sets s JOIN match_result_changes c ON c.id=s.change_id WHERE c.match_id=$1", match.ID).Scan(&sets, &played, &admin)
	if err != nil || sets != 11 || played != 8 || admin != 3 {
		t.Fatalf("history=%d played=%d admin=%d err=%v", sets, played, admin, err)
	}
}

func TestIntegrationVolleyballCompletionUsesRallyRatio(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "volley_complete@example.test", "volley_complete", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Rally tie", Sport: tournaments.SportVolleyball, BestOfSets: 5, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}}})
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, team := range value.Teams {
		ids[team.Name] = team.ID
	}
	winners := map[[2]string]string{orderedTeamPair(ids["A"], ids["B"]): ids["A"], orderedTeamPair(ids["B"], ids["C"]): ids["B"], orderedTeamPair(ids["A"], ids["C"]): ids["C"]}
	for _, m := range value.Matches {
		winner := winners[orderedTeamPair(m.HomeTeamID, m.AwayTeamID)]
		input := volleyballResult(winner == m.HomeTeamID, 0)
		if winner == ids["C"] {
			for i := range input.Sets {
				if winner == m.HomeTeamID {
					input.Sets[i].AwayScore = 20
				} else {
					input.Sets[i].HomeScore = 20
				}
			}
		}
		value, err = service.RecordResult(ctx, owner, value.ID, m.ID, input)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Each team has one win, three points and a 3/3 set ratio. A has the best rally ratio.
	value, err = service.Complete(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.ChampionTeamIDs) != 1 || value.ChampionTeamIDs[0] != ids["A"] {
		t.Fatalf("champions=%#v", value.ChampionTeamIDs)
	}
}

func TestIntegrationVolleyballMixedRepeatedTieBreakAndFinal(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "volley_mixed@example.test", "volley_mixed", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Volleyball mixed", Sport: tournaments.SportVolleyball, BestOfSets: 5, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}}})
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, team := range value.Teams {
		ids[team.Name] = team.ID
	}
	cycleWinners := map[[2]string]string{orderedTeamPair(ids["A"], ids["B"]): ids["A"], orderedTeamPair(ids["B"], ids["C"]): ids["B"], orderedTeamPair(ids["A"], ids["C"]): ids["C"]}
	for _, m := range value.Matches {
		value, err = service.RecordResult(ctx, owner, value.ID, m.ID, volleyballResult(cycleWinners[orderedTeamPair(m.HomeTeamID, m.AwayTeamID)] == m.HomeTeamID, 0))
		if err != nil {
			t.Fatal(err)
		}
	}
	value, err = service.StartElimination(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	stage := mixedStageByType(t, value, tournaments.StageTypeQualificationTieBreak)
	if len(value.TieBreakPools) != 1 {
		t.Fatalf("pools=%#v", value.TieBreakPools)
	}
	for _, m := range tiebreakMatchesForCycle(value, stage.ID, 1) {
		value, err = service.RecordResult(ctx, owner, value.ID, m.ID, volleyballResult(cycleWinners[orderedTeamPair(m.HomeTeamID, m.AwayTeamID)] == m.HomeTeamID, 0))
		if err != nil {
			t.Fatal(err)
		}
	}
	if value.TieBreakPools[0].CurrentCycle != 2 {
		t.Fatalf("cycle=%#v", value.TieBreakPools)
	}
	rank := map[string]int{ids["A"]: 0, ids["B"]: 1, ids["C"]: 2}
	for _, m := range tiebreakMatchesForCycle(value, stage.ID, 2) {
		value, err = service.RecordResult(ctx, owner, value.ID, m.ID, volleyballResult(rank[m.HomeTeamID] < rank[m.AwayTeamID], 2))
		if err != nil {
			t.Fatal(err)
		}
	}
	if value.TieBreakPools[0].State != "completed" {
		t.Fatalf("pool=%#v", value.TieBreakPools)
	}
	value, err = service.StartElimination(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	elimination := mixedStageByType(t, value, "single_elimination")
	var final tournaments.Match
	for _, m := range value.Matches {
		if m.StageID == elimination.ID {
			final = m
		}
	}
	if final.HomeTeamID == ids["C"] || final.AwayTeamID == ids["C"] || final.ID == "" {
		t.Fatalf("final=%#v", final)
	}
	value, err = service.RecordResult(ctx, owner, value.ID, final.ID, volleyballResult(true, 2))
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Complete(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.ChampionTeamIDs) != 1 || value.ChampionTeamIDs[0] != final.HomeTeamID {
		t.Fatal("final winner not retained")
	}
}

func TestIntegrationVolleyballGroupsWithdrawalAndKnockout(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "volley_groups@example.test", "volley_groups", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}, {Name: "F"}}
	value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Groups", Sport: tournaments.SportVolleyball, BestOfSets: 5, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 2, LeagueStructure: tournaments.LeagueStructureGroups, GroupCount: 2, QualifiersPerGroup: 1})
	if err != nil {
		t.Fatal(err)
	}
	withdrawn := value.Teams[0].ID
	value, err = service.WithdrawTeam(ctx, owner, value.ID, withdrawn)
	if err != nil {
		t.Fatal(err)
	}
	rank := map[string]int{}
	for i, team := range value.Teams {
		rank[team.ID] = i
	}
	for _, m := range value.Matches {
		if m.State == "completed" {
			continue
		}
		value, err = service.RecordResult(ctx, owner, value.ID, m.ID, volleyballResult(rank[m.HomeTeamID] < rank[m.AwayTeamID], 0))
		if err != nil {
			t.Fatal(err)
		}
	}
	value, err = service.StartElimination(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	stage := mixedStageByType(t, value, "single_elimination")
	var final tournaments.Match
	for _, m := range value.Matches {
		if m.StageID == stage.ID {
			final = m
		}
	}
	if final.ID == "" || final.HomeTeamID == withdrawn || final.AwayTeamID == withdrawn {
		t.Fatalf("final=%#v", final)
	}
	if _, err = service.WithdrawTeam(ctx, owner, value.ID, final.HomeTeamID); !errors.Is(err, tournaments.ErrTournamentWithdrawalConflict) {
		t.Fatalf("knockout withdrawal=%v", err)
	}
	value, err = service.RecordResult(ctx, owner, value.ID, final.ID, volleyballResult(false, 1))
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Complete(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.ChampionTeamIDs) != 1 || value.ChampionTeamIDs[0] != final.AwayTeamID {
		t.Fatal("group final winner incorrect")
	}
}
