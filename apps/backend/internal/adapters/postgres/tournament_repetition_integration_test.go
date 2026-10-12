package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationRepetitionIsIndependentIdempotentAndOwnerOnly(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "repeat-owner@example.test", "repeatowner", "correct password")
	other := createVerifiedLocalAccount(ctx, t, pool, "repeat-other@example.test", "repeatother", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	original, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Original", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Blue"}, {Name: "Red"}}})
	if err != nil {
		t.Fatal(err)
	}
	input := tournaments.RepeatInput{RequestID: "019abcde-3333-7333-8333-333333333333", Name: "Next edition", Teams: []tournaments.TeamInput{{Name: "Red"}, {Name: "Blue"}}}
	if _, err := service.Repeat(ctx, owner, original.ID, input); !errors.Is(err, tournaments.ErrTournamentRepetitionConflict) {
		t.Fatalf("open source: %v", err)
	}
	token, err := service.CreateTeamInvitation(ctx, owner, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.JoinTeamInvitation(ctx, other, token, tournaments.TeamInput{Name: "Red"}); err != nil {
		t.Fatal(err)
	}
	if err := service.AssignAdministrator(ctx, owner, original.ID, "repeatother"); err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, original.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 2})
	if err != nil {
		t.Fatal(err)
	}
	scheduler := tournaments.NewSchedulingService(NewAccountTournamentRepository(pool))
	if _, err := scheduler.SetVenue(ctx, owner, original.ID, "Old venue"); err != nil {
		t.Fatal(err)
	}
	if _, err := scheduler.SetMatchSchedule(ctx, owner, original.ID, started.Matches[0].ID, tournaments.ScheduleInput{ScheduledDate: "2026-10-10", Venue: "Old pitch"}); err != nil {
		t.Fatal(err)
	}
	for _, match := range started.Matches {
		if _, err := service.RecordResult(ctx, owner, original.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 0}); err != nil {
			t.Fatal(err)
		}
	}
	finished, err := service.Complete(ctx, owner, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Repeat(ctx, other, original.ID, input); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("delegate: %v", err)
	}
	// Locking the source makes concurrent retries return one creation intent.
	ids := make(chan string, 8)
	failures := make(chan error, 8)
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			value, err := service.Repeat(ctx, owner, original.ID, input)
			if err != nil {
				failures <- err
			} else {
				ids <- value.ID
			}
		})
	}
	group.Wait()
	close(ids)
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	var editionID string
	for id := range ids {
		if editionID != "" && id != editionID {
			t.Fatal("duplicate edition")
		}
		editionID = id
	}
	edition, err := service.GetPublic(ctx, editionID)
	if err != nil {
		t.Fatal(err)
	}
	if edition.ID == original.ID || edition.State != "published" || edition.Venue != "" || len(edition.Matches) != 0 || len(edition.Stages) != 0 || len(edition.ChampionTeamIDs) != 0 || edition.SuggestedStart == nil || edition.SuggestedStart.RoundRobinLegs != 2 {
		t.Fatalf("new edition: %#v", edition)
	}
	for _, team := range edition.Teams {
		for _, old := range original.Teams {
			if team.ID == old.ID {
				t.Fatal("team identity copied")
			}
		}
		if team.Withdrawn {
			t.Fatal("withdrawal copied")
		}
	}
	var admins, followers, links, invitations int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM tournament_administrators WHERE tournament_id=$1),(SELECT count(*) FROM tournament_followers WHERE tournament_id=$1),(SELECT count(*) FROM tournament_team_accounts WHERE tournament_id=$1),(SELECT count(*) FROM tournament_team_invitations WHERE tournament_id=$1)`, editionID).Scan(&admins, &followers, &links, &invitations); err != nil {
		t.Fatal(err)
	}
	if admins != 0 || followers != 0 || links != 1 || invitations != 0 {
		t.Fatalf("copied relations: %d %d %d %d", admins, followers, links, invitations)
	}
	unchanged, err := service.GetPublic(ctx, original.ID)
	// The source remains immutable; its derived family indicator changes when a child exists.
	if !unchanged.HasOtherEditions {
		t.Fatal("source must expose the new edition relationship")
	}
	finished.HasOtherEditions = true
	if err != nil || !reflect.DeepEqual(finished, unchanged) {
		t.Fatalf("source changed: %v", err)
	}
	next, err := service.Start(ctx, owner, editionID, *edition.SuggestedStart)
	if err != nil || next.RoundRobinLegs != 2 || next.SuggestedStart != nil {
		t.Fatalf("start inherited: %#v %v", next, err)
	}
	retry, err := service.Repeat(ctx, owner, original.ID, input)
	if err != nil || retry.ID != editionID || retry.State != "in_progress" {
		t.Fatalf("retry after start: %#v %v", retry, err)
	}
	if err := service.TransferOwnership(ctx, owner, original.ID, "repeatother"); err != nil {
		t.Fatal(err)
	}
	input.RequestID = "019abcde-4444-7444-8444-444444444444"
	if _, err := service.Repeat(ctx, owner, original.ID, input); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("former owner: %v", err)
	}
	if _, err := service.Repeat(ctx, other, original.ID, input); err != nil {
		t.Fatalf("current owner: %v", err)
	}
	cancelled, err := service.Create(ctx, other, tournaments.CreateInput{Name: "Cancelled", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Only"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Cancel(ctx, other, cancelled.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Repeat(ctx, other, cancelled.ID, input); !errors.Is(err, tournaments.ErrTournamentRepetitionConflict) {
		t.Fatalf("cancelled source: %v", err)
	}
	cancelledEdition, err := service.Repeat(ctx, other, original.ID, tournaments.RepeatInput{RequestID: "019abcde-5555-7555-8555-555555555555", Name: "Cancel edition", Teams: input.Teams})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Cancel(ctx, other, cancelledEdition.ID); err != nil {
		t.Fatalf("clear suggestion on cancel: %v", err)
	}
	if err := service.TransferOwnership(ctx, owner, editionID, "repeatother"); err != nil {
		t.Fatal(err)
	}
	repository := NewAccountTournamentRepository(pool)
	if _, err := repository.ScheduleAccountDeletion(ctx, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE accounts SET deletion_requested_at=now()-interval '31 days' WHERE id=$1`, owner); err != nil {
		t.Fatal(err)
	}
	if purged, err := repository.PurgeExpired(ctx, 10); err != nil || purged != 1 {
		t.Fatalf("repetition blocked account deletion: %d %v", purged, err)
	}
	if _, err := service.GetPublic(ctx, editionID); err != nil {
		t.Fatal(err)
	}
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := service.Repeat(cancelledCtx, other, original.ID, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}

