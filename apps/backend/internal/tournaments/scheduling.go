package tournaments

import (
	"context"
	"errors"
	"strings"
	"time"
	_ "time/tzdata" // Keep IANA validation available in the minimal runtime image.
	"unicode"
	"unicode/utf8"
)

// ErrSchedulingConflict identifies read-only tournament or bye metadata.
var ErrSchedulingConflict = errors.New("planificación no editable")

// ScheduleInput replaces all optional informational fields of a match.
type ScheduleInput struct {
	ScheduledDate string `json:"scheduledDate"`
	ScheduledTime string `json:"scheduledTime"`
	TimeZone      string `json:"timeZone"`
	Venue         string `json:"venue"`
}

// SchedulingRepository protects authorized informational updates.
type SchedulingRepository interface {
	SetVenue(context.Context, string, string, string) (Tournament, error)
	SetMatchSchedule(context.Context, string, string, string, ScheduleInput) (Tournament, error)
}

// SchedulingService validates informational data independently of transport and storage.
type SchedulingService struct{ repository SchedulingRepository }

// NewSchedulingService wires the scheduling persistence boundary.
func NewSchedulingService(repository SchedulingRepository) SchedulingService {
	return SchedulingService{repository: repository}
}

// ValidVenue validates optional plain text locations.
func ValidVenue(venue string) bool {
	return utf8.ValidString(venue) && utf8.RuneCountInString(venue) <= 160 && !strings.ContainsFunc(venue, unicode.IsControl)
}

// ValidSchedule checks civil date/time values without inventing a UTC instant.
func ValidSchedule(input ScheduleInput) bool {
	if !ValidVenue(input.Venue) {
		return false
	}
	if input.ScheduledDate != "" {
		date, err := time.Parse("2006-01-02", input.ScheduledDate)
		if err != nil || date.Year() < 1 || len(input.ScheduledDate) != 10 {
			return false
		}
	}
	if input.ScheduledTime == "" {
		return input.TimeZone == ""
	}
	if len(input.ScheduledTime) != 5 {
		return false
	}
	if _, err := time.Parse("15:04", input.ScheduledTime); err != nil {
		return false
	}
	if input.TimeZone == "" || input.TimeZone == "Local" || len(input.TimeZone) > 100 {
		return false
	}
	_, err := time.LoadLocation(input.TimeZone)
	return err == nil
}

// SetVenue replaces the optional shared location and returns the public projection.
func (s SchedulingService) SetVenue(ctx context.Context, accountID, tournamentID, venue string) (Tournament, error) {
	venue = strings.TrimSpace(venue)
	if !ValidVenue(venue) {
		return Tournament{}, ErrInvalidTournamentInput
	}
	value, err := s.repository.SetVenue(ctx, accountID, tournamentID, venue)
	return (CreationService{}).withStandings(value), err
}

// SetMatchSchedule replaces independent optional fields of one match.
func (s SchedulingService) SetMatchSchedule(ctx context.Context, accountID, tournamentID, matchID string, input ScheduleInput) (Tournament, error) {
	input.Venue = strings.TrimSpace(input.Venue)
	if !ValidSchedule(input) {
		return Tournament{}, ErrInvalidTournamentInput
	}
	value, err := s.repository.SetMatchSchedule(ctx, accountID, tournamentID, matchID, input)
	return (CreationService{}).withStandings(value), err
}
