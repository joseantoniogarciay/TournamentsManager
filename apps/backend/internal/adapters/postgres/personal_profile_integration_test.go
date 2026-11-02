package postgres

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationPersonalProfilePrivacyTotalsAndCorrections(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repo := NewAccountTournamentRepository(pool)
	owner := createVerifiedLocalAccount(ctx, t, pool, "profile-owner@example.test", "profileowner", "correct password")
	other := createVerifiedLocalAccount(ctx, t, pool, "profile-other@example.test", "profileother", "correct password")
	create := tournaments.NewCreationService(repo)
	browse := tournaments.NewBrowsingService(repo)
	value, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "Profile 100%", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Mine"}, {Name: "Two"}, {Name: "Three"}}})
	if err != nil {
		t.Fatal(err)
	}
	value, err = create.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	own := value.Teams[0].ID
	var ownMatches []tournaments.Match
	for _, m := range value.Matches {
		if m.HomeTeamID == own || m.AwayTeamID == own {
			ownMatches = append(ownMatches, m)
		}
	}
	if len(ownMatches) != 2 {
		t.Fatal(ownMatches)
	}
	win := tournaments.MatchResultInput{HomeScore: 2, AwayScore: 0}
	if ownMatches[0].AwayTeamID == own {
		win.HomeScore, win.AwayScore = 0, 2
	}
	if _, err = create.RecordResult(ctx, owner, value.ID, ownMatches[0].ID, win); err != nil {
		t.Fatal(err)
	}
	if _, err = create.RecordResult(ctx, owner, value.ID, ownMatches[1].ID, tournaments.MatchResultInput{HomeScore: 0, AwayScore: 0}); err != nil {
		t.Fatal(err)
	}
	p, err := browse.PersonalProfile(ctx, owner, "", "", "", 1)
	if err != nil || len(p.Items) != 1 || p.NextCursor == "" || len(p.Statistics) != 1 {
		t.Fatalf("profile %+v %v", p, err)
	}
	stat := p.Statistics[0]
	if stat.Played != 2 || stat.Wins != 1 || stat.Draws != 1 || stat.Losses != 0 || stat.WinPercentage == nil || *stat.WinPercentage != 50 {
		t.Fatalf("totals %+v", stat)
	}
	next, err := browse.PersonalProfile(ctx, owner, "", "", p.NextCursor, 1)
	if err != nil || len(next.Items) != 1 || next.Items[0].Match.ID == p.Items[0].Match.ID || next.Statistics[0].Played != 2 {
		t.Fatalf("page %+v %v", next, err)
	}
	if err = create.AssignAdministrator(ctx, owner, value.ID, "profileother"); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Follow(ctx, other, value.ID); err != nil {
		t.Fatal(err)
	}
	empty, err := browse.PersonalProfile(ctx, other, "", "", p.NextCursor, 20)
	if err != nil || len(empty.Items) != 0 || len(empty.Statistics) != 0 {
		t.Fatalf("privacy %+v %v", empty, err)
	}
	for _, filter := range []struct {
		sport, name string
		want        int
	}{{"football", "%", 2}, {"tennis", "", 0}, {"", "_", 0}, {"", "pRoFiLe", 2}} {
		filtered, e := browse.PersonalProfile(ctx, owner, filter.sport, filter.name, "", 20)
		if e != nil || len(filtered.Items) != filter.want {
			t.Fatalf("filter %+v %+v %v", filter, filtered, e)
		}
	}
	win.HomeScore, win.AwayScore = win.AwayScore, win.HomeScore
	if _, err = create.RecordResult(ctx, owner, value.ID, ownMatches[0].ID, win); err != nil {
		t.Fatal(err)
	}
	p, err = browse.PersonalProfile(ctx, owner, "", "", "", 20)
	if err != nil || p.Statistics[0].Losses != 1 || p.Statistics[0].Wins != 0 {
		t.Fatalf("correction %+v %v", p, err)
	}
	if _, err = create.WithdrawTeam(ctx, owner, value.ID, own); err != nil {
		t.Fatal(err)
	}
	p, err = browse.PersonalProfile(ctx, owner, "", "", "", 20)
	if err != nil || len(p.Items) != 2 || p.Statistics[0].Played != 0 || p.Statistics[0].WinPercentage != nil {
		t.Fatalf("administrative %+v %v", p, err)
	}
}

