package postgres

import (
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationOptionalSchedulingAuthorizationAndPersistence(t *testing.T) {
	pool := integrationPool(t)
	ctx := t.Context()
	owner := createVerifiedLocalAccount(ctx, t, pool, "schedule-owner@example.test", "schedule_owner", "password123")
	admin := createVerifiedLocalAccount(ctx, t, pool, "schedule-admin@example.test", "schedule_admin", "password123")
	outsider := createVerifiedLocalAccount(ctx, t, pool, "schedule-other@example.test", "schedule_other", "password123")
	repository := NewAccountTournamentRepository(pool)
	creation := tournaments.NewCreationService(repository)
	scheduling := tournaments.NewSchedulingService(repository)
	value, err := creation.Create(ctx, owner, incidentCreation(tournaments.SportFootball))
	if err != nil {
		t.Fatal(err)
	}
	other, err := creation.Create(ctx, owner, incidentCreation(tournaments.SportFootball))
	if err != nil {
		t.Fatal(err)
	}
	value, err = scheduling.SetVenue(ctx, owner, value.ID, "  Central  ")
	if err != nil || value.Venue != "Central" {
		t.Fatalf("published venue %v %+v", err, value)
	}
	if err = creation.AssignAdministrator(ctx, owner, value.ID, "schedule_admin"); err != nil {
		t.Fatal(err)
	}
	value, err = scheduling.SetVenue(ctx, admin, value.ID, "Club")
	if err != nil || value.Venue != "Club" {
		t.Fatalf("delegated venue %v", err)
	}
	value, err = creation.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	matchID := value.Matches[0].ID
	input := tournaments.ScheduleInput{ScheduledDate: "2026-10-18", ScheduledTime: "18:30", TimeZone: "Europe/Madrid", Venue: "Pista 2"}
	if _, err = scheduling.SetMatchSchedule(ctx, outsider, value.ID, matchID, input); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("outsider %v", err)
	}
	if _, err = scheduling.SetMatchSchedule(ctx, owner, other.ID, matchID, input); !errors.Is(err, tournaments.ErrTournamentNotFound) {
		t.Fatalf("cross tournament %v", err)
	}
	value, err = scheduling.SetMatchSchedule(ctx, admin, value.ID, matchID, input)
	if err != nil {
		t.Fatal(err)
	}
	if value.Matches[0].ScheduledDate != input.ScheduledDate || value.Matches[0].ScheduledTime != input.ScheduledTime || value.Matches[0].TimeZone != input.TimeZone || value.Matches[0].Venue != input.Venue || value.Matches[0].State != "pending" || len(value.Standings) != 4 {
		t.Fatalf("projection %+v", value)
	}
	value, err = creation.RecordResult(ctx, owner, value.ID, matchID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 0})
	if err != nil || value.Matches[0].Venue != input.Venue {
		t.Fatalf("result lost scheduling %v", err)
	}
	value, err = scheduling.SetMatchSchedule(ctx, owner, value.ID, matchID, tournaments.ScheduleInput{})
	if err != nil || value.Matches[0].Venue != "" || value.Matches[0].ScheduledDate != "" || value.Matches[0].ScheduledTime != "" || value.Matches[0].TimeZone != "" || value.Venue != "Club" || value.Matches[0].HomeScore == nil || *value.Matches[0].HomeScore != 1 {
		t.Fatalf("clear changed result or inheritance %v %+v", err, value)
	}
	if err = creation.RemoveAdministrator(ctx, owner, value.ID, "schedule_admin"); err != nil {
		t.Fatal(err)
	}
	if _, err = scheduling.SetVenue(ctx, admin, value.ID, "Lost permission"); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("removed admin %v", err)
	}
	if err = creation.TransferOwnership(ctx, owner, value.ID, "schedule_other"); err != nil {
		t.Fatal(err)
	}
	if _, err = scheduling.SetVenue(ctx, owner, value.ID, "Old owner"); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("old owner %v", err)
	}
	value, err = creation.Cancel(ctx, outsider, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = scheduling.SetMatchSchedule(ctx, outsider, value.ID, matchID, input); !errors.Is(err, tournaments.ErrSchedulingConflict) {
		t.Fatalf("closed %v", err)
	}
	value, err = creation.GetPublic(ctx, value.ID)
	if err != nil || value.Venue != "Club" {
		t.Fatalf("public read %v", err)
	}
}

func TestIntegrationByeAndCompletedSchedulingAreReadOnly(t *testing.T) {
	pool := integrationPool(t)
	ctx := t.Context()
	owner := createVerifiedLocalAccount(ctx, t, pool, "schedule-bye@example.test", "schedule_owner", "password123")
	repository := NewAccountTournamentRepository(pool)
	creation := tournaments.NewCreationService(repository)
	scheduling := tournaments.NewSchedulingService(repository)
	input := incidentCreation(tournaments.SportFootball)
	input.Teams = input.Teams[:3]
	value, err := creation.Create(ctx, owner, input)
	if err != nil {
		t.Fatal(err)
	}
	value, err = creation.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range value.Matches {
		if match.State == "bye" {
			if _, err = scheduling.SetMatchSchedule(ctx, owner, value.ID, match.ID, tournaments.ScheduleInput{Venue: "Pista"}); !errors.Is(err, tournaments.ErrSchedulingConflict) {
				t.Fatalf("bye %v", err)
			}
		}
	}
	for _, match := range value.Matches {
		if match.State != "bye" {
			value, err = creation.RecordResult(ctx, owner, value.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 0})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	value, err = creation.Complete(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = scheduling.SetVenue(ctx, owner, value.ID, "Pista"); !errors.Is(err, tournaments.ErrSchedulingConflict) {
		t.Fatalf("completed %v", err)
	}
}
