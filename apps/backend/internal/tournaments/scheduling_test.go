package tournaments

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type schedulingStub struct {
	input ScheduleInput
	calls int
	err   error
}

func (r *schedulingStub) SetVenue(_ context.Context, _, _, venue string) (Tournament, error) {
	r.input.Venue = venue
	r.calls++
	return Tournament{}, r.err
}
func (r *schedulingStub) SetMatchSchedule(_ context.Context, _, _, _ string, input ScheduleInput) (Tournament, error) {
	r.input = input
	r.calls++
	return Tournament{}, r.err
}

func TestSchedulingValidatesCivilValuesBeforePersistence(t *testing.T) {
	t.Parallel()
	for _, input := range []ScheduleInput{
		{ScheduledDate: "2026-02-29"}, {ScheduledDate: "0000-01-01"}, {ScheduledDate: "2026-2-01"},
		{ScheduledTime: "24:00", TimeZone: "Europe/Madrid"}, {ScheduledTime: "18:30"},
		{ScheduledTime: "18:30", TimeZone: "Local"}, {ScheduledTime: "18:30", TimeZone: "private/invalid"},
		{TimeZone: "Europe/Madrid"}, {Venue: "one\ntwo"}, {Venue: strings.Repeat("ñ", 161)},
	} {
		repository := &schedulingStub{}
		_, err := NewSchedulingService(repository).SetMatchSchedule(t.Context(), "owner", "tournament", "match", input)
		if !errors.Is(err, ErrInvalidTournamentInput) || repository.calls != 0 {
			t.Fatalf("validation: %+v %v", input, err)
		}
	}
	for _, input := range []ScheduleInput{{}, {Venue: "  Pista 2  "}, {ScheduledDate: "2028-02-29"}, {ScheduledTime: "18:30", TimeZone: "Europe/Madrid"}, {ScheduledDate: "2026-10-18", ScheduledTime: "18:30", TimeZone: "Europe/Madrid", Venue: strings.Repeat("ñ", 160)}} {
		repository := &schedulingStub{}
		_, err := NewSchedulingService(repository).SetMatchSchedule(t.Context(), "owner", "tournament", "match", input)
		if err != nil || repository.calls != 1 || repository.input.Venue != strings.TrimSpace(input.Venue) {
			t.Fatalf("optional values: %+v %v", input, err)
		}
	}
}

func TestVenueNormalizationAndPersistenceFailure(t *testing.T) {
	t.Parallel()
	repository := &schedulingStub{err: ErrTournamentForbidden}
	_, err := NewSchedulingService(repository).SetVenue(t.Context(), "owner", "tournament", "  Central  ")
	if !errors.Is(err, ErrTournamentForbidden) || repository.input.Venue != "Central" {
		t.Fatal("lost normalization or failure")
	}
	_, err = NewSchedulingService(repository).SetVenue(t.Context(), "owner", "tournament", strings.Repeat("a", 161))
	if !errors.Is(err, ErrInvalidTournamentInput) || repository.calls != 1 {
		t.Fatal("invalid location reached persistence")
	}
}
