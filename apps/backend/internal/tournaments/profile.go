package tournaments

import (
	"context"
	"math"
	"strings"
	"unicode/utf8"
)

// SportStatistics aggregates confirmed sporting outcomes, independently of pagination.
type SportStatistics struct {
	Sport         Sport    `json:"sport"`
	Played        int      `json:"played"`
	Wins          int      `json:"wins"`
	Draws         int      `json:"draws"`
	Losses        int      `json:"losses"`
	DrawsAllowed  bool     `json:"drawsAllowed"`
	WinPercentage *float64 `json:"winPercentage"`
}

// ProfileMatch identifies the account's side without exposing account identifiers.
type ProfileMatch struct {
	PersonalMatch
	OwnSide string `json:"ownSide"`
}

// PersonalProfile is accessible only to the authenticated account.
type PersonalProfile struct {
	Statistics []SportStatistics `json:"statistics"`
	Items      []ProfileMatch    `json:"items"`
	NextCursor string            `json:"nextCursor,omitempty"`
}

// PersonalProfile returns bounded history and complete, filtered statistics.
func (s BrowsingService) PersonalProfile(ctx context.Context, account, sport, tournament, cursor string, limit int) (PersonalProfile, error) {
	limit, err := browsingLimit(limit)
	tournament = strings.TrimSpace(tournament)
	if err != nil || (sport != "" && !ValidSport(Sport(sport))) || utf8.RuneCountInString(tournament) > 120 || strings.ContainsAny(tournament, "\r\n\x00") {
		return PersonalProfile{}, ErrInvalidPage
	}
	page, err := s.repository.PersonalProfile(ctx, account, sport, tournament, cursor, limit+1)
	if err != nil {
		return PersonalProfile{}, err
	}
	for i := range page.Statistics {
		stat := &page.Statistics[i]
		stat.WinPercentage = nil
		if stat.Played > 0 {
			rate := math.Round(float64(stat.Wins)*1000/float64(stat.Played)) / 10
			stat.WinPercentage = &rate
		}
	}
	if len(page.Items) > limit {
		page.NextCursor = page.Items[limit-1].Match.ID
		page.Items = page.Items[:limit]
	}
	return page, nil
}
