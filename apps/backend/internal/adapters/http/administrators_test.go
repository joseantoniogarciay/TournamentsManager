package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestListTournamentAdministratorsReturnsUsernames(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{administrators: []string{"alex", "bea"}}))
	request := httptest.NewRequest(http.MethodGet, "/v1/tournaments/"+leagueID+"/administrators", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if body := recorder.Body.String(); body != "{\"usernames\":[\"alex\",\"bea\"]}\n" {
		t.Errorf("body = %s, want usernames", body)
	}
}

func TestTransferTournamentOwnershipMapsBusinessErrors(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"forbidden": {tournaments.ErrTournamentForbidden, http.StatusForbidden},
		"conflict":  {tournaments.ErrTournamentOwnershipTransferConflict, http.StatusConflict},
		"not found": {tournaments.ErrTournamentNotFound, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{transferErr: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/v1/tournaments/"+leagueID+"/transfer", strings.NewReader(`{"username":"alex"}`))
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("X-CSRF-Token", "token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestAdministratorManagementMapsBusinessErrors(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	const leagueID = "019abcde-2222-7222-8222-222222222222"
	for name, test := range map[string]struct {
		method string
		err    error
		status int
	}{
		"list forbidden":   {method: http.MethodGet, err: tournaments.ErrTournamentForbidden, status: http.StatusForbidden},
		"remove not found": {method: http.MethodDelete, err: tournaments.ErrTournamentNotFound, status: http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{administratorsErr: test.err, removeErr: test.err}))
			path := "/v1/tournaments/" + leagueID + "/administrators"
			if test.method == http.MethodDelete {
				path += "/alex"
			}
			request := httptest.NewRequest(test.method, path, nil)
			request.Header.Set("Authorization", "Bearer session-token")
			request.Header.Set("X-CSRF-Token", "token")
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d; body = %s", recorder.Code, test.status, recorder.Body.String())
			}
		})
	}
}
