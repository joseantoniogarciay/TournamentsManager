package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	httpadapter "github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/sessions"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func managementRefreshHash(token string) []byte {
	hash := sha256.Sum256([]byte("refresh:" + token))
	return hash[:]
}
func createManagementSession(t *testing.T, pool *pgxpool.Pool, accountID, token string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(t.Context(), `INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at) VALUES ($1,$2,now()+interval '7 days',now()+interval '7 days') RETURNING id::text`, accountID, sessionHash(token)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), `INSERT INTO session_refresh_tokens(session_id,token_hash,expires_at) VALUES ($1,$2,now()+interval '30 days')`, id, managementRefreshHash(token)); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestIntegrationSessionManagementIsolationRotationAndRevocation(t *testing.T) {
	ctx := t.Context()
	pool := integrationPool(t)
	account := createVerifiedLocalAccount(ctx, t, pool, "sessions@example.test", "sessions", "correct password")
	foreign := createVerifiedLocalAccount(ctx, t, pool, "foreign@example.test", "foreign", "correct password")
	current := createManagementSession(t, pool, account, "current")
	other := createManagementSession(t, pool, account, "other")
	expired := createManagementSession(t, pool, account, "expired-access")
	foreignID := createManagementSession(t, pool, foreign, "foreign")
	if _, err := pool.Exec(ctx, `UPDATE sessions SET created_at=now()-interval '8 days', last_seen_at=now()-interval '1 day', idle_expires_at=now()-interval '1 hour',absolute_expires_at=now()-interval '1 hour' WHERE id=$1`, expired); err != nil {
		t.Fatal(err)
	}
	repo := NewAccountTournamentRepository(pool)
	if id, err := repo.AuthenticateWithClient(ctx, "current", sessions.Client{Platform: "ios", Kind: "app"}); err != nil || id != account {
		t.Fatalf("auth %s %v", id, err)
	}
	page, err := repo.ListSessions(ctx, "current", "")
	if err != nil || len(page.Items) != 3 {
		t.Fatalf("list %#v %v", page, err)
	}
	var currentSeen bool
	for _, item := range page.Items {
		if item.ID == foreignID {
			t.Fatal("foreign session leaked")
		}
		if item.ID == current {
			currentSeen = item.Current && item.Platform == "ios" && item.Client == "app"
		}
	}
	if !currentSeen {
		t.Fatal("missing current session/categories")
	}
	var before, after string
	if err := pool.QueryRow(ctx, `SELECT last_seen_at::text FROM sessions WHERE id=$1`, current).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AuthenticateWithClient(ctx, "current", sessions.Client{Platform: "android", Kind: "app"}); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT last_seen_at::text FROM sessions WHERE id=$1`, current).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("activity not throttled")
	}
	registrationRepo := NewRegistrationRepository(pool)
	if _, err := registrationRepo.RotateSessionTokens(ctx, managementRefreshHash("other"), sessionHash("other-next"), managementRefreshHash("other-next")); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{foreignID, current, "019abcde-1111-7111-8111-111111111111"} {
		if err := repo.RevokeOtherSessions(ctx, "current", target); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repo.Authenticate(ctx, "foreign"); err != nil {
		t.Fatal("foreign revoked", err)
	}
	if _, err := repo.Authenticate(ctx, "current"); err != nil {
		t.Fatal("current revoked", err)
	}
	if err := repo.RevokeOtherSessions(ctx, "current", other); err != nil {
		t.Fatal(err)
	}
	if err := repo.RevokeOtherSessions(ctx, "current", other); err != nil {
		t.Fatal("non-idempotent", err)
	}
	if _, err := repo.Authenticate(ctx, "other-next"); !errors.Is(err, tournaments.ErrUnauthenticated) {
		t.Fatal("rotated access still works", err)
	}
	if _, err := registrationRepo.RotateSessionTokens(ctx, managementRefreshHash("other-next"), sessionHash("stolen"), managementRefreshHash("stolen")); !errors.Is(err, registration.ErrRefreshInvalid) {
		t.Fatal("revoked refresh still works", err)
	}
	if err := repo.RevokeOtherSessions(ctx, "current", ""); err != nil {
		t.Fatal(err)
	}
	page, err = repo.ListSessions(ctx, "current", "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != current {
		t.Fatalf("others %#v %v", page, err)
	}
	var unrevoked int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM session_refresh_tokens WHERE session_id IN ($1,$2) AND revoked_at IS NULL`, other, expired).Scan(&unrevoked); err != nil || unrevoked != 0 {
		t.Fatal("refresh family not revoked", unrevoked, err)
	}
	if _, err := registrationRepo.RotateSessionTokens(ctx, managementRefreshHash("current"), sessionHash("current-next"), managementRefreshHash("current-next")); err != nil {
		t.Fatal("current refresh revoked", err)
	}
	if err := repo.RevokeOtherSessions(ctx, "current", ""); !errors.Is(err, tournaments.ErrUnauthenticated) {
		t.Fatal("stale caller authorized", err)
	}
}

