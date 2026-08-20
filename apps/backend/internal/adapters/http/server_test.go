package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type testAuthenticator struct{ accountID string }

type testDeletionAuthenticator struct{ testAuthenticator }

func (testDeletionAuthenticator) ScheduleAccountDeletion(context.Context, string) (time.Time, error) {
	return time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC), nil
}

var testAllowedOrigins = []string{"http://localhost:8082"}

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

type testCreationRepository struct {
	administrators    []string
	administratorsErr error
	cancelled         leagues.League
	cancelErr         error
	team              leagues.Team
	teamErr           error
	removeErr         error
	transferErr       error
	withdrawn         leagues.League
	withdrawErr       error
	result            leagues.League
	resultErr         error
	completed         leagues.League
	completeErr       error
}

func (testCreationRepository) Create(context.Context, string, leagues.CreateInput) (leagues.League, error) {
	return leagues.League{}, nil
}

func (r testCreationRepository) AddTeam(context.Context, string, string, leagues.TeamInput) (leagues.Team, error) {
	return r.team, r.teamErr
}

func (r testCreationRepository) RemoveTeam(context.Context, string, string, string) error {
	return r.removeErr
}

func (r testCreationRepository) WithdrawTeam(context.Context, string, string, string) (leagues.League, error) {
	return r.withdrawn, r.withdrawErr
}

func (testCreationRepository) Start(context.Context, string, string, leagues.StartInput) (leagues.League, error) {
	return leagues.League{}, nil
}

func (r testCreationRepository) Cancel(context.Context, string, string) (leagues.League, error) {
	return r.cancelled, r.cancelErr
}

func (testCreationRepository) AssignAdministrator(context.Context, string, string, string) error {
	return nil
}

func (r testCreationRepository) ListAdministrators(context.Context, string, string) ([]string, error) {
	return r.administrators, r.administratorsErr
}

func (r testCreationRepository) RemoveAdministrator(context.Context, string, string, string) error {
	return r.removeErr
}

func (r testCreationRepository) TransferOwnership(context.Context, string, string, string) error {
	return r.transferErr
}

func (r testCreationRepository) RecordResult(context.Context, string, string, string, leagues.MatchResultInput) (leagues.League, error) {
	return r.result, r.resultErr
}

func (r testCreationRepository) Complete(context.Context, string, string) (leagues.League, error) {
	return r.completed, r.completeErr
}

func (testCreationRepository) GetPublic(context.Context, string) (leagues.League, error) {
	return leagues.League{}, nil
}

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

