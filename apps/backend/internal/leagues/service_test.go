package leagues

import (
	"context"
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

func TestValidCreateInputEnforcesLeagueNameCharacterLimit(t *testing.T) {
	t.Parallel()

	input := CreateInput{Teams: []TeamInput{{Name: "Azules"}, {Name: "Rojos"}}}
	input.Name = strings.Repeat("a", MaximumLeagueNameLength)
	if !validCreateInput(input) {
		t.Fatalf("validCreateInput() rejected %d characters", MaximumLeagueNameLength)
	}
	input.Name += "a"
	if validCreateInput(input) {
		t.Errorf("validCreateInput() accepted %d characters", MaximumLeagueNameLength+1)
	}
}
