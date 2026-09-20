package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type googleSessionRequest struct {
	ChallengeID      string       `json:"challengeId"`
	IDToken          string       `json:"idToken"`
	SessionTransport string       `json:"sessionTransport"`
	Username         string       `json:"username"`
	Locale           string       `json:"locale"`
	TermsVersion     string       `json:"termsVersion"`
	Draft            *leagueInput `json:"draft"`
}

func createGoogleChallenge(service federated.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		challenge, err := service.CreateChallenge(r.Context())
		if err != nil {
			observability.RecordDatabaseEndpointFailure(r.Context(), err)
			writeProblem(w, http.StatusInternalServerError, "Could not start Google sign-in")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": challenge.ID, "nonce": challenge.Nonce, "expiresAt": challenge.ExpiresAt})
	}
}

func createGoogleSession(service federated.Service, cookies sessionCookieSettings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body googleSessionRequest
		if err := decodeBody(r, &body); err != nil {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		draft := toRegistrationDraft(body.Draft)
		if !uuidPattern.MatchString(body.ChallengeID) || body.IDToken == "" || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") || (body.Username != "" && !usernamePattern.MatchString(body.Username)) || (body.Locale != "" && !registration.IsSupportedLocale(registration.Locale(body.Locale))) || (body.Username == "") != (body.Locale == "") || (body.Username != "" && body.TermsVersion != legal.CurrentTermsVersion) || !validRegistrationDraft(draft) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		var registrationInput *federated.Registration
		if body.Username != "" {
			registrationInput = &federated.Registration{Username: body.Username, Locale: body.Locale, TermsVersion: body.TermsVersion}
		}
		established, err := service.Authenticate(r.Context(), body.ChallengeID, body.IDToken, registrationInput, toFederatedDraft(draft))
		if errors.Is(err, federated.ErrRegistration) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if errors.Is(err, federated.ErrEmailConflict) {
			observability.RecordEndpointFailure(r.Context(), "identity.email_conflict")
			writeProblem(w, http.StatusConflict, "Could not sign in with this access method")
			return
		}
		if errors.Is(err, federated.ErrChallengeInvalid) {
			observability.RecordEndpointFailure(r.Context(), "credential.google_challenge_invalid")
			writeValidationProblem(w)
			return
		}
		if err != nil {
			recordAuthenticationTechnicalFailure(r.Context(), err)
			writeProblem(w, http.StatusInternalServerError, "Could not sign in")
			return
		}
		writeFederatedSession(w, body.SessionTransport, established, cookies)
	}
}

func toFederatedDraft(draft *registration.Draft) *federated.Draft {
	if draft == nil {
		return nil
	}
	return &federated.Draft{ID: draft.ID, Name: strings.TrimSpace(draft.Name), Sport: draft.Sport, Teams: draft.Teams}
}

func writeFederatedSession(w http.ResponseWriter, transport string, established federated.EstablishedSession, cookies sessionCookieSettings) {
	response := map[string]any{"user": sessionUserResponse(established.AccountID, established.Username, established.LastTeamName), "delivery": transport, "expiresAt": established.IdleExpiresAt, "refreshExpiresAt": established.RefreshExpiresAt}
	if transport == "cookie" {
		cookies.setSession(w, established.AccessToken, established.RefreshToken, registration.Session{AccountID: established.AccountID, Username: established.Username, LastTeamName: established.LastTeamName, IdleExpiresAt: established.IdleExpiresAt, RefreshExpiresAt: established.RefreshExpiresAt})
	} else {
		response["sessionToken"], response["refreshToken"] = established.AccessToken, established.RefreshToken
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
