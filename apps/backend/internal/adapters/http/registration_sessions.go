package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type passwordResetRequest struct {
	Email string `json:"email"`
}
type passwordResetTokenRequest struct {
	Token string `json:"token"`
}
type passwordResetConfirmationRequest struct {
	Token            string `json:"token"`
	Password         string `json:"password"`
	SessionTransport string `json:"sessionTransport"`
}

func requestPasswordReset(service registration.Service, limiter *requestLimiter, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body passwordResetRequest
		if err := decodeBody(r, &body); err != nil || !validEmail(strings.TrimSpace(body.Email)) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		if allowed, retry := limiter.allow(resolveClientIP(r)); !allowed {
			observability.RecordEndpointFailure(r.Context(), "rate_limit.exceeded")
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeProblem(w, http.StatusTooManyRequests, "Too many requests")
			return
		}
		if err := service.RequestPasswordReset(r.Context(), body.Email); err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not request password reset")
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
func inspectPasswordReset(service registration.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body passwordResetTokenRequest
		if err := decodeBody(r, &body); err != nil || body.Token == "" {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		email, err := service.InspectPasswordReset(r.Context(), body.Token)
		if errors.Is(err, registration.ErrPasswordResetInvalid) {
			observability.RecordEndpointFailure(r.Context(), "credential.reset_link_invalid")
			writeProblem(w, http.StatusConflict, "Invalid link")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not inspect link")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"email": email})
	}
}
func confirmPasswordReset(service registration.Service, cookies sessionCookieSettings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body passwordResetConfirmationRequest
		if err := decodeBody(r, &body); err != nil || body.Token == "" || len(body.Password) < 8 || len(body.Password) > 1024 || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		session, access, refresh, err := service.ResetPassword(r.Context(), body.Token, body.Password)
		if errors.Is(err, registration.ErrPasswordResetInvalid) {
			observability.RecordEndpointFailure(r.Context(), "credential.reset_link_invalid")
			writeProblem(w, http.StatusConflict, "Invalid link")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not change password")
			return
		}
		response := map[string]any{"user": sessionUserResponse(session.AccountID, session.Username, session.LastTeamName), "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
		if body.SessionTransport == "cookie" {
			cookies.setSession(w, access, refresh, session)
		} else {
			response["sessionToken"], response["refreshToken"] = access, refresh
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func decodeBody(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return err
	}
	return nil
}

func refreshSession(service registration.Service, cookies sessionCookieSettings) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		refreshCookie, cookieErr := request.Cookie(cookies.refreshName)
		if authorization != "" && cookieErr == nil {
			observability.RecordEndpointFailure(request.Context(), "session.refresh_invalid")
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		var token, transport string
		if authorization != "" {
			if !strings.HasPrefix(authorization, "Bearer ") || len(authorization) == len("Bearer ") {
				observability.RecordEndpointFailure(request.Context(), "session.refresh_invalid")
				writeProblem(writer, http.StatusUnauthorized, "Invalid session")
				return
			}
			token, transport = strings.TrimPrefix(authorization, "Bearer "), "bearer"
		} else if cookieErr == nil && refreshCookie.Value != "" {
			token, transport = refreshCookie.Value, "cookie"
		} else {
			observability.RecordEndpointFailure(request.Context(), "session.refresh_invalid")
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		session, accessToken, refreshToken, err := service.Refresh(request.Context(), token)
		if errors.Is(err, registration.ErrRefreshInvalid) {
			observability.RecordEndpointFailure(request.Context(), "session.refresh_invalid")
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		if err != nil {
			observability.RecordDatabaseEndpointFailure(request.Context(), err)
			writeProblem(writer, http.StatusInternalServerError, "Could not refresh session")
			return
		}
		response := map[string]any{"user": sessionUserResponse(session.AccountID, session.Username, session.LastTeamName), "delivery": transport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
		if transport == "cookie" {
			cookies.setSession(writer, accessToken, refreshToken, session)
		} else {
			response["sessionToken"], response["refreshToken"] = accessToken, refreshToken
		}
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}
}

func usernameAvailability(service registration.Service, limiter *requestLimiter, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		if !usernamePattern.MatchString(username) {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}
		if allowed, retryAfter := limiter.allow(resolveClientIP(request)); !allowed {
			observability.RecordEndpointFailure(request.Context(), "rate_limit.exceeded")
			writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(writer, http.StatusTooManyRequests, "Too many username lookups")
			return
		}

		available, err := service.UsernameAvailable(request.Context(), username)
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not retrieve username")
			return
		}
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]bool{"available": available})
	}
}

type verificationRequest struct {
	Token            string `json:"token"`
	SessionTransport string `json:"sessionTransport"`
}

func verifyRegistration(service registration.Service, cookies sessionCookieSettings) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body verificationRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || decoder.Decode(&struct{}{}) != io.EOF || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") || body.Token == "" {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}
		previousSession, _ := sessionToken(request)
		session, sessionToken, refreshToken, err := service.Verify(request.Context(), body.Token, previousSession.token)
		if errors.Is(err, registration.ErrVerificationInvalid) {
			observability.RecordEndpointFailure(request.Context(), "credential.verification_link_invalid")
			writeProblem(writer, http.StatusConflict, "Invalid verification")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not verify account")
			return
		}
		response := map[string]any{"user": sessionUserResponse(session.AccountID, session.Username, session.LastTeamName), "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
		if body.SessionTransport == "cookie" {
			cookies.setSession(writer, sessionToken, refreshToken, session)
		} else {
			response["sessionToken"] = sessionToken
			response["refreshToken"] = refreshToken
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}
}