func TestIntegrationSessionManagementPaginationAndInactiveFamilies(t *testing.T) {
	ctx := t.Context()
	pool := integrationPool(t)
	account := createVerifiedLocalAccount(ctx, t, pool, "pages@example.test", "pages", "correct password")
	createManagementSession(t, pool, account, "current")
	for i := 0; i < 55; i++ {
		createManagementSession(t, pool, account, fmt.Sprintf("session-%d", i))
	}
	dead := createManagementSession(t, pool, account, "dead")
	if _, err := pool.Exec(ctx, `UPDATE sessions SET created_at=now()-interval '8 days', last_seen_at=now()-interval '1 day', idle_expires_at=now()-interval '1 hour',absolute_expires_at=now()-interval '1 hour' WHERE id=$1`, dead); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE session_refresh_tokens SET consumed_at=now() WHERE session_id=$1`, dead); err != nil {
		t.Fatal(err)
	}
	repo := NewAccountTournamentRepository(pool)
	seen := map[string]bool{}
	cursor := ""
	for {
		page, err := repo.ListSessions(ctx, "current", cursor)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) > 50 {
			t.Fatal("unbounded page")
		}
		for _, item := range page.Items {
			if seen[item.ID] || item.ID == dead {
				t.Fatal("duplicate/inactive")
			}
			seen[item.ID] = true
		}
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
	}
	if len(seen) != 56 {
		t.Fatal("sessions omitted", len(seen))
	}
}

func TestIntegrationSessionManagementConcurrentRefreshAndRevoke(t *testing.T) {
	ctx := t.Context()
	pool := integrationPool(t)
	account := createVerifiedLocalAccount(ctx, t, pool, "race@example.test", "race", "correct password")
	createManagementSession(t, pool, account, "current")
	repo := NewAccountTournamentRepository(pool)
	registrationRepo := NewRegistrationRepository(pool)
	for i := 0; i < 20; i++ {
		token := fmt.Sprintf("race-%d", i)
		next := token + "-next"
		id := createManagementSession(t, pool, account, token)
		gate := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		var refreshErr, revokeErr error
		go func() {
			defer wg.Done()
			<-gate
			_, refreshErr = registrationRepo.RotateSessionTokens(ctx, managementRefreshHash(token), sessionHash(next), managementRefreshHash(next))
		}()
		go func() { defer wg.Done(); <-gate; revokeErr = repo.RevokeOtherSessions(ctx, "current", id) }()
		close(gate)
		wg.Wait()
		if revokeErr != nil || (refreshErr != nil && !errors.Is(refreshErr, registration.ErrRefreshInvalid)) {
			t.Fatalf("race: revoke %v refresh %v", revokeErr, refreshErr)
		}
		if _, err := repo.Authenticate(ctx, next); !errors.Is(err, tournaments.ErrUnauthenticated) {
			t.Fatal("race resurrected access", err)
		}
		if _, err := registrationRepo.RotateSessionTokens(ctx, managementRefreshHash(next), sessionHash(next+"-again"), managementRefreshHash(next+"-again")); !errors.Is(err, registration.ErrRefreshInvalid) {
			t.Fatal("race resurrected refresh", err)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := repo.RevokeOtherSessions(cancelled, "current", ""); !errors.Is(err, context.Canceled) {
		t.Fatal("cancel", err)
	}
	if _, err := repo.Authenticate(ctx, "current"); err != nil {
		t.Fatal("cancel changed current", err)
	}
}

func TestIntegrationSessionManagementHTTPAccessAndRefresh(t *testing.T) {
	ctx := t.Context()
	pool := integrationPool(t)
	account := createVerifiedLocalAccount(ctx, t, pool, "http@example.test", "httptest", "correct password")
	createManagementSession(t, pool, account, "current-http")
	target := createManagementSession(t, pool, account, "other-http")
	repo := NewAccountTournamentRepository(pool)
	handler := httpadapter.NewHandlerWithConfig(httpadapter.HandlerConfig{}, httpadapter.HandlerDependencies{Authenticator: repo, Registration: registration.NewService(NewRegistrationRepository(pool), nil)})
	request := func(method, path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(ctx, method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-TM-Client-Platform", "android")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	rec := request("GET", "/v1/me/sessions", "current-http")
	var page sessions.Page
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &page) != nil || len(page.Items) != 2 {
		t.Fatalf("HTTP list status %d", rec.Code)
	}
	for _, item := range page.Items {
		if item.Current && (item.Platform != "android" || item.Client != "app") {
			t.Fatal("metadata did not persist")
		}
	}
	if rec := request("DELETE", "/v1/me/sessions/"+target, "current-http"); rec.Code != 204 {
		t.Fatal("HTTP revoke", rec.Code)
	}
	if rec := request("GET", "/v1/sessions", "other-http"); rec.Code != 401 {
		t.Fatal("revoked HTTP access", rec.Code)
	}
	if rec := request("POST", "/v1/sessions/refresh", "other-http"); rec.Code != 401 {
		t.Fatal("revoked HTTP refresh", rec.Code)
	}
	if rec := request("GET", "/v1/sessions", "current-http"); rec.Code != 200 {
		t.Fatal("current HTTP access", rec.Code)
	}
	if rec := request("POST", "/v1/sessions/refresh", "current-http"); rec.Code != 200 {
		t.Fatal("current HTTP refresh", rec.Code)
	}
}