func TestIntegrationPersonalProfileByesShootoutsAndIncidents(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repo := NewAccountTournamentRepository(pool)
	owner := createVerifiedLocalAccount(ctx, t, pool, "profile-cup@example.test", "profilecup", "correct password")
	other := createVerifiedLocalAccount(ctx, t, pool, "profile-bye@example.test", "profilebye", "correct password")
	create := tournaments.NewCreationService(repo)
	browse := tournaments.NewBrowsingService(repo)
	value, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "Cup", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Mine"}, {Name: "Two"}, {Name: "Three"}}})
	if err != nil {
		t.Fatal(err)
	}
	invite, e := create.CreateTeamInvitation(ctx, owner, value.ID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = create.JoinTeamInvitation(ctx, other, invite, tournaments.TeamInput{Name: "Two"}); e != nil {
		t.Fatal(e)
	}
	value, err = create.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	own := value.Teams[0].ID
	// Play all currently resolved encounters, including the final after advancement.
	for cycle := 0; cycle < 3; cycle++ {
		for _, m := range value.Matches {
			if m.State != "pending" || m.HomeTeamID == "" || m.AwayTeamID == "" {
				continue
			}
			hp, ap := 5, 4
			input := tournaments.MatchResultInput{HomeScore: 1, AwayScore: 1, HomePenalties: &hp, AwayPenalties: &ap}
			if m.AwayTeamID == own {
				hp, ap = 4, 5
			}
			value, err = create.RecordResult(ctx, owner, value.ID, m.ID, input)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	p, err := browse.PersonalProfile(ctx, owner, "", "", "", 20)
	if err != nil || len(p.Statistics) != 1 || p.Statistics[0].Wins != p.Statistics[0].Played || p.Statistics[0].Draws != 0 || p.Statistics[0].DrawsAllowed {
		t.Fatalf("shootout %+v %v", p, err)
	}
	byeProfile, e := browse.PersonalProfile(ctx, other, "", "", "", 20)
	if e != nil || len(byeProfile.Statistics) != 1 || byeProfile.Statistics[0].Played != 1 || byeProfile.Statistics[0].Losses != 1 {
		t.Fatalf("away %+v %v", byeProfile, e)
	}
	byes := 0
	for _, item := range byeProfile.Items {
		if item.Match.State == "bye" {
			byes++
			if item.Match.HomeTeamID != "" && item.Match.AwayTeamID != "" {
				t.Fatal("invalid bye")
			}
		}
	}
	if byes != 1 || len(byeProfile.Items) != byeProfile.Statistics[0].Played+1 {
		t.Fatalf("byes %+v", byeProfile)
	}
	for _, m := range value.Matches {
		if m.State == "completed" && m.RoundNumber == 2 && (m.HomeTeamID == own || m.AwayTeamID == own) {
			_, err = create.RecordResult(ctx, owner, value.ID, m.ID, tournaments.MatchResultInput{Incident: &tournaments.MatchIncident{Type: tournaments.ResultNoShow, Side: "home"}})
			if err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	p, err = browse.PersonalProfile(ctx, owner, "", "", "", 20)
	if err != nil || p.Statistics[0].Played != 1 || len(p.Items) != 2 {
		t.Fatalf("incident %+v %v", p, err)
	}
}

func TestIntegrationPersonalProfileSeparatesAllSports(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repo := NewAccountTournamentRepository(pool)
	owner := createVerifiedLocalAccount(ctx, t, pool, "profile-sports@example.test", "profilesports", "correct password")
	create := tournaments.NewCreationService(repo)
	browse := tournaments.NewBrowsingService(repo)
	for _, sport := range []tournaments.Sport{tournaments.SportFootball, tournaments.SportBasketball, tournaments.SportHandball, tournaments.SportTennis, tournaments.SportPadel, tournaments.SportTableTennis, tournaments.SportVolleyball, tournaments.SportBadminton} {
		input := tournaments.CreateInput{Name: string(sport), Sport: sport, Teams: []tournaments.TeamInput{{Name: "Mine"}, {Name: "Opponent"}}}
		format := "league"
		score := tournaments.MatchResultInput{HomeScore: 2, AwayScore: 0}
		switch sport {
		case tournaments.SportTennis, tournaments.SportPadel:
			input.BestOfSets = 3
			format = "single_elimination"
			score = tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 6, AwayScore: 0}, {HomeScore: 6, AwayScore: 0}}}
		case tournaments.SportTableTennis:
			input.BestOfSets = 3
			format = "single_elimination"
			score = tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 11, AwayScore: 0}, {HomeScore: 11, AwayScore: 0}}}
		case tournaments.SportBadminton:
			input.BestOfSets = 3
			input.PointsPerGame = 21
			format = "single_elimination"
			score = tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 21, AwayScore: 0}, {HomeScore: 21, AwayScore: 0}}}
		case tournaments.SportVolleyball:
			input.BestOfSets = 5
			score = volleyballResult(true, 0)
		}
		value, err := create.Create(ctx, owner, input)
		if err != nil {
			t.Fatal(sport, err)
		}
		own := value.Teams[0].ID
		legs := 1
		if format == "single_elimination" {
			legs = 0
		}
		value, err = create.Start(ctx, owner, value.ID, tournaments.StartInput{Format: format, RoundRobinLegs: legs})
		if err != nil {
			t.Fatal(sport, err)
		}
		match := value.Matches[0]
		if match.AwayTeamID == own {
			score.HomeScore, score.AwayScore = score.AwayScore, score.HomeScore
			for i := range score.Sets {
				score.Sets[i].HomeScore, score.Sets[i].AwayScore = score.Sets[i].AwayScore, score.Sets[i].HomeScore
			}
		}
		if _, err = create.RecordResult(ctx, owner, value.ID, match.ID, score); err != nil {
			t.Fatal(sport, err)
		}
	}
	p, err := browse.PersonalProfile(ctx, owner, "", "", "", 1)
	if err != nil || len(p.Statistics) != 8 || len(p.Items) != 1 || p.NextCursor == "" {
		t.Fatalf("sports %+v %v", p, err)
	}
	for _, stat := range p.Statistics {
		if stat.Played != 1 || stat.Wins != 1 || stat.Losses != 0 || stat.Draws != 0 || stat.WinPercentage == nil || *stat.WinPercentage != 100 {
			t.Fatalf("stat %+v", stat)
		}
	}
	filtered, err := browse.PersonalProfile(ctx, owner, "volleyball", "", "", 20)
	if err != nil || len(filtered.Statistics) != 1 || len(filtered.Items) != 1 || filtered.Statistics[0].Sport != tournaments.SportVolleyball || len(filtered.Items[0].Match.Sets) != 3 {
		t.Fatalf("filter %+v %v", filtered, err)
	}
}
