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

func (a testAuthenticator) GetCurrentSession(context.Context, string) (leagues.CurrentSession, error) {
	if a.accountID == "" {
		return leagues.CurrentSession{}, leagues.ErrUnauthenticated
	}
	return leagues.CurrentSession{
		AccountID:         a.accountID,
		Username:          "person",
		IdleExpiresAt:     "2026-08-09T12:00:00Z",
		AbsoluteExpiresAt: "2026-08-09T12:00:00Z",
	}, nil
}

func (testAuthenticator) RevokeSession(context.Context, string) error { return nil }

type testLeagueRepository struct {
	items         []leagues.Item
	recentItems   []leagues.Item
	followVisible bool
}

func (r testLeagueRepository) List(_ context.Context, _ string, _ leagues.Relationship, _ string, _ int) ([]leagues.Item, error) {
	return r.items, nil
}

func (r testLeagueRepository) ListRecent(context.Context, string) ([]leagues.Item, error) {
	return r.recentItems, nil
}

func (r testLeagueRepository) Follow(_ context.Context, _ string, _ string) (bool, error) {
	return r.followVisible, nil
}

func (r testLeagueRepository) Unfollow(context.Context, string, string) error { return nil }

type testRegistrationRepository struct {
	available    bool
	loginAccount registration.LocalAccount
	loginSession registration.Session
}

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
func (r testRegistrationRepository) CreatePasswordReset(context.Context, string, []byte) (string, registration.Locale, bool, error) {
	return "", "", false, nil
}
func (r testRegistrationRepository) InspectPasswordReset(context.Context, []byte) (string, error) {
	return "", registration.ErrPasswordResetInvalid
}
func (r testRegistrationRepository) ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrPasswordResetInvalid
}
func (r testRegistrationRepository) FindLocalAccountForLogin(context.Context, string) (registration.LocalAccount, error) {
	if r.loginAccount.ID == "" {
		return registration.LocalAccount{}, registration.ErrLoginInvalid
	}
	return r.loginAccount, nil
}
func (r testRegistrationRepository) CreateLocalLoginSession(context.Context, string, []byte, []byte) (registration.Session, error) {
	return r.loginSession, nil
}

func TestCreateLocalSessionReturnsBearerSession(t *testing.T) {
	t.Parallel()
	passwordHash, err := registration.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("crear hash: %v", err)
	}
	repository := testRegistrationRepository{
		loginAccount: registration.LocalAccount{ID: "019abcde-1111-7111-8111-111111111111", PasswordHash: passwordHash, Verified: true},
		loginSession: registration.Session{AccountID: "019abcde-1111-7111-8111-111111111111", Username: "person", IdleExpiresAt: "2026-08-09T12:00:00Z", RefreshExpiresAt: "2026-09-01T12:00:00Z"},
	}
	handler := NewHandler(registration.NewService(repository, nil), nil, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/sessions", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","sessionTransport":"bearer"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"delivery":"bearer"`) || !strings.Contains(recorder.Body.String(), `"sessionToken"`) || !strings.Contains(recorder.Body.String(), `"refreshToken"`) {
		t.Errorf("body = %s, want bearer session tokens", recorder.Body.String())
	}
}
func (r testRegistrationRepository) RenewLoginVerification(context.Context, string, []byte) (string, registration.Locale, error) {
	return "", "", registration.ErrLoginInvalid
}

func testHandler() http.Handler {
	return NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
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

func TestGetCurrentSessionReturnsIdentity(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/sessions", nil)
	request.Header.Set("Authorization", "Bearer opaque-session")
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"id":"019abcde-1111-7111-8111-111111111111"`) || !strings.Contains(body, `"username":"person"`) {
		t.Errorf("body = %s, want current session identity", body)
	}
}

func TestGetCurrentSessionRequiresSession(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/sessions", nil)
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
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
	handler := NewHandler(registrationService, nil, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
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
	handler := NewHandler(registrationService, nil, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
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

func TestRegistrationRequiresSupportedLocale(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.NewService(testRegistrationRepository{}, nil), nil, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	for _, test := range []struct {
		name   string
		body   string
		status int
	}{
		{"supported", `{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"fr"}`, http.StatusAccepted},
		{"missing", `{"email":"person@example.test","password":"correct horse battery staple","username":"person_name"}`, http.StatusBadRequest},
		{"unsupported", `{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"de"}`, http.StatusBadRequest},
		{"non-canonical", `{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"FR"}`, http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/registrations", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
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

func TestRevokeCurrentSessionExpiresCookie(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/v1/sessions", nil)
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "opaque-session"})
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if cookie := recorder.Result().Cookies(); len(cookie) != 1 || cookie[0].MaxAge >= 0 {
		t.Errorf("logout cookie = %#v, want expired cookie", cookie)
	}
}

func TestListAccountLeaguesReturnsPage(t *testing.T) {
	t.Parallel()

	items := []leagues.Item{{ID: "019abcde-1111-7111-8111-111111111111", Name: "Liga", State: "published", CreatedAt: "2026-07-28T10:00:00Z", Relationship: "organizer"}}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{items: items}), testAllowedOrigins)
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

func TestListRecentAccountLeaguesReturnsSummary(t *testing.T) {
	t.Parallel()

	items := []leagues.Item{{ID: "019abcde-1111-7111-8111-111111111111", Name: "Liga", State: "in_progress", CreatedAt: "2026-07-28T10:00:00Z", LastActivityAt: "2026-08-02T10:00:00Z", Relationship: "organizer"}}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{recentItems: items}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/me/recent-leagues", nil)
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

func TestFollowLeagueRejectsCrossSiteCookieRequest(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}), testAllowedOrigins)
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

	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-2222-7222-8222-222222222222"}, leagues.NewService(testLeagueRepository{followVisible: true}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/me/leagues/019abcde-1111-7111-8111-111111111111/follow", nil)
	request.Header.Set("Authorization", "Bearer mobile-session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
