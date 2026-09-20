package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestUsernameAvailabilityReturnsCurrentAvailability(t *testing.T) {
	t.Parallel()

	registrationService := registration.NewService(testRegistrationRepository{available: false}, nil)
	handler := NewHandler(registrationService, nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/usernames/already_taken/availability", nil)
	request.Header.Set("Origin", "http://localhost:8082")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8082" {
		t.Errorf("Access-Control-Allow-Origin = %q, want http://localhost:8082", got)
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
	handler := NewHandler(registrationService, nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
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

func TestClientIPUsesForwardedAddressOnlyFromTrustedProxy(t *testing.T) {
	t.Parallel()

	trusted := newClientIPResolver([]netip.Prefix{netip.MustParsePrefix("192.168.65.0/24")}, "")
	for _, test := range []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{"trusted docker proxy", "192.168.65.1:54321", "203.0.113.8", "203.0.113.8"},
		{"untrusted peer cannot spoof", "198.51.100.2:54321", "203.0.113.8", "198.51.100.2"},
		{"trusted proxy with invalid header", "192.168.65.1:54321", "not-an-ip", "192.168.65.1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = test.remoteAddr
			request.Header.Set("X-Client-IP", test.forwarded)
			if got := trusted(request); got != test.want {
				t.Errorf("client IP = %q, want %q", got, test.want)
			}
		})
	}
}

func TestClientIPUsesForwardedAddressWithValidEdgeToken(t *testing.T) {
	t.Parallel()

	resolver := newClientIPResolver(nil, "edge-token-for-test")
	for _, test := range []struct {
		name      string
		forwarded string
		token     string
		want      string
	}{
		{"valid token", "203.0.113.8", "edge-token-for-test", "203.0.113.8"},
		{"missing token", "203.0.113.8", "", "10.42.0.9"},
		{"invalid token", "203.0.113.8", "wrong-token", "10.42.0.9"},
		{"invalid forwarded address", "not-an-ip", "edge-token-for-test", "10.42.0.9"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = "10.42.0.9:54321"
			request.Header.Set("X-Client-IP", test.forwarded)
			if test.token != "" {
				request.Header.Set("X-FastTourney-Edge-Token", test.token)
			}
			if got := resolver(request); got != test.want {
				t.Errorf("client IP = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRegistrationRequiresSupportedLocale(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.NewService(testRegistrationRepository{}, nil), nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
	for _, test := range []struct {
		name   string
		body   string
		status int
	}{
		{"supported", `{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"fr","termsVersion":"2026-08-22"}`, http.StatusAccepted},
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

func TestValidRegistrationDraftEnforcesTournamentNameCharacterLimit(t *testing.T) {
	t.Parallel()

	draft := &registration.Draft{
		ID:    "019abcde-1111-7111-8111-111111111112",
		Name:  strings.Repeat("a", tournaments.MaximumTournamentNameLength),
		Sport: tournaments.SportFootball,
		Teams: []string{"Azules", "Rojos"},
	}
	if !validRegistrationDraft(draft) {
		t.Fatalf("validRegistrationDraft() rejected %d characters", tournaments.MaximumTournamentNameLength)
	}
	draft.ID = ""
	if validRegistrationDraft(draft) {
		t.Error("validRegistrationDraft() accepted a transferred draft without draftId")
	}
	draft.ID = "019abcde-1111-7111-8111-111111111112"
	draft.Name += "a"
	if validRegistrationDraft(draft) {
		t.Errorf("validRegistrationDraft() accepted %d characters", tournaments.MaximumTournamentNameLength+1)
	}
}

func TestRegistrationRateLimitsByClientIP(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.NewService(testRegistrationRepository{}, nil), nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
	for range registrationLimit {
		request := httptest.NewRequest(http.MethodPost, "/v1/registrations", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"es","termsVersion":"2026-08-22"}`))
		request.RemoteAddr = "203.0.113.1:10000"
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusAccepted {
			t.Fatalf("status = %d before limit, want %d", recorder.Code, http.StatusAccepted)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/registrations", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"es","termsVersion":"2026-08-22"}`))
	request.RemoteAddr = "203.0.113.1:10000"
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header is missing")
	}
}
