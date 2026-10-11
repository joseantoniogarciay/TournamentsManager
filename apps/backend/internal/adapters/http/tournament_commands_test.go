package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTournamentInvitationFailureReasonsAreClosedAndSafe(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		err    error
		reason string
	}{
		"unavailable": {err: tournaments.ErrTournamentInvitationNotFound, reason: "tournament.invitation_not_found"},
		"conflict":    {err: tournaments.ErrTournamentInvitationConflict, reason: "tournament.invitation_conflict"},
	} {
		t.Run(name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			ctx, span := provider.Tracer("test").Start(context.Background(), "invitation")
			recordTournamentFailure(ctx, test.err)
			span.End()
			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); got != test.reason {
				t.Fatalf("failure reason = %q, want %q", got, test.reason)
			}
		})
	}
}

func TestCreateLocalSessionReturnsBearerSession(t *testing.T) {
	t.Parallel()
	var receivedDraft *registration.Draft
	passwordHash, err := registration.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("crear hash: %v", err)
	}
	repository := testRegistrationRepository{
		loginAccount: registration.LocalAccount{ID: "019abcde-1111-7111-8111-111111111111", PasswordHash: passwordHash, Verified: true},
		loginDraft:   &receivedDraft,
		loginSession: registration.Session{AccountID: "019abcde-1111-7111-8111-111111111111", Username: "person", IdleExpiresAt: "2026-08-09T12:00:00Z", RefreshExpiresAt: "2026-09-01T12:00:00Z"},
	}
	handler := NewHandler(registration.NewService(repository, nil), nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/sessions", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","sessionTransport":"bearer","draft":{"draftId":"019abcde-1111-7111-8111-111111111112","name":" Copa ","sport":"badminton","bestOfSets":3,"pointsPerGame":15,"teams":[{"name":" Norte "},{"name":"Sur"}]}}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"delivery":"bearer"`) || !strings.Contains(recorder.Body.String(), `"sessionToken"`) || !strings.Contains(recorder.Body.String(), `"refreshToken"`) {
		t.Errorf("body = %s, want bearer session tokens", recorder.Body.String())
	}
	if receivedDraft == nil || receivedDraft.ID != "019abcde-1111-7111-8111-111111111112" || receivedDraft.Name != "Copa" || receivedDraft.Teams[0] != "Norte" || receivedDraft.PointsPerGame != 15 {
		t.Errorf("draft = %#v, want normalized tournament", receivedDraft)
	}
}

func TestCreateGoogleSessionPassesDraftForExistingIdentity(t *testing.T) {
	t.Parallel()
	var receivedDraft *federated.Draft
	repository := testFederatedRepository{authenticateDraft: &receivedDraft}
	service := federated.NewService(repository, testGoogleVerifier{identity: federated.Identity{
		Issuer: federated.GoogleIssuer, Subject: "subject", Email: "person@example.test", Nonce: "nonce", EmailVerified: true,
	}})
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/google-sessions", strings.NewReader(`{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"google-token","sessionTransport":"bearer","draft":{"draftId":"019abcde-1111-7111-8111-111111111112","name":"Copa Google","sport":"badminton","bestOfSets":3,"pointsPerGame":21,"teams":[{"name":"Uno"},{"name":"Dos"}]}}`))
	recorder := httptest.NewRecorder()

	createGoogleSession(service, sessionCookies(false)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if receivedDraft == nil || receivedDraft.ID != "019abcde-1111-7111-8111-111111111112" || receivedDraft.Name != "Copa Google" || receivedDraft.Sport != tournaments.SportBadminton {
		t.Fatalf("draft = %#v, want badminton tournament", receivedDraft)
	}
	if receivedDraft == nil || receivedDraft.PointsPerGame != 21 || receivedDraft.BestOfSets != 3 {
		t.Fatalf("Google lost badminton config: %+v", receivedDraft)
	}
}

func TestCreateTournamentAcceptsAndReturnsBasketball(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	created := tournaments.Tournament{
		ID: "019abcde-2222-7222-8222-222222222222", Name: "Liga de baloncesto",
		Sport: tournaments.SportBasketball, State: "published", Teams: []tournaments.Team{{ID: "a", Name: "Azules"}, {ID: "b", Name: "Rojos"}}, Matches: []tournaments.Match{},
	}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{created: created}))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments", strings.NewReader(`{"name":"Liga de baloncesto","sport":"basketball","teams":[{"name":"Azules"},{"name":"Rojos"}]}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s; want %d", recorder.Code, recorder.Body.String(), http.StatusCreated)
	}
	if !strings.Contains(recorder.Body.String(), `"sport":"basketball"`) {
		t.Fatalf("body = %s; want basketball sport", recorder.Body.String())
	}
}

