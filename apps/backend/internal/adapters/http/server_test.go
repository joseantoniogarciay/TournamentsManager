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

var testAllowedOrigins = []string{"http://localhost:8081"}

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

type testRegistrationRepository struct{ available bool }

func (r testRegistrationRepository) CreatePending(context.Context, registration.Input, string, []byte) (bool, error) {
	return false, nil
}

func (r testRegistrationRepository) IsUsernameAvailable(context.Context, string) (bool, error) {
	return r.available, nil
}

func (r testRegistrationRepository) VerifyAndCreateSession(context.Context, []byte, []byte, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrVerificationInvalid
}

func (r testRegistrationRepository) RotateSessionTokens(context.Context, []byte, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrRefreshInvalid
}

func testHandler() http.Handler {
	return NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
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

func TestCORSPreflightAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/v1/registrations", nil)
	request.Header.Set("Origin", "http://localhost:8081")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Errorf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", got)
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "https://evil.example")
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUsernameAvailabilityReturnsCurrentAvailability(t *testing.T) {
	t.Parallel()

	registrationService := registration.NewService(testRegistrationRepository{available: false}, nil)
	handler := NewHandler(registrationService, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/usernames/already_taken/availability", nil)
	request.Header.Set("Origin", "http://localhost:8081")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Errorf("Access-Control-Allow-Origin = %q, want http://localhost:8081", got)
	}
	body, _ := io.ReadAll(recorder.Result().Body)
	if !strings.Contains(string(body), `"available":false`) {
		t.Errorf("body = %s, want available false", body)
	}
}

func TestUsernameAvailabilityRejectsInvalidUsername(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/usernames/No/availability", nil)
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUsernameAvailabilityRateLimitsByClientIP(t *testing.T) {
	t.Parallel()

	registrationService := registration.NewService(testRegistrationRepository{available: true}, nil)
	handler := NewHandler(registrationService, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	for range usernameAvailabilityLimit {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/usernames/available_name/availability", nil)
		request.RemoteAddr = "203.0.113.1:10000"
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d before limit, want %d", recorder.Code, http.StatusOK)
		}
	}

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/usernames/available_name/availability", nil)
	request.RemoteAddr = "203.0.113.1:10000"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header is missing")
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
	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{items: items}), testAllowedOrigins)
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

	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}), testAllowedOrigins)
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

	handler := NewHandler(registration.Service{}, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/leagues/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Authorization", "Bearer mobile-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
