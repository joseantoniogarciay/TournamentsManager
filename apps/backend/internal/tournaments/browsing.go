package tournaments

import "context"

// PersonalMatch is a read projection; following or administering is not participation.
type PersonalMatch struct {
	TournamentID   string `json:"tournamentId"`
	TournamentName string `json:"tournamentName"`
	Sport          Sport  `json:"sport"`
	StageType      string `json:"stageType"`
	HomeTeamName   string `json:"homeTeamName"`
	AwayTeamName   string `json:"awayTeamName"`
	Match          Match  `json:"match"`
}

// PersonalMatchPage includes global existence independently of the selected view.
type PersonalMatchPage struct {
	Items      []PersonalMatch `json:"items"`
	HasMatches bool            `json:"hasMatches"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

// Edition summarizes an independently managed tournament in a repetition family.
type Edition struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	State     string   `json:"state"`
	CreatedAt string   `json:"createdAt"`
	Champions []string `json:"champions"`
}

// EditionPage contains a stable page of related tournament editions.
type EditionPage struct {
	Items      []Edition `json:"items"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// BrowsingRepository projects collections without changing the tournament lifecycle.
type BrowsingRepository interface {
	PersonalMatches(context.Context, string, string, string, int) (PersonalMatchPage, error)
	Editions(context.Context, string, string, int) ([]Edition, error)
}

// BrowsingService validates and bounds tournament read collections.
type BrowsingService struct{ repository BrowsingRepository }

// NewBrowsingService uses a read port without introducing infrastructure into business rules.
func NewBrowsingService(repository BrowsingRepository) BrowsingService {
	return BrowsingService{repository}
}

func browsingLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultLimit, nil
	}
	if limit < 1 || limit > maximumLimit {
		return 0, ErrInvalidPage
	}
	return limit, nil
}

// PersonalMatches returns actual participating encounters in the selected view.
func (s BrowsingService) PersonalMatches(ctx context.Context, account, view, cursor string, limit int) (PersonalMatchPage, error) {
	if view != "pending" && view != "results" {
		return PersonalMatchPage{}, ErrInvalidPage
	}
	limit, err := browsingLimit(limit)
	if err != nil {
		return PersonalMatchPage{}, err
	}
	page, err := s.repository.PersonalMatches(ctx, account, view, cursor, limit+1)
	if err != nil {
		return PersonalMatchPage{}, err
	}
	if len(page.Items) > limit {
		page.NextCursor = page.Items[limit-1].Match.ID
		page.Items = page.Items[:limit]
	}
	return page, nil
}

// Editions returns the same family from any of its members.
func (s BrowsingService) Editions(ctx context.Context, id, cursor string, limit int) (EditionPage, error) {
	limit, err := browsingLimit(limit)
	if err != nil {
		return EditionPage{}, err
	}
	items, err := s.repository.Editions(ctx, id, cursor, limit+1)
	if err != nil {
		return EditionPage{}, err
	}
	page := EditionPage{Items: items}
	if len(items) > limit {
		page.NextCursor = items[limit-1].ID
		page.Items = items[:limit]
	}
	return page, nil
}