func TestIntegrationRepetitionPreservesSportRulesAndFormats(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "matrix@example.test", "matrixowner", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	for i, c := range []struct {
		sport        tournaments.Sport
		best, points int
		config       tournaments.StartInput
	}{
		{tournaments.SportFootball, 0, 0, tournaments.StartInput{Format: "league", RoundRobinLegs: 2}},
		{tournaments.SportBasketball, 0, 0, tournaments.StartInput{Format: "single_elimination"}},
		{tournaments.SportHandball, 0, 0, tournaments.StartInput{Format: "league_then_single_elimination", RoundRobinLegs: 1, LeagueStructure: "single_table", QualifierCount: 2}},
		{tournaments.SportVolleyball, 5, 0, tournaments.StartInput{Format: "league_then_single_elimination", RoundRobinLegs: 2, LeagueStructure: "groups", GroupCount: 2, QualifiersPerGroup: 1}},
		{tournaments.SportTennis, 5, 0, tournaments.StartInput{Format: "single_elimination"}},
		{tournaments.SportPadel, 3, 0, tournaments.StartInput{Format: "single_elimination"}},
		{tournaments.SportTableTennis, 7, 0, tournaments.StartInput{Format: "single_elimination"}},
		{tournaments.SportBadminton, 3, 15, tournaments.StartInput{Format: "single_elimination"}},
	} {
		original, err := service.Create(ctx, owner, tournaments.CreateInput{Name: fmt.Sprintf("Source %d", i), Sport: c.sport, BestOfSets: c.best, PointsPerGame: c.points, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.Start(ctx, owner, original.ID, c.config); err != nil {
			t.Fatal(err)
		}
		// Fixture: freeze the started source to test copying rules independently of results.
		if _, err := pool.Exec(ctx, `UPDATE tournaments SET state='completed' WHERE id=$1`, original.ID); err != nil {
			t.Fatal(err)
		}
		edition, err := service.Repeat(ctx, owner, original.ID, tournaments.RepeatInput{RequestID: fmt.Sprintf("019abcde-3333-7333-8333-%012d", i), Name: "Edition", Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}}})
		if err != nil {
			t.Fatal(err)
		}
		if edition.Sport != c.sport || edition.BestOfSets != c.best || edition.PointsPerGame != c.points || edition.SuggestedStart == nil || !reflect.DeepEqual(*edition.SuggestedStart, c.config) {
			t.Fatalf("rules %s: %#v", c.sport, edition)
		}
		if _, err := service.Start(ctx, owner, edition.ID, *edition.SuggestedStart); err != nil {
			t.Fatalf("restart %s: %v", c.sport, err)
		}
	}
}
