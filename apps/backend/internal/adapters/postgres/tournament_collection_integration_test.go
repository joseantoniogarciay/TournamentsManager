package postgres

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationRecentTournamentsOrdersActivityAndDeduplicatesRelationships(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "correct password")
	otherAccountID := createVerifiedLocalAccount(t, ctx, pool, "other@example.test", "other", "correct password")
	var administeredID, followedID, newestID, oldestID string
	if err := pool.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Administrada', now(), now() - interval '2 hours') RETURNING id::text`, accountID).Scan(&administeredID); err != nil {
		t.Fatalf("crear liga administrada: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Seguida', now(), now() - interval '1 hour') RETURNING id::text`, otherAccountID).Scan(&followedID); err != nil {
		t.Fatalf("crear liga seguida: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Más reciente', now(), now()) RETURNING id::text`, accountID).Scan(&newestID); err != nil {
		t.Fatalf("crear liga más reciente: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Más antigua', now(), now() - interval '3 hours') RETURNING id::text`, accountID).Scan(&oldestID); err != nil {
		t.Fatalf("crear liga más antigua: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tournament_followers (tournament_id, account_id) VALUES ($1, $2), ($3, $2)`, administeredID, accountID, followedID); err != nil {
		t.Fatalf("seguir ligas: %v", err)
	}

	items, err := tournaments.NewService(NewAccountTournamentRepository(pool)).ListRecent(ctx, accountID)
	if err != nil {
		t.Fatalf("consultar recientes: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("ligas recientes = %#v, se esperaban tres sin duplicados", items)
	}
	if items[0].ID != newestID || items[0].Relationship != "organizer" {
		t.Fatalf("primera liga = %#v, se esperaba la administrada más reciente", items[0])
	}
	if items[1].ID != followedID || items[1].Relationship != "follower" {
		t.Fatalf("segunda liga = %#v, se esperaba la seguida", items[1])
	}
	if items[2].ID != administeredID || items[2].Relationship != "organizer" {
		t.Fatalf("tercera liga = %#v, se esperaba la administrada una sola vez", items[2])
	}
	for _, item := range items {
		if item.ID == oldestID {
			t.Fatalf("liga más antigua incluida pese al límite de tres: %#v", items)
		}
	}
}

func TestIntegrationTournamentPagesIncludeEveryBoundaryItem(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "pagination@example.test", "pagination", "correct password")
	for range 5 {
		if _, err := pool.Exec(ctx, `INSERT INTO tournaments (organizer_account_id, name, published_at) VALUES ($1, 'Página', now())`, accountID); err != nil {
			t.Fatalf("create fixture: %v", err)
		}
	}
	service := tournaments.NewService(NewAccountTournamentRepository(pool))
	seen := make(map[string]bool)
	cursor := ""
	for pageNumber := 0; pageNumber < 3; pageNumber++ {
		page, err := service.List(ctx, accountID, tournaments.Administered, cursor, 2)
		if err != nil {
			t.Fatalf("page %d: %v", pageNumber, err)
		}
		for _, item := range page.Items {
			if seen[item.ID] {
				t.Fatalf("repeated item %s", item.ID)
			}
			seen[item.ID] = true
		}
		if pageNumber < 2 && (len(page.Items) != 2 || page.NextCursor != page.Items[len(page.Items)-1].ID) {
			t.Fatalf("page %d: cursor must equal the last returned item", pageNumber)
		}
		if pageNumber == 2 && (len(page.Items) != 1 || page.NextCursor != "") {
			t.Fatal("terminal page must have one item and no cursor")
		}
		cursor = page.NextCursor
	}
	if len(seen) != 5 {
		t.Fatalf("returned %d items, want all 5", len(seen))
	}
}
