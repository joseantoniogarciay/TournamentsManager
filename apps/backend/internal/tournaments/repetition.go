package tournaments

import (
	"context"
	"errors"
	"regexp"
)

// ErrTournamentRepetitionConflict indicates a source that has not completed.
var ErrTournamentRepetitionConflict = errors.New("torneo no finalizado")

// RepeatInput is a reviewed new edition. The first team belongs to its organizer.
type RepeatInput struct {
	RequestID string      `json:"requestId"`
	Name      string      `json:"name"`
	Teams     []TeamInput `json:"teams"`
}

var repetitionIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// Repeat validates the reviewed composition; persistence checks the source atomically.
func (s CreationService) Repeat(ctx context.Context, accountID, sourceID string, input RepeatInput) (Tournament, error) {
	composition := CreateInput{Name: input.Name, Sport: SportFootball, Teams: input.Teams}
	if !repetitionIDPattern.MatchString(input.RequestID) || !validCreateInput(composition) {
		return Tournament{}, ErrInvalidTournamentInput
	}
	normalized := normalizedInput(composition)
	input.Name, input.Teams = normalized.Name, normalized.Teams
	value, err := s.repository.Repeat(ctx, accountID, sourceID, input)
	return s.withStandings(value), err
}

// RepetitionStartConfiguration copies only the source's initial sporting rules.
func RepetitionStartConfiguration(source Tournament) StartInput {
	config := StartInput{Format: source.Format}
	if source.Format == "single_elimination" {
		return config
	}
	config.RoundRobinLegs = source.RoundRobinLegs
	if source.Format == "league_then_single_elimination" {
		for _, stage := range source.Stages {
			if stage.Type == "league" {
				config.LeagueStructure = stage.LeagueStructure
				config.QualifierCount = stage.QualifierCount
				config.GroupCount = stage.GroupCount
				config.QualifiersPerGroup = stage.QualifiersPerGroup
				break
			}
		}
	}
	return config
}