func TestCreateTournamentAcceptsAndReturnsHandball(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	created := tournaments.Tournament{
		ID: "019abcde-2222-7222-8222-222222222222", Name: "Liga de balonmano",
		Sport: tournaments.SportHandball, State: "published", Teams: []tournaments.Team{{ID: "a", Name: "Azules"}, {ID: "b", Name: "Rojos"}}, Matches: []tournaments.Match{},
	}
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{created: created}))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments", strings.NewReader(`{"name":"Liga de balonmano","sport":"handball","teams":[{"name":"Azules"},{"name":"Rojos"}]}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s; want %d", recorder.Code, recorder.Body.String(), http.StatusCreated)
	}
	if !strings.Contains(recorder.Body.String(), `"sport":"handball"`) {
		t.Fatalf("body = %s; want handball sport", recorder.Body.String())
	}
}

func TestStartMixedTournamentMapsConfigurationConflict(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	creation := tournaments.NewCreationService(testCreationRepository{startErr: tournaments.ErrInvalidMixedConfiguration})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+tournamentID+"/start", strings.NewReader(`{"format":"league_then_single_elimination","roundRobinLegs":1,"leagueStructure":"groups","groupCount":4,"qualifiersPerGroup":2}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnprocessableEntity, recorder.Body.String())
	}
}

func TestStartTournamentEliminationMapsTransitionConflict(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	creation := tournaments.NewCreationService(testCreationRepository{eliminationErr: tournaments.ErrTournamentStageTransitionConflict})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+tournamentID+"/stages/elimination/start", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
}

func TestCancelTournamentAllowsBearerSession(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	creation := tournaments.NewCreationService(testCreationRepository{cancelled: tournaments.Tournament{ID: leagueID, Name: "Liga", State: "cancelled", Teams: []tournaments.Team{}, Matches: []tournaments.Match{}}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/cancel", nil)
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

func TestCancelTournamentMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"wrong state":   {err: tournaments.ErrTournamentCancellationConflict, status: http.StatusConflict},
		"not found":     {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{cancelErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/cancel", nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestCompleteTournamentMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer":   {err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"pending matches": {err: tournaments.ErrTournamentCompletionConflict, status: http.StatusConflict},
		"not found":       {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{completeErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/complete", nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestAddTournamentTeamReturnsTheCreatedTeam(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	team := tournaments.Team{ID: "019abcde-3333-7333-8333-333333333333", Name: "Azules", Position: 3}
	creation := tournaments.NewCreationService(testCreationRepository{team: team})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/teams", strings.NewReader(`{"name":"Azules"}`))
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

func TestAddTournamentTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"wrong state":   {err: tournaments.ErrTournamentTeamConflict, status: http.StatusConflict},
		"not found":     {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{teamErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/teams", strings.NewReader(`{"name":"Azules"}`))
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

func TestCreateTournamentTeamInvitationReturnsOneTimeSecret(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{}))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+tournamentID+"/team-invitation", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil || len(body.Token) != 43 {
		t.Fatalf("token = %q, decode error = %v; want 43-character secret", body.Token, err)
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", recorder.Header().Get("Cache-Control"))
	}
}

func TestInspectTournamentTeamInvitationIsPublicAndReturnsSafeProjection(t *testing.T) {
	t.Parallel()
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	const token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	creation := tournaments.NewCreationService(testCreationRepository{invitation: tournaments.TeamInvitation{TournamentID: tournamentID, TournamentName: "Copa abierta"}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/team-invitations/inspection", strings.NewReader(`{"token":"`+token+`"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"tournamentName":"Copa abierta"`) || strings.Contains(recorder.Body.String(), token) {
		t.Errorf("body = %s; want safe tournament projection without token", recorder.Body.String())
	}
}

func TestJoinTournamentTeamInvitationReturnsRegistration(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	const teamID = "019abcde-3333-7333-8333-333333333333"
	const token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	creation := tournaments.NewCreationService(testCreationRepository{registration: tournaments.TeamRegistration{
		TournamentID: tournamentID,
		Team:         tournaments.Team{ID: teamID, Name: "Mi equipo", Position: 2},
	}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/team-invitations/registration", strings.NewReader(`{"token":"`+token+`","name":" Mi equipo "}`))
	request.Header.Set("Authorization", "Bearer session-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"tournamentId":"`+tournamentID+`"`) || !strings.Contains(recorder.Body.String(), `"name":"Mi equipo"`) {
		t.Errorf("body = %s; want tournament and created team", recorder.Body.String())
	}
}

func TestTournamentTeamInvitationEndpointsMapBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const tournamentID = "019abcde-2222-7222-8222-222222222222"
	const token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name       string
		path       string
		repository testCreationRepository
		want       int
	}{
		{name: "create forbidden", path: "/v1/tournaments/" + tournamentID + "/team-invitation", repository: testCreationRepository{saveInviteErr: tournaments.ErrTournamentForbidden}, want: http.StatusForbidden},
		{name: "create after start", path: "/v1/tournaments/" + tournamentID + "/team-invitation", repository: testCreationRepository{saveInviteErr: tournaments.ErrTournamentInvitationConflict}, want: http.StatusConflict},
		{name: "join unavailable", path: "/v1/team-invitations/registration", repository: testCreationRepository{registrationErr: tournaments.ErrTournamentInvitationNotFound}, want: http.StatusNotFound},
		{name: "join conflict", path: "/v1/team-invitations/registration", repository: testCreationRepository{registrationErr: tournaments.ErrTournamentInvitationConflict}, want: http.StatusConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(test.repository))
			body := io.Reader(nil)
			if strings.HasSuffix(test.path, "/registration") {
				body = strings.NewReader(`{"token":"` + token + `","name":"Azules"}`)
			}
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, test.path, body)
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Errorf("status = %d, want %d; body = %s", recorder.Code, test.want, recorder.Body.String())
			}
		})
	}
}

func TestRemoveTournamentTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const teamID = "019abcde-3333-7333-8333-333333333333"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"minimum teams": {err: tournaments.ErrTournamentTeamConflict, status: http.StatusConflict},
		"not found":     {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{removeErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/v1/tournaments/"+leagueID+"/teams/"+teamID, nil)
			request.Header.Set("Authorization", "Bearer session-token")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestWithdrawTournamentTeamMapsBusinessErrors(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const teamID = "019abcde-3333-7333-8333-333333333333"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"not organizer": {err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"wrong state":   {err: tournaments.ErrTournamentWithdrawalConflict, status: http.StatusConflict},
		"not found":     {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{withdrawErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/tournaments/"+leagueID+"/teams/"+teamID+"/withdraw", nil)
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
	creation := tournaments.NewCreationService(testCreationRepository{result: tournaments.Tournament{ID: leagueID, State: "in_progress", Teams: []tournaments.Team{}, Matches: []tournaments.Match{{ID: matchID, RoundNumber: 1, State: "completed", ResultType: tournaments.ResultPlayed}}}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/v1/tournaments/"+leagueID+"/matches/"+matchID+"/result", strings.NewReader(`{"homeScore":2,"awayScore":1}`))
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

func TestRecordMatchResultEnforcesExclusiveResultShapes(t *testing.T) {
	t.Parallel()
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	const matchID = "019abcde-3333-7333-8333-333333333333"
	creation := tournaments.NewCreationService(testCreationRepository{result: tournaments.Tournament{ID: leagueID, State: "in_progress", Teams: []tournaments.Team{}, Matches: []tournaments.Match{}}})
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, creation)
	for name, test := range map[string]struct {
		body   string
		status int
	}{
		"sets":               {body: `{"sets":[{"homeScore":6,"awayScore":4},{"homeScore":7,"awayScore":5}]}`, status: http.StatusOK},
		"mixed":              {body: `{"homeScore":2,"awayScore":0,"sets":[{"homeScore":6,"awayScore":4},{"homeScore":7,"awayScore":5}]}`, status: http.StatusBadRequest},
		"zero games":         {body: `{"sets":[{"homeScore":6,"awayScore":0},{"homeScore":6,"awayScore":0}]}`, status: http.StatusOK},
		"missing away games": {body: `{"sets":[{"homeScore":6},{"homeScore":6}]}`, status: http.StatusBadRequest},
		"missing home games": {body: `{"sets":[{"awayScore":6},{"awayScore":6}]}`, status: http.StatusBadRequest},
		"null away games":    {body: `{"sets":[{"homeScore":6,"awayScore":null},{"homeScore":6,"awayScore":null}]}`, status: http.StatusBadRequest},
		"null home games":    {body: `{"sets":[{"homeScore":null,"awayScore":6},{"homeScore":null,"awayScore":6}]}`, status: http.StatusBadRequest},
		"null set":           {body: `{"sets":[null,{"homeScore":6,"awayScore":0}]}`, status: http.StatusBadRequest},
		"empty sets":         {body: `{"sets":[]}`, status: http.StatusBadRequest},
		"partial score":      {body: `{"homeScore":2}`, status: http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
			ctx, span := provider.Tracer("test").Start(context.Background(), "result")
			request := httptest.NewRequestWithContext(ctx, http.MethodPut, "/v1/tournaments/"+leagueID+"/matches/"+matchID+"/result", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)
			span.End()
			if test.status == http.StatusBadRequest {
				spans := exporter.GetSpans()
				if len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != "validation.rejected" {
					t.Fatal("invalid shape must record only the safe validation reason")
				}
			}

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d; body = %s", recorder.Code, test.status, recorder.Body.String())
			}
		})
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
		"not administrator":   {err: tournaments.ErrMatchResultForbidden, status: http.StatusForbidden},
		"wrong state":         {err: tournaments.ErrMatchResultConflict, status: http.StatusConflict},
		"basketball tie":      {err: tournaments.ErrInvalidTournamentInput, status: http.StatusBadRequest},
		"draw without winner": {err: tournaments.ErrInvalidBracketResult, status: http.StatusBadRequest},
		"dependent result":    {err: tournaments.ErrBracketResultDependency, status: http.StatusConflict},
		"unresolved slots":    {err: tournaments.ErrBracketMatchNotReady, status: http.StatusConflict},
		"technical failure":   {err: errors.New("private database error token=secret"), status: http.StatusInternalServerError},
		"not found":           {err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{resultErr: test.err}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/v1/tournaments/"+leagueID+"/matches/"+matchID+"/result", strings.NewReader(`{"homeScore":2,"awayScore":1}`))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
			if strings.Contains(recorder.Body.String(), "secret") || strings.Contains(recorder.Body.String(), "private database") {
				t.Fatal("internal error exposed")
			}
		})
	}
}
func (r testRegistrationRepository) RenewLoginVerification(context.Context, string, []byte) (string, registration.Locale, error) {
	return "", "", registration.ErrLoginInvalid
}

func testHandler() http.Handler {
	return NewHandler(registration.Service{}, nil, testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
}

func TestTableTennisDraftAndCreationShareSetFormat(t *testing.T) {
	for _, best := range []int{3, 5, 7} {
		draft := &registration.Draft{ID: "019abcde-1111-7111-8111-111111111111", Name: "Table tennis", Sport: tournaments.SportTableTennis, BestOfSets: best, Teams: []string{"A"}}
		if !validRegistrationDraft(draft) {
			t.Errorf("draft best of %d rejected", best)
		}
	}
	for _, best := range []int{0, 2, 6, 9} {
		draft := &registration.Draft{ID: "019abcde-1111-7111-8111-111111111111", Name: "Table tennis", Sport: tournaments.SportTableTennis, BestOfSets: best, Teams: []string{"A"}}
		if validRegistrationDraft(draft) {
			t.Errorf("draft best of %d accepted", best)
		}
	}
}

func TestVolleyballDraftRequiresBestOfFive(t *testing.T) {
	for _, best := range []int{0, 3, 5, 7} {
		draft := &registration.Draft{ID: "019abcde-1111-7111-8111-111111111111", Name: "Volleyball", Sport: tournaments.SportVolleyball, BestOfSets: best, Teams: []string{"A"}}
		if got := validRegistrationDraft(draft); got != (best == 5) {
			t.Errorf("best=%d valid=%v", best, got)
		}
	}
}

func TestBadmintonDraftAndHTTPConfiguration(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	for _, points := range []int{0, 11, 15, 21, 30} {
		draft := &registration.Draft{ID: accountID, Name: "Badminton", Sport: tournaments.SportBadminton, BestOfSets: 3, PointsPerGame: points, Teams: []string{"A"}}
		valid := points == 15 || points == 21
		if validRegistrationDraft(draft) != valid {
			t.Errorf("draft points=%d", points)
		}
		created := tournaments.Tournament{ID: accountID, Name: "Badminton", Sport: tournaments.SportBadminton, BestOfSets: 3, PointsPerGame: points, State: "published", Teams: []tournaments.Team{}, Matches: []tournaments.Match{}}
		var received tournaments.CreateInput
		repository := captureBadmintonCreation{testCreationRepository: testCreationRepository{created: created}, input: &received}
		handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(repository))
		payload, _ := json.Marshal(map[string]any{"name": "Badminton", "sport": "badminton", "bestOfSets": 3, "pointsPerGame": points, "teams": []map[string]string{{"name": "A"}}})
		exporter := tracetest.NewInMemoryExporter()
		provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		ctx, span := provider.Tracer("test").Start(context.Background(), "POST /v1/tournaments")
		request := httptest.NewRequestWithContext(ctx, http.MethodPost, "/v1/tournaments", strings.NewReader(string(payload)))
		request.Header.Set("Authorization", "Bearer session-token")
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		span.End()
		if valid {
			if recorder.Code != http.StatusCreated || received.PointsPerGame != points {
				t.Fatalf("points=%d status=%d received=%+v", points, recorder.Code, received)
			}
			var response tournaments.Tournament
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.PointsPerGame != points {
				t.Fatalf("response=%+v %v", response, err)
			}
		} else {
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("points=%d status=%d", points, recorder.Code)
			}
			spans := exporter.GetSpans()
			if len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != "validation.rejected" {
				t.Fatal("unsafe or missing validation diagnostic")
			}
		}
	}
}

type captureBadmintonCreation struct {
	testCreationRepository
	input *tournaments.CreateInput
}

func (r captureBadmintonCreation) Create(_ context.Context, _ string, input tournaments.CreateInput) (tournaments.Tournament, error) {
	*r.input = input
	return r.created, nil
}

func TestMatchIncidentHTTPShapesAndSafeFeedback(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	cases := []struct {
		name, payload string
		sport         tournaments.Sport
		status        int
	}{
		{"no show", `{"incident":{"type":"no_show","side":"away"}}`, tournaments.SportFootball, 200},
		{"retirement score", `{"incident":{"type":"retirement","side":"home","partialHomeScore":52,"partialAwayScore":48}}`, tournaments.SportBasketball, 200},
		{"partial set", `{"incident":{"type":"retirement","side":"home","partialSets":[{"homeScore":6,"awayScore":4},{"homeScore":2,"awayScore":1}]}}`, tournaments.SportPadel, 200},
		{"both absent", `{"incident":{"type":"no_show","side":"both"}}`, tournaments.SportFootball, 400},
		{"unknown reason", `{"incident":{"type":"injury","side":"home"}}`, tournaments.SportFootball, 400},
		{"missing side", `{"incident":{"type":"no_show"}}`, tournaments.SportFootball, 400},
		{"mixed result", `{"incident":{"type":"no_show","side":"home"},"homeScore":0,"awayScore":0}`, tournaments.SportFootball, 400},
		{"mixed sets", `{"incident":{"type":"no_show","side":"home"},"sets":[]}`, tournaments.SportFootball, 400},
		{"no show partial", `{"incident":{"type":"no_show","side":"home","partialHomeScore":0,"partialAwayScore":0}}`, tournaments.SportFootball, 400},
		{"one partial", `{"incident":{"type":"retirement","side":"home","partialHomeScore":0}}`, tournaments.SportFootball, 400},
		{"negative partial", `{"incident":{"type":"retirement","side":"home","partialHomeScore":-1,"partialAwayScore":0}}`, tournaments.SportFootball, 400},
		{"empty sets", `{"incident":{"type":"retirement","side":"home","partialSets":[]}}`, tournaments.SportPadel, 400},
		{"null set", `{"incident":{"type":"retirement","side":"home","partialSets":[{"homeScore":null,"awayScore":0}]}}`, tournaments.SportPadel, 400},
		{"mixed partial types", `{"incident":{"type":"retirement","side":"home","partialHomeScore":1,"partialAwayScore":0,"partialSets":[{"homeScore":1,"awayScore":0}]}}`, tournaments.SportPadel, 400},
		{"invalid cap", `{"incident":{"type":"retirement","side":"home","partialSets":[{"homeScore":31,"awayScore":29}]}}`, tournaments.SportBadminton, 400},
		{"medical free text", `{"incident":{"type":"retirement","side":"home","detail":"private reason"}}`, tournaments.SportFootball, 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var received tournaments.MatchResultInput
			repository := incidentHTTPRepository{testCreationRepository: testCreationRepository{}, sport: c.sport, input: &received}
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(repository))
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			ctx, span := provider.Tracer("test").Start(context.Background(), "PUT /v1/tournaments/{tournamentId}/matches/{matchId}/result")
			request := httptest.NewRequestWithContext(ctx, http.MethodPut, "/v1/tournaments/"+accountID+"/matches/"+accountID+"/result", strings.NewReader(c.payload))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			span.End()
			if recorder.Code != c.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if c.status == 200 && received.Incident == nil {
				t.Fatal("lost incident")
			}
			if c.status == 400 {
				spans := exporter.GetSpans()
				if len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != "validation.rejected" {
					t.Fatal("unsafe validation diagnostic")
				}
				if strings.Contains(recorder.Body.String(), "private reason") {
					t.Fatal("leaked raw input")
				}
			}
		})
	}
}

type incidentHTTPRepository struct {
	testCreationRepository
	sport tournaments.Sport
	input *tournaments.MatchResultInput
}

func (r incidentHTTPRepository) RecordResult(_ context.Context, _ string, _ string, _ string, input tournaments.MatchResultInput) (tournaments.Tournament, error) {
	*r.input = input
	best, points := 0, 0
	if tournaments.SetSport(r.sport) {
		best = 3
	}
	if r.sport == tournaments.SportBadminton {
		points = 21
	}
	_, err := tournaments.NormalizeIncidentResult(r.sport, best, points, input)
	return tournaments.Tournament{ID: "019abcde-1111-7111-8111-111111111111", Sport: r.sport, State: "in_progress"}, err
}