func (testRegistrationRepository) SearchUsernames(context.Context, string) ([]string, error) {
	return []string{}, nil
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

type testFederatedRepository struct {
	challengeErr        error
	authenticateErr     error
	addWithTicketErr    error
	removeWithTicketErr error
	reauthenticationErr error
}

type testGoogleVerifier struct{ identity federated.Identity }

func (v testGoogleVerifier) Verify(context.Context, string) (federated.Identity, error) {
	return v.identity, nil
}

func (r testFederatedRepository) CreateChallenge(context.Context, []byte, time.Time) (string, error) {
	return "019abcde-1111-7111-8111-111111111111", r.challengeErr
}
func (r testFederatedRepository) AuthenticateGoogle(context.Context, string, []byte, federated.Identity, *federated.Registration, []byte, []byte) (federated.Session, error) {
	return federated.Session{}, r.authenticateErr
}
func (testFederatedRepository) AddGoogleIdentity(context.Context, string, string, []byte, federated.Identity) error {
	return nil
}
func (r testFederatedRepository) ReauthenticateGoogle(context.Context, string, string, string, []byte, federated.Identity, []byte) error {
	return r.reauthenticationErr
}

func TestCreateReauthenticationTicketReportsSelectedGoogleAccountConflict(t *testing.T) {
	service := federated.NewService(
		testFederatedRepository{reauthenticationErr: federated.ErrIdentityConflict},
		testGoogleVerifier{identity: federated.Identity{
			Email:         "person@example.test",
			EmailVerified: true,
			Issuer:        federated.GoogleIssuer,
			Nonce:         "nonce",
			Subject:       "other-google-account",
		}},
	)
	request := httptest.NewRequest(http.MethodPost, "/v1/me/reauthentication-tickets", strings.NewReader(`{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"google-id-token","purpose":"set-local-password"}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request = request.WithContext(context.WithValue(request.Context(), accountContextKey{}, "019abcde-2222-7222-8222-222222222222"))
	recorder := httptest.NewRecorder()

	createReauthenticationTicket(access.Service{}, &service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}
func (r testFederatedRepository) AddGoogleIdentityWithTicket(context.Context, string, string, []byte, federated.Identity, []byte) error {
	return r.addWithTicketErr
}
func (r testFederatedRepository) RemoveGoogleIdentityWithTicket(context.Context, string, []byte) error {
	return r.removeWithTicketErr
}

func TestCreateGoogleChallengeReturnsCreated(t *testing.T) {
	service := federated.NewService(testFederatedRepository{}, nil)
	recorder := httptest.NewRecorder()

	createGoogleChallenge(service).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/google-login-challenges", nil))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestFederatedHandlersRecordOnlySafeRootFailureReasons(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	validIdentity := federated.Identity{Issuer: federated.GoogleIssuer, Subject: "google-subject", Email: "person@example.test", Nonce: "nonce", EmailVerified: true}
	const challengeID = "019abcde-1111-7111-8111-111111111111"
	for _, test := range []struct {
		name    string
		route   string
		handler http.Handler
		body    string
		reason  string
	}{
		{
			name: "challenge database failure", route: "POST /v1/google-login-challenges", body: "",
			handler: createGoogleChallenge(federated.NewService(testFederatedRepository{challengeErr: errors.New("database refused secret-nonce")}, nil)), reason: "database.query_failed",
		},
		{
			name: "google unavailable", route: "POST /v1/google-sessions", body: "",
			handler: http.HandlerFunc(unavailableFederatedLogin), reason: "identity.google_unavailable",
		},
		{
			name: "session validation", route: "POST /v1/google-sessions", body: `{"challengeId":"not-a-uuid","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{}, testGoogleVerifier{identity: validIdentity}), sessionCookies(true)), reason: "validation.rejected",
		},
		{
			name: "session email conflict", route: "POST /v1/google-sessions", body: `{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{authenticateErr: federated.ErrEmailConflict}, testGoogleVerifier{identity: validIdentity}), sessionCookies(true)), reason: "identity.email_conflict",
		},
		{
			name: "session invalid google proof", route: "POST /v1/google-sessions", body: `{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{}, testGoogleVerifier{}), sessionCookies(true)), reason: "credential.google_challenge_invalid",
		},
		{
			name: "link identity conflict", route: "POST /v1/me/google-identities", body: `{"ticket":"secret-ticket","challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token"}`,
			handler: createGoogleIdentity(federated.NewService(testFederatedRepository{addWithTicketErr: federated.ErrIdentityConflict}, testGoogleVerifier{identity: validIdentity})), reason: "identity.google_conflict",
		},
		{
			name: "link invalid reauthentication", route: "POST /v1/me/google-identities", body: `{"ticket":"secret-ticket","challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token"}`,
			handler: createGoogleIdentity(federated.NewService(testFederatedRepository{addWithTicketErr: federated.ErrChallengeInvalid}, testGoogleVerifier{identity: validIdentity})), reason: "credential.reauthentication_invalid",
		},
		{
			name: "unlink invalid reauthentication", route: "DELETE /v1/me/google-identities", body: `{"ticket":"secret-ticket"}`,
			handler: deleteGoogleIdentity(federated.NewService(testFederatedRepository{removeWithTicketErr: federated.ErrChallengeInvalid}, nil)), reason: "credential.reauthentication_invalid",
		},
		{
			name: "unlink database failure", route: "DELETE /v1/me/google-identities", body: `{"ticket":"secret-ticket"}`,
			handler: deleteGoogleIdentity(federated.NewService(testFederatedRepository{removeWithTicketErr: errors.New("database rejected session-token")}, nil)), reason: "database.query_failed",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer secret-session-token")
			ctx, span := provider.Tracer("test").Start(request.Context(), test.route)
			test.handler.ServeHTTP(httptest.NewRecorder(), request.WithContext(ctx))
			span.End()

			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); got != test.reason {
				t.Fatalf("failure reason = %q, want %q", got, test.reason)
			}
			if got := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); strings.Contains(got, "secret") || strings.Contains(got, "person@example.test") {
				t.Fatalf("failure reason leaked request data: %q", got)
			}
			exporter.Reset()
		})
	}
}

