package tournaments

import "sort"

// AdministrativeVolleyballSets retains the ordered points of the fixed 3-0 result.
func AdministrativeVolleyballSets(homeWins bool) []SetScore {
	sets := make([]SetScore, 3)
	for index := range sets {
		if homeWins {
			sets[index].HomeScore = 25
		} else {
			sets[index].AwayScore = 25
		}
	}
	return sets
}

func calculateVolleyballStandings(tournament Tournament) []Standing {
	rows := make(map[string]*Standing, len(tournament.Teams))
	for _, team := range tournament.Teams {
		rows[team.ID] = &Standing{TeamID: team.ID}
	}
	for _, match := range tournament.Matches {
		if match.State != "completed" || match.HomeScore == nil || match.AwayScore == nil {
			continue
		}
		home, away := rows[match.HomeTeamID], rows[match.AwayTeamID]
		if home == nil || away == nil {
			continue
		}
		home.Played++
		away.Played++
		home.ScoreFor += *match.HomeScore
		home.ScoreAgainst += *match.AwayScore
		away.ScoreFor += *match.AwayScore
		away.ScoreAgainst += *match.HomeScore
		for _, set := range match.Sets {
			home.RallyPointsFor += set.HomeScore
			home.RallyPointsAgainst += set.AwayScore
			away.RallyPointsFor += set.AwayScore
			away.RallyPointsAgainst += set.HomeScore
		}
		winner, loser, losingSets := home, away, *match.AwayScore
		if *match.HomeScore < *match.AwayScore {
			winner, loser, losingSets = away, home, *match.HomeScore
		}
		winner.Won++
		loser.Lost++
		if losingSets == 2 {
			winner.Points += 2
			loser.Points++
		} else {
			winner.Points += 3
		}
	}
	result := make([]Standing, 0, len(tournament.Teams))
	for _, team := range tournament.Teams {
		row := rows[team.ID]
		row.ScoreDifference = row.ScoreFor - row.ScoreAgainst
		result = append(result, *row)
	}
	sort.SliceStable(result, func(i, j int) bool { return compareVolleyballStanding(result[i], result[j]) > 0 })
	position := 1
	for index := range result {
		if index > 0 && compareVolleyballStanding(result[index-1], result[index]) != 0 {
			position = index + 1
		}
		result[index].Position = position
	}
	return result
}

func compareVolleyballStanding(left, right Standing) int {
	if left.Won != right.Won {
		return left.Won - right.Won
	}
	if left.Points != right.Points {
		return left.Points - right.Points
	}
	if result := compareRatio(left.ScoreFor, left.ScoreAgainst, right.ScoreFor, right.ScoreAgainst); result != 0 {
		return result
	}
	return compareRatio(left.RallyPointsFor, left.RallyPointsAgainst, right.RallyPointsFor, right.RallyPointsAgainst)
}

// Compare integer ratios exactly: positive/0 is infinity, 0/0 is zero.
// A league has at most 126 matches per team and each set score fits smallint;
// cross-products of the accumulated point totals therefore fit int64.
func compareRatio(leftNum, leftDen, rightNum, rightDen int) int {
	leftInfinite, rightInfinite := leftNum > 0 && leftDen == 0, rightNum > 0 && rightDen == 0
	if leftInfinite || rightInfinite {
		if leftInfinite == rightInfinite {
			return 0
		}
		if leftInfinite {
			return 1
		}
		return -1
	}
	if leftDen == 0 {
		leftDen = 1
	}
	if rightDen == 0 {
		rightDen = 1
	}
	left, right := int64(leftNum)*int64(rightDen), int64(rightNum)*int64(leftDen)
	if left > right {
		return 1
	}
	if left < right {
		return -1
	}
	return 0
}
