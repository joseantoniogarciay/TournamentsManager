package tournaments

import "errors"

var (
	// ErrInvalidBracketResult rejects scores that do not determine one winner.
	ErrInvalidBracketResult = errors.New("resultado de eliminatoria inválido")
	// ErrBracketMatchNotReady rejects a result before both participants are known.
	ErrBracketMatchNotReady = errors.New("bracket match not ready")
	// ErrBracketMatchNotFound rejects a logical round and sequence outside the bracket.
	ErrBracketMatchNotFound = errors.New("partido no disponible")
	// ErrBracketResultDependency prevents rewriting history already consumed downstream.
	ErrBracketResultDependency = errors.New("resultado posterior ya registrado")
	// ErrInvalidBracketStructure rejects a malformed or inconsistent match graph.
	ErrInvalidBracketStructure = errors.New("cuadro inválido")
)

// BracketResult keeps shootout goals separate from the match score. A tied
// match requires a decisive shootout; an unequal score must not have one.
type BracketResult struct {
	HomeScore     int
	AwayScore     int
	HomePenalties *int
	AwayPenalties *int
}

func (r BracketResult) validate() error {
	if r.HomeScore < 0 || r.AwayScore < 0 {
		return ErrInvalidBracketResult
	}
	if r.HomeScore != r.AwayScore {
		if r.HomePenalties != nil || r.AwayPenalties != nil {
			return ErrInvalidBracketResult
		}
		return nil
	}
	if r.HomePenalties == nil || r.AwayPenalties == nil || *r.HomePenalties < 0 || *r.AwayPenalties < 0 || *r.HomePenalties == *r.AwayPenalties {
		return ErrInvalidBracketResult
	}
	return nil
}

// ResolvedBracketMatch preserves each source alongside its currently known
// participant. Empty team IDs mean unresolved slots, never fictional teams.
type ResolvedBracketMatch struct {
	BracketMatch
	HomeTeamID   string
	AwayTeamID   string
	WinnerTeamID string
}

type matchPosition struct{ round, sequence int }

// Resolve derives participants and winners in round order from persisted
// results. It rejects dangling, duplicated or out-of-order dependencies.
func (b Bracket) Resolve() ([]ResolvedBracketMatch, error) {
	if b.Size < minimumEntrants || b.Size > maximumEntrants || nextPowerOfTwo(b.Size) != b.Size || len(b.Matches) != b.Size-1 {
		return nil, ErrInvalidBracketStructure
	}
	resolved := make([]ResolvedBracketMatch, 0, len(b.Matches))
	winners := make(map[matchPosition]string, len(b.Matches))
	teams := make(map[string]bool)
	expectedRound, expectedSequence, roundSize := 1, 1, b.Size/2
	for _, match := range b.Matches {
		if match.Round != expectedRound || match.Sequence != expectedSequence {
			return nil, ErrInvalidBracketStructure
		}
		resolveSource := func(source SlotSource, side int) (string, error) {
			if match.Round == 1 {
				if source.SourceRound != 0 || source.SourceSequence != 0 {
					return "", ErrInvalidBracketStructure
				}
				switch source.Kind {
				case SeededTeam:
					if source.TeamID == "" || teams[source.TeamID] {
						return "", ErrInvalidBracketStructure
					}
					teams[source.TeamID] = true
					return source.TeamID, nil
				case Bye:
					if source.TeamID == "" {
						return "", nil
					}
				}
				return "", ErrInvalidBracketStructure
			}
			if source.Kind != Winner || source.TeamID != "" || source.SourceRound != match.Round-1 || source.SourceSequence != 2*match.Sequence-1+side {
				return "", ErrInvalidBracketStructure
			}
			winner, exists := winners[matchPosition{source.SourceRound, source.SourceSequence}]
			if !exists {
				return "", ErrInvalidBracketStructure
			}
			return winner, nil
		}
		home, err := resolveSource(match.Home, 0)
		if err != nil {
			return nil, err
		}
		away, err := resolveSource(match.Away, 1)
		if err != nil {
			return nil, err
		}
		winner := ""
		isBye := match.Home.Kind == Bye || match.Away.Kind == Bye
		if isBye {
			if match.Home.Kind == Bye && match.Away.Kind == Bye || match.Result != nil {
				return nil, ErrInvalidBracketStructure
			}
			winner = home
			if winner == "" {
				winner = away
			}
		}
		if match.AutoWinnerTeamID != winner {
			return nil, ErrInvalidBracketStructure
		}
		if match.Result != nil {
			if home == "" || away == "" {
				return nil, ErrBracketMatchNotReady
			}
			if err := match.Result.validate(); err != nil {
				return nil, err
			}
			homeWins := match.Result.HomeScore > match.Result.AwayScore
			if match.Result.HomeScore == match.Result.AwayScore {
				homeWins = *match.Result.HomePenalties > *match.Result.AwayPenalties
			}
			winner = away
			if homeWins {
				winner = home
			}
		}
		winners[matchPosition{match.Round, match.Sequence}] = winner
		resolved = append(resolved, ResolvedBracketMatch{BracketMatch: match, HomeTeamID: home, AwayTeamID: away, WinnerTeamID: winner})
		expectedSequence++
		if expectedSequence > roundSize {
			expectedRound++
			expectedSequence = 1
			roundSize /= 2
		}
	}
	return resolved, nil
}

// RecordResult returns a new bracket, leaving the previous snapshot untouched
// for history. A result cannot be corrected after a descendant has a result.
func (b Bracket) RecordResult(round, sequence int, result BracketResult) (Bracket, error) {
	if err := result.validate(); err != nil {
		return Bracket{}, err
	}
	resolved, err := b.Resolve()
	if err != nil {
		return Bracket{}, err
	}
	index := -1
	for i, match := range resolved {
		if match.Round == round && match.Sequence == sequence {
			index = i
			break
		}
	}
	if index == -1 {
		return Bracket{}, ErrBracketMatchNotFound
	}
	match := resolved[index]
	if match.HomeTeamID == "" || match.AwayTeamID == "" || match.AutoWinnerTeamID != "" {
		return Bracket{}, ErrBracketMatchNotReady
	}
	ancestorRound, ancestorSequence := round+1, (sequence+1)/2
	for _, descendant := range b.Matches[index+1:] {
		if descendant.Round == ancestorRound && descendant.Sequence == ancestorSequence {
			if descendant.Result != nil {
				return Bracket{}, ErrBracketResultDependency
			}
			ancestorRound++
			ancestorSequence = (ancestorSequence + 1) / 2
		}
	}
	next := Bracket{Size: b.Size, Matches: make([]BracketMatch, len(b.Matches))}
	copy(next.Matches, b.Matches)
	for i := range next.Matches {
		if next.Matches[i].Result != nil {
			cloned := cloneBracketResult(*next.Matches[i].Result)
			next.Matches[i].Result = &cloned
		}
	}
	cloned := cloneBracketResult(result)
	next.Matches[index].Result = &cloned
	return next, nil
}

func cloneBracketResult(result BracketResult) BracketResult {
	if result.HomePenalties != nil {
		value := *result.HomePenalties
		result.HomePenalties = &value
	}
	if result.AwayPenalties != nil {
		value := *result.AwayPenalties
		result.AwayPenalties = &value
	}
	return result
}