func testSpanAttribute(attributes []attribute.KeyValue, key string) string {
	for _, item := range attributes {
		if string(item.Key) == key {
			return item.Value.AsString()
		}
	}
	return ""
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

func TestCancelLeagueAllowsBearerSession(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	creation := leagues.NewCreationService(testCreationRepository{cancelled: leagues.League{ID: leagueID, Name: "Liga", State: "cancelled", Teams: []leagues.Team{}, Matches: []leagues.Match{}}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/cancel", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"state":"cancelled"`) {
		t.Errorf("body = %s, want cancelled league", recorder.Body.String())
	}
}

func TestCancelLeagueMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: leagues.ErrLeagueForbidden, status: http.StatusForbidden},
		"wrong state":   {err: leagues.ErrLeagueCancellationConflict, status: http.StatusConflict},
		"not found":     {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{cancelErr: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/cancel", nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestCompleteLeagueMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer":   {err: leagues.ErrLeagueForbidden, status: http.StatusForbidden},
		"pending matches": {err: leagues.ErrLeagueCompletionConflict, status: http.StatusConflict},
		"not found":       {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{completeErr: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/complete", nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestAddLeagueTeamReturnsTheCreatedTeam(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	team := leagues.Team{ID: "019abcde-3333-7333-8333-333333333333", Name: "Azules", Position: 3}
	creation := leagues.NewCreationService(testCreationRepository{team: team})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/teams", strings.NewReader(`{"name":"Azules"}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"Azules"`) || !strings.Contains(recorder.Body.String(), `"id":"019abcde-3333-7333-8333-333333333333"`) {
		t.Errorf("body = %s, want created team response", recorder.Body.String())
	}
}

func TestAddLeagueTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: leagues.ErrLeagueForbidden, status: http.StatusForbidden},
		"wrong state":   {err: leagues.ErrLeagueTeamConflict, status: http.StatusConflict},
		"not found":     {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{teamErr: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/teams", strings.NewReader(`{"name":"Azules"}`))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestRemoveLeagueTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const teamID = "019abcde-3333-7333-8333-333333333333"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: leagues.ErrLeagueForbidden, status: http.StatusForbidden},
		"minimum teams": {err: leagues.ErrLeagueTeamConflict, status: http.StatusConflict},
		"not found":     {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{removeErr: test.err}))
			request := httptest.NewRequest(http.MethodDelete, "/v1/leagues/"+leagueID+"/teams/"+teamID, nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestWithdrawLeagueTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const teamID = "019abcde-3333-7333-8333-333333333333"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: leagues.ErrLeagueForbidden, status: http.StatusForbidden},
		"wrong state":   {err: leagues.ErrLeagueWithdrawalConflict, status: http.StatusConflict},
		"not found":     {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{withdrawErr: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/v1/leagues/"+leagueID+"/teams/"+teamID+"/withdraw", nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestRecordMatchResultUsesTheContractRoundField(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const matchID = "019abcde-3333-7333-8333-333333333333"
	creation := leagues.NewCreationService(testCreationRepository{result: leagues.League{ID: leagueID, State: "in_progress", Teams: []leagues.Team{}, Matches: []leagues.Match{{ID: matchID, RoundNumber: 1, State: "completed"}}}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequest(http.MethodPut, "/v1/leagues/"+leagueID+"/matches/"+matchID+"/result", strings.NewReader(`{"homeScore":2,"awayScore":1}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"round":1`) || strings.Contains(recorder.Body.String(), `"roundNumber"`) {
		t.Errorf("body = %s, want the OpenAPI field round", recorder.Body.String())
	}
}

func TestRecordMatchResultMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const matchID = "019abcde-3333-7333-8333-333333333333"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not administrator": {err: leagues.ErrMatchResultForbidden, status: http.StatusForbidden},
		"wrong state":       {err: leagues.ErrMatchResultConflict, status: http.StatusConflict},
		"not found":         {err: leagues.ErrLeagueNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins, leagues.NewCreationService(testCreationRepository{resultErr: test.err}))
			request := httptest.NewRequest(http.MethodPut, "/v1/leagues/"+leagueID+"/matches/"+matchID+"/result", strings.NewReader(`{"homeScore":2,"awayScore":1}`))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
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
	request.Header.Set("Origin", "http://localhost:8082")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	recorder := httptest.NewRecorder()

	testHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8082" {
		t.Errorf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", got)
	}
}

func TestScheduleAccountDeletionAllowsConfiguredCookieOrigin(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.Service{}, nil, testDeletionAuthenticator{testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/v1/me/account", nil)
	request.Header.Set("Origin", "http://localhost:8082")
	request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "opaque-session"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"deletionEffectiveAt":"2026-09-07T12:00:00Z"`) {
		t.Errorf("body = %s, want deletion effective date", recorder.Body.String())
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

func TestClientIPUsesForwardedAddressOnlyFromTrustedProxy(t *testing.T) {
	t.Parallel()

	trusted := newClientIPResolver([]netip.Prefix{netip.MustParsePrefix("192.168.65.0/24")})
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

func TestValidRegistrationDraftEnforcesLeagueNameCharacterLimit(t *testing.T) {
	t.Parallel()

	draft := &registration.Draft{
		Name:  strings.Repeat("a", leagues.MaximumLeagueNameLength),
		Teams: []string{"Azules", "Rojos"},
	}
	if !validRegistrationDraft(draft) {
		t.Fatalf("validRegistrationDraft() rejected %d characters", leagues.MaximumLeagueNameLength)
	}
	draft.Name += "a"
	if validRegistrationDraft(draft) {
		t.Errorf("validRegistrationDraft() accepted %d characters", leagues.MaximumLeagueNameLength+1)
	}
}

func TestRegistrationRateLimitsByClientIP(t *testing.T) {
	t.Parallel()

	handler := NewHandler(registration.NewService(testRegistrationRepository{}, nil), nil, testAuthenticator{}, leagues.NewService(testLeagueRepository{}), testAllowedOrigins)
	for range registrationLimit {
		request := httptest.NewRequest(http.MethodPost, "/v1/registrations", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"es"}`))
		request.RemoteAddr = "203.0.113.1:10000"
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusAccepted {
			t.Fatalf("status = %d before limit, want %d", recorder.Code, http.StatusAccepted)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/registrations", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","username":"person_name","locale":"es"}`))
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
