package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func searchUsers(service registration.Service, limiter *requestLimiter, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if !usernamePattern.MatchString(query) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if allowed, retryAfter := limiter.allow(resolveClientIP(r)); !allowed {
			observability.RecordEndpointFailure(r.Context(), "rate_limit.exceeded")
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(w, http.StatusTooManyRequests, "Too many searches")
			return
		}
		usernames, err := service.SearchUsernames(r.Context(), query)
		if err != nil {
			writeProblem(w, http.StatusServiceUnavailable, "Could not search users")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"usernames": usernames})
	}
}

type reauthenticationRequest struct {
	Password, ChallengeID, IDToken string
	Purpose                        access.Purpose
}
type localCredentialRequest struct {
	Ticket   string `json:"ticket"`
	Password string `json:"password"`
}

func createReauthenticationTicket(service access.Service, federatedService *federated.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body reauthenticationRequest
		credential, ok := sessionToken(r)
		if !ok || decodeBody(r, &body) != nil || !access.IsPurpose(body.Purpose) || (body.Password == "" && (body.ChallengeID == "" || body.IDToken == "")) || (body.Password != "" && (body.ChallengeID != "" || body.IDToken != "")) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		var ticket, expiresAt string
		var err error
		if body.Password != "" {
			if len(body.Password) < 8 || len(body.Password) > 1024 {
				observability.RecordEndpointFailure(r.Context(), "validation.rejected")
				writeValidationProblem(w)
				return
			}
			if body.Purpose == access.RemoveLocalPassword {
				observability.RecordEndpointFailure(r.Context(), "validation.rejected")
				writeValidationProblem(w)
				return
			}
			ticket, expiresAt, err = service.ReauthenticateWithPassword(r.Context(), credential.token, body.Password, body.Purpose)
		} else if federatedService == nil || !uuidPattern.MatchString(body.ChallengeID) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		} else {
			accountID, _ := currentAccountID(r.Context())
			if body.Purpose != access.SetLocalPassword && body.Purpose != access.RemoveLocalPassword {
				observability.RecordEndpointFailure(r.Context(), "validation.rejected")
				writeValidationProblem(w)
				return
			}
			ticket, expiresAt, err = federatedService.ReauthenticateGoogle(r.Context(), accountID, credential.token, body.ChallengeID, body.IDToken, string(body.Purpose))
		}
		if errors.Is(err, federated.ErrIdentityConflict) {
			observability.RecordEndpointFailure(r.Context(), "reauthentication.identity_conflict")
			writeProblem(w, http.StatusConflict, "Selected Google account is not linked to this account")
			return
		}
		if errors.Is(err, access.ErrReauthenticationInvalid) || errors.Is(err, federated.ErrChallengeInvalid) {
			observability.RecordEndpointFailure(r.Context(), "reauthentication.invalid")
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not reauthenticate")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"ticket": ticket, "expiresAt": expiresAt})
	}
}

type googleIdentityLinkRequest struct{ Ticket, ChallengeID, IDToken string }

func createGoogleIdentity(service federated.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body googleIdentityLinkRequest
		credential, ok := sessionToken(r)
		if !ok || decodeBody(r, &body) != nil || body.Ticket == "" || body.IDToken == "" || !uuidPattern.MatchString(body.ChallengeID) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if err := service.AddGoogleWithTicket(r.Context(), credential.token, body.Ticket, body.ChallengeID, body.IDToken); errors.Is(err, federated.ErrIdentityConflict) {
			observability.RecordEndpointFailure(r.Context(), "identity.google_conflict")
			writeProblem(w, http.StatusConflict, "Could not link this access method")
			return
		} else if errors.Is(err, federated.ErrChallengeInvalid) {
			observability.RecordEndpointFailure(r.Context(), "credential.reauthentication_invalid")
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
			return
		} else if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not link Google")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type reauthenticationTicketRequest struct{ Ticket string }

func deleteGoogleIdentity(service federated.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body reauthenticationTicketRequest
		credential, ok := sessionToken(r)
		if !ok || decodeBody(r, &body) != nil || body.Ticket == "" {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if err := service.RemoveGoogleWithTicket(r.Context(), credential.token, body.Ticket); errors.Is(err, federated.ErrChallengeInvalid) {
			observability.RecordEndpointFailure(r.Context(), "credential.reauthentication_invalid")
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
			return
		} else if err != nil {
			observability.RecordDatabaseEndpointFailure(r.Context(), err)
			writeProblem(w, http.StatusInternalServerError, "Could not unlink Google")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func putLocalCredential(service access.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body localCredentialRequest
		credential, ok := sessionToken(r)
		if !ok || decodeBody(r, &body) != nil || body.Ticket == "" || len(body.Password) < 8 || len(body.Password) > 1024 {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if err := service.SetPassword(r.Context(), credential.token, body.Ticket, body.Password); errors.Is(err, access.ErrReauthenticationInvalid) {
			observability.RecordEndpointFailure(r.Context(), "reauthentication.invalid")
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
			return
		} else if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not change password")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteLocalCredential(service access.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body reauthenticationTicketRequest
		credential, ok := sessionToken(r)
		if !ok || decodeBody(r, &body) != nil || body.Ticket == "" {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if err := service.RemovePassword(r.Context(), credential.token, body.Ticket); errors.Is(err, access.ErrLastAccessMethod) {
			observability.RecordEndpointFailure(r.Context(), "access_method.last_remaining")
			writeProblem(w, http.StatusConflict, "Cannot remove the last access method")
			return
		} else if err != nil {
			observability.RecordEndpointFailure(r.Context(), "reauthentication.invalid")
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getAccessMethods(authenticator sessionAuthenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		reader, ok := authenticator.(interface {
			GetAccessMethods(context.Context, string) (tournaments.AccessMethods, error)
		})
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve account")
			return
		}
		access, err := reader.GetAccessMethods(r.Context(), accountID)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve account")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"email": access.Email, "username": access.Username, "methods": map[string]bool{"password": access.HasPassword, "google": access.HasGoogle}})
	}
}

type sessionRevoker interface {
	RevokeSession(context.Context, string) error
}

func revokeCurrentSession(authenticator sessionAuthenticator, cookies sessionCookieSettings) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		credential, ok := sessionToken(request)
		if !ok {
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		revoker, ok := authenticator.(sessionRevoker)
		if !ok {
			writeProblem(writer, http.StatusInternalServerError, "Could not revoke session")
			return
		}
		if err := revoker.RevokeSession(request.Context(), credential.token); err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not revoke session")
			return
		}
		if credential.transport == cookieSession {
			cookies.clear(writer)
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func unavailableFederatedLogin(w http.ResponseWriter, r *http.Request) {
	observability.RecordEndpointFailure(r.Context(), "identity.google_unavailable")
	writeProblem(w, http.StatusServiceUnavailable, "Google sign-in is unavailable")
}
