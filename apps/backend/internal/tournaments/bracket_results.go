package tournaments

import (
	"errors"
	"slices"
)

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
	Sets          []SetScore
	Incident      *MatchIncident
}

// NormalizeSetResult validates every set and derives the aggregate score.
func NormalizeSetResult(sport Sport, bestOfSets, pointsPerGame int, input MatchResultInput) (MatchResultInput, error) {
	if !SetSport(sport) || !ValidSetFormat(sport, bestOfSets) || !ValidPointsPerGame(sport, pointsPerGame) || input.HomePenalties != nil || input.AwayPenalties != nil {
		return MatchResultInput{}, ErrInvalidBracketResult
	}
	winsNeeded := bestOfSets/2 + 1
	homeWins, awayWins := 0, 0
	sets := make([]SetScore, len(input.Sets))
	copy(sets, input.Sets)
	for index, set := range sets {
		if homeWins == winsNeeded || awayWins == winsNeeded || !validSportSet(sport, pointsPerGame, index, set) {
			return MatchResultInput{}, ErrInvalidBracketResult
		}
		if set.HomeScore > set.AwayScore {
			homeWins++
		} else {
			awayWins++
		}
	}
	if (homeWins != winsNeeded && awayWins != winsNeeded) || (homeWins == winsNeeded && awayWins == winsNeeded) {
		return MatchResultInput{}, ErrInvalidBracketResult
	}
	input.HomeScore, input.AwayScore, input.Sets = homeWins, awayWins, sets
	return input, nil
}

// MaximumSetScore is the storage limit, not a sporting limit on deuce.
const MaximumSetScore = 32767

// ValidSetFormat keeps configuration policy consistent across creation and drafts.
func ValidSetFormat(sport Sport, bestOfSets int) bool {
	switch sport {
	case SportTennis:
		return bestOfSets == 3 || bestOfSets == 5
	case SportVolleyball:
		return bestOfSets == 5
	case SportPadel, SportBadminton:
		return bestOfSets == 3
	case SportTableTennis:
		return bestOfSets == 3 || bestOfSets == 5 || bestOfSets == 7
	default:
		return ValidSport(sport) && bestOfSets == 0
	}
}

// ValidPointsPerGame rejects missing badminton profiles and irrelevant configuration.
func ValidPointsPerGame(sport Sport, points int) bool {
	if sport == SportBadminton {
		return points == 15 || points == 21
	}
	return points == 0
}

func validSportSet(sport Sport, pointsPerGame, index int, set SetScore) bool {
	if set.HomeScore > MaximumSetScore || set.AwayScore > MaximumSetScore {
		return false
	}
	if sport == SportBadminton {
		pointCap := 30
		if pointsPerGame == 15 {
			pointCap = 21
		}
		winner, loser := set.HomeScore, set.AwayScore
		if loser > winner {
			winner, loser = loser, winner
		}
		return ValidPointsPerGame(sport, pointsPerGame) && loser >= 0 && winner <= pointCap && (validPointSet(set, pointsPerGame) || winner == pointCap && loser == pointCap-1)
	}
	if sport == SportVolleyball {
		target := 25
		if index == 4 {
			target = 15
		}
		return validPointSet(set, target)
	}
	if sport == SportTableTennis {
		return validPointSet(set, 11)
	}
	return validTieBreakSet(set)
}

func validPointSet(set SetScore, target int) bool {
	winner, loser := set.HomeScore, set.AwayScore
	if loser > winner {
		winner, loser = loser, winner
	}
	return loser >= 0 && (winner == target && loser <= target-2 || winner > target && winner-loser == 2)
}

func validTieBreakSet(set SetScore) bool {
	if set.HomeScore < 0 || set.AwayScore < 0 || set.HomeScore == set.AwayScore {
		return false
	}
	winner, loser := set.HomeScore, set.AwayScore
	if loser > winner {
		winner, loser = loser, winner
	}
	return winner == 6 && loser <= 4 || winner == 7 && (loser == 5 || loser == 6)
}

// DecisiveWinnerTeamID validates a match that cannot end tied and returns its winner.
func DecisiveWinnerTeamID(sport Sport, bestOfSets, pointsPerGame int, homeTeamID, awayTeamID string, input MatchResultInput) (string, error) {
	if homeTeamID == "" || awayTeamID == "" || homeTeamID == awayTeamID {
		return "", ErrInvalidBracketResult
	}
	if input.Incident != nil {
		var err error
		input, err = NormalizeIncidentResult(sport, bestOfSets, pointsPerGame, input)
		if err != nil {
			return "", err
		}
	} else if SetSport(sport) {
		var err error
		input, err = NormalizeSetResult(sport, bestOfSets, pointsPerGame, input)
		if err != nil {
			return "", err
		}
	}
	result := BracketResult(input)
	if err := result.validate(sport, bestOfSets, pointsPerGame); err != nil {
		return "", err
	}
	if result.HomeScore > result.AwayScore || (result.HomeScore == result.AwayScore && *result.HomePenalties > *result.AwayPenalties) {
		return homeTeamID, nil
	}
	return awayTeamID, nil
}

func (r BracketResult) validate(sport Sport, bestOfSets, pointsPerGame int) error {
	if r.Incident != nil {
		normalized, err := NormalizeIncidentResult(sport, bestOfSets, pointsPerGame, MatchResultInput(r))
		if err != nil || r.HomePenalties != nil || r.AwayPenalties != nil || normalized.HomeScore != r.HomeScore || normalized.AwayScore != r.AwayScore || !slices.Equal(normalized.Sets, r.Sets) {
			return ErrInvalidBracketResult
		}
		return nil
	}
	if r.HomeScore < 0 || r.AwayScore < 0 {
		return ErrInvalidBracketResult
	}
	if SetSport(sport) {
		normalized, err := NormalizeSetResult(sport, bestOfSets, pointsPerGame, MatchResultInput(r))
		if err != nil || normalized.HomeScore != r.HomeScore || normalized.AwayScore != r.AwayScore {
			return ErrInvalidBracketResult
		}
		return nil
	}
	if len(r.Sets) != 0 {
		return ErrInvalidBracketResult
	}
	if sport == SportBasketball {
		if r.HomeScore == r.AwayScore || r.HomePenalties != nil || r.AwayPenalties != nil {
			return ErrInvalidBracketResult
		}
		return nil
	}
	// An empty sport is retained as the internal default for brackets generated
	// before their tournament policy is attached; it follows football semantics.
	if sport != "" && sport != SportFootball && sport != SportHandball {
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
			if err := match.Result.validate(b.Sport, b.BestOfSets, b.PointsPerGame); err != nil {
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
	if err := result.validate(b.Sport, b.BestOfSets, b.PointsPerGame); err != nil {
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
	next := Bracket{Size: b.Size, Sport: b.Sport, BestOfSets: b.BestOfSets, PointsPerGame: b.PointsPerGame, Matches: make([]BracketMatch, len(b.Matches))}
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
	result.Incident = cloneIncident(result.Incident)
	result.Sets = append([]SetScore(nil), result.Sets...)
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
