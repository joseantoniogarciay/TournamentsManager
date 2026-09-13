package tournaments

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type repositoryStub struct {
	items       []Item
	recentItems []Item
	limit       int
}

func (r *repositoryStub) List(_ context.Context, _ string, _ Relationship, _ string, limit int) ([]Item, error) {
	r.limit = limit
	return r.items, nil
}

func (r *repositoryStub) ListRecent(context.Context, string) ([]Item, error) {
	return r.recentItems, nil
}

func (r *repositoryStub) Follow(context.Context, string, string) (bool, error) { return true, nil }

func (r *repositoryStub) Unfollow(context.Context, string, string) error { return nil }

func TestListUsesExtraItemToBuildNextCursor(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{items: []Item{{ID: "first"}, {ID: "second"}, {ID: "third"}}}
	page, err := NewService(repository).List(context.Background(), "account", Administered, "", 2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.limit != 3 {
		t.Errorf("repository limit = %d, want 3", repository.limit)
	}
	if len(page.Items) != 2 || page.NextCursor != "third" {
		t.Errorf("page = %#v, want two items and third cursor", page)
	}
}

func TestListRejectsUnknownRelationship(t *testing.T) {
	t.Parallel()

	_, err := NewService(&repositoryStub{}).List(context.Background(), "account", "unknown", "", 20)
	if err != ErrInvalidRelationship {
		t.Errorf("List() error = %v, want %v", err, ErrInvalidRelationship)
	}
}

func TestListRecentReturnsRepositorySummary(t *testing.T) {
	t.Parallel()

	items, err := NewService(&repositoryStub{recentItems: []Item{{ID: "recent"}}}).ListRecent(context.Background(), "account")
	if err != nil || len(items) != 1 || items[0].ID != "recent" {
		t.Fatalf("ListRecent() = %#v, %v; want the repository summary", items, err)
	}
}

func TestValidCreateInputEnforcesTournamentNameCharacterLimit(t *testing.T) {
	t.Parallel()

	input := CreateInput{Sport: SportFootball, Teams: []TeamInput{{Name: "Azules"}, {Name: "Rojos"}}}
	input.Name = strings.Repeat("a", MaximumTournamentNameLength)
	if !validCreateInput(input) {
		t.Fatalf("validCreateInput() rejected %d characters", MaximumTournamentNameLength)
	}
	input.Name += "a"
	if validCreateInput(input) {
		t.Errorf("validCreateInput() accepted %d characters", MaximumTournamentNameLength+1)
	}
}

func TestValidCreateInputRejectsUnknownSport(t *testing.T) {
	t.Parallel()

	input := CreateInput{Name: "Torneo", Sport: "unknown", Teams: []TeamInput{{Name: "Azules"}, {Name: "Rojos"}}}
	if validCreateInput(input) {
		t.Fatal("validCreateInput() accepted an unknown sport")
	}
}

func TestValidateLeagueResultUsesSportPolicy(t *testing.T) {
	t.Parallel()

	if err := ValidateLeagueResult(SportFootball, MatchResultInput{HomeScore: 1, AwayScore: 1}); err != nil {
		t.Fatalf("football draw rejected: %v", err)
	}
	if err := ValidateLeagueResult(SportBasketball, MatchResultInput{HomeScore: 80, AwayScore: 80}); !errors.Is(err, ErrInvalidTournamentInput) {
		t.Fatalf("basketball draw error = %v, want invalid input", err)
	}
	penalties := 4
	if err := ValidateLeagueResult(SportBasketball, MatchResultInput{HomeScore: 80, AwayScore: 79, HomePenalties: &penalties}); !errors.Is(err, ErrInvalidTournamentInput) {
		t.Fatalf("basketball penalties error = %v, want invalid input", err)
	}
}

func TestAdministrativeWinningScoreUsesSportPolicy(t *testing.T) {
	t.Parallel()

	for sport, want := range map[Sport]int{SportFootball: 3, SportBasketball: 20} {
		got, err := AdministrativeWinningScore(sport)
		if err != nil || got != want {
			t.Errorf("AdministrativeWinningScore(%q) = %d, %v; want %d", sport, got, err, want)
		}
	}
}
