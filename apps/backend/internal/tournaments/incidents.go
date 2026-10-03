package tournaments

// MatchIncident records why a match ended and the actual optional partial score.
// Its partial score never contributes to standings or determines the winner.
type MatchIncident struct {
	Type             ResultType `json:"type"`
	Side             string     `json:"side"`
	PartialHomeScore *int       `json:"partialHomeScore,omitempty"`
	PartialAwayScore *int       `json:"partialAwayScore,omitempty"`
	PartialSets      []SetScore `json:"partialSets,omitempty"`
}

// NormalizeIncidentResult applies the accepted fixed outcome independently of partial scores.
func NormalizeIncidentResult(sport Sport, bestOfSets, pointsPerGame int, input MatchResultInput) (MatchResultInput, error) {
	i := input.Incident
	if !validIncidentShape(i) || !ValidSport(sport) {
		return MatchResultInput{}, ErrInvalidBracketResult
	}
	hasScore := i.PartialHomeScore != nil || i.PartialAwayScore != nil
	if i.Type == ResultNoShow && (hasScore || len(i.PartialSets) != 0) {
		return MatchResultInput{}, ErrInvalidBracketResult
	}
	if SetSport(sport) {
		if hasScore || !validPartialSets(sport, bestOfSets, pointsPerGame, i.PartialSets) {
			return MatchResultInput{}, ErrInvalidBracketResult
		}
	} else if len(i.PartialSets) != 0 || hasScore && (i.PartialHomeScore == nil || i.PartialAwayScore == nil || *i.PartialHomeScore < 0 || *i.PartialAwayScore < 0 || *i.PartialHomeScore > 2147483647 || *i.PartialAwayScore > 2147483647) {
		return MatchResultInput{}, ErrInvalidBracketResult
	}
	winning, err := AdministrativeWinningScore(sport)
	if RacketSport(sport) {
		winning, err = bestOfSets/2+1, nil
	}
	if err != nil {
		return MatchResultInput{}, err
	}
	result := MatchResultInput{Incident: cloneIncident(i)}
	if i.Side == "home" {
		result.AwayScore = winning
	} else {
		result.HomeScore = winning
	}
	if sport == SportVolleyball {
		result.Sets = AdministrativeVolleyballSets(i.Side == "away")
	}
	return result, nil
}

func validPartialSets(sport Sport, bestOfSets, pointsPerGame int, sets []SetScore) bool {
	if !ValidSetFormat(sport, bestOfSets) || !ValidPointsPerGame(sport, pointsPerGame) || len(sets) > bestOfSets {
		return false
	}
	home, away, needed := 0, 0, bestOfSets/2+1
	for index, set := range sets {
		if home == needed || away == needed || set.HomeScore < 0 || set.AwayScore < 0 || set.HomeScore > MaximumSetScore || set.AwayScore > MaximumSetScore {
			return false
		}
		if validSportSet(sport, pointsPerGame, index, set) {
			if set.HomeScore > set.AwayScore {
				home++
			} else {
				away++
			}
		} else if index != len(sets)-1 || !validUnfinishedSet(sport, pointsPerGame, index, set) {
			return false
		}
	}
	return home < needed && away < needed
}

func validUnfinishedSet(sport Sport, pointsPerGame, index int, set SetScore) bool {
	winner, loser := max(set.HomeScore, set.AwayScore), min(set.HomeScore, set.AwayScore)
	if sport == SportTennis || sport == SportPadel {
		return winner <= 5 || winner == 6 && loser >= 5
	}
	target, pointCap := 11, MaximumSetScore
	if sport == SportVolleyball {
		target = 25
		if index == 4 {
			target = 15
		}
	}
	if sport == SportBadminton {
		target = pointsPerGame
		pointCap = 30
		if pointsPerGame == 15 {
			pointCap = 21
		}
	}
	return winner < pointCap && (winner < target || winner-loser <= 1)
}

func cloneIncident(i *MatchIncident) *MatchIncident {
	if i == nil {
		return nil
	}
	snapshot := *i
	snapshot.PartialSets = append([]SetScore(nil), i.PartialSets...)
	if i.PartialHomeScore != nil {
		value := *i.PartialHomeScore
		snapshot.PartialHomeScore = &value
	}
	if i.PartialAwayScore != nil {
		value := *i.PartialAwayScore
		snapshot.PartialAwayScore = &value
	}
	return &snapshot
}

// MatchResultType returns the persisted kind for a normalized outcome.
func MatchResultType(input MatchResultInput) ResultType {
	if input.Incident != nil {
		return input.Incident.Type
	}
	return ResultPlayed
}

func validIncidentShape(i *MatchIncident) bool {
	if i == nil || (i.Type != ResultNoShow && i.Type != ResultRetirement) || (i.Side != "home" && i.Side != "away") {
		return false
	}
	hasScore := i.PartialHomeScore != nil || i.PartialAwayScore != nil
	if i.Type == ResultNoShow && (hasScore || len(i.PartialSets) != 0) || len(i.PartialSets) > 7 || hasScore && len(i.PartialSets) != 0 {
		return false
	}
	if hasScore && (i.PartialHomeScore == nil || i.PartialAwayScore == nil || *i.PartialHomeScore < 0 || *i.PartialAwayScore < 0 || *i.PartialHomeScore > 2147483647 || *i.PartialAwayScore > 2147483647) {
		return false
	}
	for _, set := range i.PartialSets {
		if set.HomeScore < 0 || set.AwayScore < 0 || set.HomeScore > MaximumSetScore || set.AwayScore > MaximumSetScore {
			return false
		}
	}
	return true
}
