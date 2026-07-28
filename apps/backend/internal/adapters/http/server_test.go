package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type testAuthenticator struct{ accountID string }

func (a testAuthenticator) Authenticate(context.Context, string) (string, error) {
	if a.accountID == "" {
		return "", leagues.ErrUnauthenticated
	}
	return a.accountID, nil
}

type testLeagueRepository struct {
	items         []leagues.Item
	followVisible bool
}

func (r testLeagueRepository) List(_ context.Context, _ string, _ leagues.Relationship, _ string, _ int) ([]leagues.Item, error) {
	return r.items, nil
}

func (r testLeagueRepository) Follow(_ context.Context, _ string, _ string) (bool, error) {
	return r.followVisible, nil
}

func (r testLeagueRepository) Unfollow(context.Context, string, string) error { return nil }

func testHandler() http.Handler {
	return NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, leagues.NewService(testLeagueRepository{}))
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestListAccountLeaguesRequiresSession(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/leagues?relationship=administered", nil)
	recorder := httptest.NewRecorder()
	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestListAccountLeaguesRejectsCookieAndBearerTogether(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/leagues?relationship=administered", nil)
	request.Header.Set("Authorization", "Bearer opaque-session")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "other-session"})
	recorder := httptest.NewRecorder()
	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestListAccountLeaguesReturnsPage(t *testing.T) {
	t.Parallel()

	items := []leagues.Item{{ID: "019abcde-1111-7111-8111-111111111111", Name: "Liga", State: "published", CreatedAt: "2026-07-28T10:00:00Z", Relationship: "organizer"}}
	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{items: items}))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/leagues?relationship=administered&limit=1", nil)
	request.Header.Set("Authorization", "Bearer opaque-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(body), `"relationship":"organizer"`) {
		t.Errorf("body = %s, want organizer relationship", body)
	}
}

func TestFollowLeagueRejectsCrossSiteCookieRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/leagues/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Origin", "https://evil.example")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "web-session"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestFollowLeagueAllowsBearerRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/leagues/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Authorization", "Bearer mobile-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
