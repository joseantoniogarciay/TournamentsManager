package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestListAccountTournamentsRequiresSession(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/tournaments?relationship=administered", nil)
	recorder := httptest.NewRecorder()
	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestListAccountTournamentsRejectsCookieAndBearerTogether(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/tournaments?relationship=administered", nil)
	request.Header.Set("Authorization", "Bearer opaque-session")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "other-session"})
	recorder := httptest.NewRecorder()
	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRevokeCurrentSessionExpiresCookie(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/v1/sessions", nil)
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "opaque-session"})
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 2 || cookies[0].MaxAge >= 0 || cookies[1].MaxAge >= 0 {
		t.Errorf("logout cookies = %#v, want expired access and refresh cookies", cookies)
	}
}

func TestSessionCookiesPersistUntilTheirOwnExpirations(t *testing.T) {
	t.Parallel()

	accessExpiry := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
	recorder := httptest.NewRecorder()
	sessionCookies(true).setSession(recorder, "access", "refresh", registration.Session{IdleExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry})

	cookies := recorder.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies = %#v, want access and refresh", cookies)
	}
	if cookies[0].Name != "__Host-tm_session" || cookies[0].MaxAge <= 0 || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Errorf("access cookie = %#v, want persistent secure HttpOnly access cookie", cookies[0])
	}
	if cookies[1].Name != "__Host-tm_refresh" || cookies[1].MaxAge <= cookies[0].MaxAge || !cookies[1].HttpOnly || !cookies[1].Secure {
		t.Errorf("refresh cookie = %#v, want longer persistent secure HttpOnly refresh cookie", cookies[1])
	}
}

func TestRefreshCookieRejectsCrossSiteRequest(t *testing.T) {
	t.Parallel()

	handler := refreshCookieCSRF(testAllowedOrigins, sessionCookies(true), http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/v1/sessions/refresh", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_refresh", Value: "refresh"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestListAccountTournamentsReturnsPage(t *testing.T) {
	t.Parallel()

	items := []tournaments.Item{{ID: "019abcde-1111-7111-8111-111111111111", Name: "Liga", State: "published", CreatedAt: "2026-07-28T10:00:00Z", Relationship: "organizer"}}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, tournaments.NewService(testTournamentRepository{items: items}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/tournaments?relationship=administered&limit=1", nil)
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

func TestListRecentAccountTournamentsReturnsSummary(t *testing.T) {
	t.Parallel()

	items := []tournaments.Item{{ID: "019abcde-1111-7111-8111-111111111111", Name: "Liga", State: "in_progress", CreatedAt: "2026-07-28T10:00:00Z", LastActivityAt: "2026-08-02T10:00:00Z", Relationship: "organizer"}}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, tournaments.NewService(testTournamentRepository{recentItems: items}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/recent-tournaments", nil)
	request.Header.Set("Authorization", "Bearer opaque-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(body), `"lastActivityAt":"2026-08-02T10:00:00Z"`) {
		t.Errorf("body = %s, want recent activity", body)
	}
}

func TestFollowTournamentRejectsCrossSiteCookieRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, tournaments.NewService(testTournamentRepository{followVisible: true}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/tournaments/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Origin", "https://evil.example")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "web-session"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestFollowTournamentAllowsBearerRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, tournaments.NewService(testTournamentRepository{followVisible: true}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/tournaments/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Authorization", "Bearer mobile-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
