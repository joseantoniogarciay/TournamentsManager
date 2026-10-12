package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationPersonalMatchesAndEditionFamilies(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repo := NewAccountTournamentRepository(pool)
	owner := createVerifiedLocalAccount(ctx, t, pool, "browse-owner@example.test", "browseowner", "correct password")
	participant := createVerifiedLocalAccount(ctx, t, pool, "browse-player@example.test", "browseplayer", "correct password")
	delegate := createVerifiedLocalAccount(ctx, t, pool, "browse-delegate@example.test", "browsedelegate", "correct password")
	create := tournaments.NewCreationService(repo)
	browse := tournaments.NewBrowsingService(repo)
	original, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "First", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}, {Name: "Green"}}})
	if err != nil {
		t.Fatal(err)
	}
	invitation, err := create.CreateTeamInvitation(ctx, owner, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = create.JoinTeamInvitation(ctx, participant, invitation, tournaments.TeamInput{Name: "Red"}); err != nil {
		t.Fatal(err)
	}
	if err = create.AssignAdministrator(ctx, owner, original.ID, "browsedelegate"); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Follow(ctx, delegate, original.ID); err != nil {
		t.Fatal(err)
	}
	page, err := browse.PersonalMatches(ctx, owner, "pending", "", 1)
	if err != nil || page.HasMatches || len(page.Items) != 0 {
		t.Fatalf("published: %+v %v", page, err)
	}
	started, err := create.Start(ctx, owner, original.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	page, err = browse.PersonalMatches(ctx, delegate, "pending", "", 1)
	if err != nil || page.HasMatches {
		t.Fatalf("delegate/follower is not participant: %+v %v", page, err)
	}
	page, err = browse.PersonalMatches(ctx, owner, "pending", "", 1)
	if err != nil || len(page.Items) != 1 || !page.HasMatches || page.NextCursor == "" {
		t.Fatalf("pending page: %+v %v", page, err)
	}
	next, err := browse.PersonalMatches(ctx, owner, "pending", page.NextCursor, 1)
	if err != nil || len(next.Items) != 1 || next.Items[0].Match.ID == page.Items[0].Match.ID {
		t.Fatalf("cursor: %+v %v", next, err)
	}
	scheduled := next.Items[0].Match.ID
	scheduling := tournaments.NewSchedulingService(repo)
	if _, err = scheduling.SetVenue(ctx, owner, original.ID, "Common court"); err != nil {
		t.Fatal(err)
	}
	if _, err = scheduling.SetMatchSchedule(ctx, owner, original.ID, scheduled, tournaments.ScheduleInput{ScheduledDate: "2026-10-18", ScheduledTime: "18:30", TimeZone: "Europe/Madrid"}); err != nil {
		t.Fatal(err)
	}
	page, err = browse.PersonalMatches(ctx, owner, "pending", "", 1)
	if err != nil || page.Items[0].Match.ID != scheduled || page.Items[0].Match.Venue != "Common court" {
		t.Fatalf("scheduled first/inheritance: %+v %v", page, err)
	}
	if _, err = create.RecordResult(ctx, owner, original.ID, scheduled, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}); err != nil {
		t.Fatal(err)
	}
	results, err := browse.PersonalMatches(ctx, owner, "results", "", 20)
	if err != nil || len(results.Items) != 1 || results.Items[0].Match.HomeScore == nil {
		t.Fatalf("result: %+v %v", results, err)
	}
	for _, match := range started.Matches {
		if match.ID == scheduled {
			continue
		}
		if _, err = create.RecordResult(ctx, owner, original.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 0}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = create.Complete(ctx, owner, original.ID); err != nil {
		t.Fatal(err)
	}
	page, err = browse.PersonalMatches(ctx, owner, "pending", "", 20)
	if err != nil || len(page.Items) != 0 || !page.HasMatches {
		t.Fatalf("empty view retains global evidence: %+v %v", page, err)
	}
	edition, err := create.Repeat(ctx, owner, original.ID, tournaments.RepeatInput{RequestID: "019abcde-1111-7111-8111-111111111111", Name: "Second", Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}}})
	if err != nil {
		t.Fatal(err)
	}
	sibling, err := create.Repeat(ctx, owner, original.ID, tournaments.RepeatInput{RequestID: "019abcde-2222-7222-8222-222222222222", Name: "Branch", Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{original.ID, edition.ID, sibling.ID} {
		family, err := browse.Editions(ctx, id, "", 20)
		if err != nil || len(family.Items) != 3 {
			t.Fatalf("family: %+v %v", family, err)
		}
		value, err := repo.GetPublic(ctx, id)
		if err != nil || !value.HasOtherEditions {
			t.Fatalf("edition indicator: %+v %v", value, err)
		}
	}
	family, err := browse.Editions(ctx, edition.ID, "", 1)
	if err != nil || family.NextCursor == "" {
		t.Fatalf("edition cursor: %+v %v", family, err)
	}
	continuation, err := browse.Editions(ctx, original.ID, family.NextCursor, 1)
	if err != nil || len(continuation.Items) != 1 || continuation.Items[0].ID == family.Items[0].ID {
		t.Fatalf("edition next: %+v %v", continuation, err)
	}

	child, err := create.Start(ctx, owner, edition.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range child.Matches {
		if _, err = create.RecordResult(ctx, owner, child.ID, m.ID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = create.Complete(ctx, owner, child.ID); err != nil {
		t.Fatal(err)
	}
	grandchild, err := create.Repeat(ctx, owner, child.ID, tournaments.RepeatInput{RequestID: "019abcde-4444-7444-8444-444444444444", Name: "Third", Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{original.ID, sibling.ID, grandchild.ID} {
		family, err := browse.Editions(ctx, id, "", 20)
		if err != nil || len(family.Items) != 4 || family.Items[0].ID != grandchild.ID {
			t.Fatalf("nested family/newest first: %+v %v", family, err)
		}
	}
	independent, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "First", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}}})
	if err != nil {
		t.Fatal(err)
	}
	family, err = browse.Editions(ctx, independent.ID, "", 20)
	if err != nil || len(family.Items) != 1 {
		t.Fatalf("same name is not relationship: %+v %v", family, err)
	}

	standalone, err := repo.GetPublic(ctx, independent.ID)
	if err != nil || standalone.HasOtherEditions {
		t.Fatalf("single edition indicator: %+v %v", standalone, err)
	}
	if _, err = browse.Editions(ctx, "019abcde-3333-7333-8333-333333333333", "", 20); !errors.Is(err, tournaments.ErrTournamentNotFound) {
		t.Fatalf("missing %v", err)
	}
}

func TestIntegrationPersonalMatchesExcludeWithdrawalsCancellationAndUnresolvedRivals(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repo := NewAccountTournamentRepository(pool)
	owner := createVerifiedLocalAccount(ctx, t, pool, "browse-edges@example.test", "browseedges", "correct password")
	create := tournaments.NewCreationService(repo)
	browse := tournaments.NewBrowsingService(repo)
	teams := []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}}
	league, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "Withdrawal", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	league, err = create.Start(ctx, owner, league.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = create.WithdrawTeam(ctx, owner, league.ID, league.Teams[0].ID); err != nil {
		t.Fatal(err)
	}
	page, err := browse.PersonalMatches(ctx, owner, "pending", "", 20)
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("withdrawn pending: %+v %v", page, err)
	}
	results, err := browse.PersonalMatches(ctx, owner, "results", "", 20)
	if err != nil || len(results.Items) != 4 {
		t.Fatalf("withdrawal results: %+v %v", results, err)
	}
	for _, item := range results.Items {
		if item.Match.ResultType != tournaments.ResultAdministrative {
			t.Fatal("withdrawal lost administrative result type")
		}
	}
	bracket, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "Bracket", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	bracket, err = create.Start(ctx, owner, bracket.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	page, err = browse.PersonalMatches(ctx, owner, "pending", "", 20)
	if err != nil {
		t.Fatal(err)
	}
	expected := 0
	for _, m := range bracket.Matches {
		if m.State == "pending" && m.HomeTeamID != "" && m.AwayTeamID != "" && (m.HomeTeamID == bracket.Teams[0].ID || m.AwayTeamID == bracket.Teams[0].ID) {
			expected++
		}
	}
	if len(page.Items) != expected {
		t.Fatalf("byes/unresolved rivals: got %d expected %d", len(page.Items), expected)
	}
	for _, item := range page.Items {
		if item.Match.HomeTeamID == "" || item.Match.AwayTeamID == "" {
			t.Fatal("unresolved rival included")
		}
	}
	if _, err = create.Cancel(ctx, owner, bracket.ID); err != nil {
		t.Fatal(err)
	}
	page, err = browse.PersonalMatches(ctx, owner, "pending", "", 20)
	if err != nil || len(page.Items) != 0 || !page.HasMatches {
		t.Fatalf("cancelled: %+v %v", page, err)
	}
}
