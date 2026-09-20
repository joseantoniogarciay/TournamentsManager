// Package tournaments contains competition rules independent from HTTP and PostgreSQL.
package tournaments

import (
	"errors"
)

const (
	minimumEntrants = 2
	maximumEntrants = 64
)

// ErrInvalidBracketEntrants rejects a field outside the supported size or with
// empty or duplicated team identities.
var ErrInvalidBracketEntrants = errors.New("invalid bracket entrants")

// SlotSourceKind states why a bracket slot contains a team.
type SlotSourceKind string

const (
	// SeededTeam identifies a slot fixed from the tournament's initial seeding.
	SeededTeam SlotSourceKind = "seeded_team"
	// Winner identifies a slot supplied by a previous bracket match.
	Winner SlotSourceKind = "winner"
	// Bye identifies an empty slot that gives its opponent an automatic advance.
	Bye SlotSourceKind = "bye"
)

// SlotSource is a stable, explainable source for one side of a match. A winner
// reference uses its source match's logical round and sequence; persistence
// later replaces that logical identity with the stored match id.
type SlotSource struct {
	Kind           SlotSourceKind
	TeamID         string
	SourceRound    int
	SourceSequence int
}

// BracketMatch is the structural representation of a single-elimination match.
// AutoWinnerTeamID is populated only for a bye; it is not a recorded result.
type BracketMatch struct {
	Round            int
	Sequence         int
	Home             SlotSource
	Away             SlotSource
	AutoWinnerTeamID string
	Result           *BracketResult
}

// Bracket is a complete, ordered tree for one single-elimination phase.
type Bracket struct {
	Size    int
	Sport   Sport
	Matches []BracketMatch
}

// GenerateSingleElimination materializes every round at start time. Team input
// order is the frozen seed order. Non-power-of-two fields are padded with byes,
// so every future slot has a deterministic and visible origin.
func GenerateSingleElimination(teamIDs []string) (Bracket, error) {
	return generateSingleElimination(teamIDs, false)
}

// GenerateSeededSingleElimination builds a full bracket where the first seed
// faces the last and the two best seeds can only meet in the final.
func GenerateSeededSingleElimination(teamIDs []string) (Bracket, error) {
	return generateSingleElimination(teamIDs, true)
}

func generateSingleElimination(teamIDs []string, requireFull bool) (Bracket, error) {
	if len(teamIDs) < minimumEntrants || len(teamIDs) > maximumEntrants {
		return Bracket{}, ErrInvalidBracketEntrants
	}
	seen := make(map[string]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == "" {
			return Bracket{}, ErrInvalidBracketEntrants
		}
		if _, exists := seen[teamID]; exists {
			return Bracket{}, ErrInvalidBracketEntrants
		}
		seen[teamID] = struct{}{}
	}

	size := nextPowerOfTwo(len(teamIDs))
	if requireFull && size != len(teamIDs) {
		return Bracket{}, ErrInvalidBracketEntrants
	}
	if requireFull {
		return bracketFromSlots(seedOrderedSlots(teamIDs), size)
	}
	// Spread seeds through first-round pairs before giving any pair a second
	// seed. Since size is the next power of two, this guarantees no pair is
	// composed of two byes while preserving a deterministic seed order.
	slots := make([]SlotSource, size)
	for index := range slots {
		slots[index] = SlotSource{Kind: Bye}
	}
	pairCount := size / 2
	for index, teamID := range teamIDs {
		pair := index % pairCount
		side := index / pairCount
		slots[pair*2+side] = SlotSource{Kind: SeededTeam, TeamID: teamID}
	}

	return bracketFromSlots(slots, size)
}

func seedOrderedSlots(teamIDs []string) []SlotSource {
	order := []int{1, 2}
	for field := 4; field <= len(teamIDs); field *= 2 {
		next := make([]int, 0, field)
		for _, seed := range order {
			next = append(next, seed, field+1-seed)
		}
		order = next
	}
	slots := make([]SlotSource, len(teamIDs))
	for index, seed := range order {
		slots[index] = SlotSource{Kind: SeededTeam, TeamID: teamIDs[seed-1]}
	}
	return slots
}

func bracketFromSlots(slots []SlotSource, size int) (Bracket, error) {
	bracket := Bracket{Size: size}
	for round := 1; len(slots) > 1; round++ {
		nextSlots := make([]SlotSource, 0, len(slots)/2)
		for index := 0; index < len(slots); index += 2 {
			sequence := index/2 + 1
			match := BracketMatch{Round: round, Sequence: sequence, Home: slots[index], Away: slots[index+1]}
			if match.Home.Kind == Bye && match.Away.Kind == Bye {
				return Bracket{}, ErrInvalidBracketEntrants
			}
			if match.Home.Kind == Bye {
				match.AutoWinnerTeamID = match.Away.TeamID
			}
			if match.Away.Kind == Bye {
				match.AutoWinnerTeamID = match.Home.TeamID
			}
			bracket.Matches = append(bracket.Matches, match)
			nextSlots = append(nextSlots, SlotSource{Kind: Winner, SourceRound: round, SourceSequence: sequence})
		}
		slots = nextSlots
	}
	return bracket, nil
}

func nextPowerOfTwo(value int) int {
	result := 1
	for result < value {
		result <<= 1
	}
	return result
}
