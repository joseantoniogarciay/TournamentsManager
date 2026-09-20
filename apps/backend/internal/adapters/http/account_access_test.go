package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

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
