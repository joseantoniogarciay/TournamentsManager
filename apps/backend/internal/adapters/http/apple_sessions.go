package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

func socialRateLimit(next http.HandlerFunc, limiter *requestLimiter, resolve clientIPResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowed, retry := limiter.allow(resolve(r)); !allowed {
			observability.RecordEndpointFailure(r.Context(), "rate_limit.exceeded")
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeProblem(w, http.StatusTooManyRequests, "Too many sign-in attempts")
			return
		}
		next(w, r)
	}
}

func unavailableAppleLogin(w http.ResponseWriter, r *http.Request) {
	observability.RecordEndpointFailure(r.Context(), "identity.apple_unavailable")
	writeProblem(w, http.StatusServiceUnavailable, "Sign-in is not available")
}

func createAppleChallenge(service federated.AppleService, config HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Platform       string `json:"platform"`
			ProofChallenge string `json:"proofChallenge"`
		}
		if decodeBody(r, &body) != nil || (body.Platform != "web" && body.Platform != "ios" && body.Platform != "android") || len(body.ProofChallenge) != 64 {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		challenge, err := service.CreateChallenge(r.Context(), body.Platform, body.ProofChallenge)
		if err != nil {
			writeAppleFailure(w, r, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": challenge.ID, "authorizationUrl": challenge.AuthorizationURL, "returnUrl": appleReturnURL(config, body.Platform), "expiresAt": challenge.ExpiresAt})
	}
}

func appleReturnURL(config HandlerConfig, platform string) string {
	if platform == "web" {
		return config.AppleWebReturnURL
	}
	return config.AppleNativeReturnURL
}

func receiveAppleCallback(service federated.AppleService, config HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") || r.ParseForm() != nil || len(r.PostForm["state"]) != 1 || len(r.PostForm.Get("state")) != 43 || len(r.PostForm.Get("code")) > 4096 || len(r.PostForm["code"]) > 1 || len(r.PostForm["error"]) > 1 {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		providerError := r.PostForm.Get("error")
		denied := providerError == "access_denied"
		callback, err := service.Complete(r.Context(), r.PostForm.Get("state"), r.PostForm.Get("code"), providerError != "")
		if callback.ID == "" {
			writeAppleFailure(w, r, err)
			return
		}
		status := "ready"
		if denied {
			status = "cancelled"
		} else if err != nil || providerError != "" {
			status = "failed"
			recordAppleFailure(r.Context(), err)
		}
		if r.Context().Err() != nil {
			return
		}
		location := appleReturnURL(config, callback.Platform) + "?" + url.Values{"challengeId": {callback.ID}, "status": {status}}.Encode()
		http.Redirect(w, r, location, http.StatusSeeOther)
	}
}

type appleSessionRequest struct {
	ChallengeID      string       `json:"challengeId"`
	Proof            string       `json:"proof"`
	SessionTransport string       `json:"sessionTransport"`
	Username         string       `json:"username"`
	Locale           string       `json:"locale"`
	TermsVersion     string       `json:"termsVersion"`
	Draft            *leagueInput `json:"draft"`
}

func createAppleSession(service federated.AppleService, cookies sessionCookieSettings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		var body appleSessionRequest
		if decodeBody(r, &body) != nil {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		draft := toRegistrationDraft(body.Draft)
		if !uuidPattern.MatchString(body.ChallengeID) || len(body.Proof) != 43 || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") || (body.Username != "" && !usernamePattern.MatchString(body.Username)) || (body.Locale != "" && !registration.IsSupportedLocale(registration.Locale(body.Locale))) || (body.Username == "") != (body.Locale == "") || (body.Username != "" && body.TermsVersion != legal.CurrentTermsVersion) || !validRegistrationDraft(draft) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		var input *federated.Registration
		if body.Username != "" {
			input = &federated.Registration{Username: body.Username, Locale: body.Locale, TermsVersion: body.TermsVersion}
		}
		established, err := service.Authenticate(r.Context(), body.ChallengeID, body.Proof, input, toFederatedDraft(draft))
		if errors.Is(err, federated.ErrRegistration) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if err != nil {
			writeAppleFailure(w, r, err)
			return
		}
		writeFederatedSession(w, body.SessionTransport, established, cookies)
	}
}

func recordAppleFailure(ctx context.Context, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		return
	case errors.Is(err, federated.ErrAppleUnavailable):
		observability.RecordEndpointFailure(ctx, "identity.provider_unavailable")
	case errors.Is(err, federated.ErrChallengeInvalid):
		observability.RecordEndpointFailure(ctx, "credential.apple_challenge_invalid")
	case errors.Is(err, federated.ErrEmailConflict):
		observability.RecordEndpointFailure(ctx, "identity.email_conflict")
	case err == nil:
		observability.RecordEndpointFailure(ctx, "credential.apple_challenge_invalid")
	default:
		observability.RecordDatabaseEndpointFailure(ctx, err)
	}
}

func writeAppleFailure(w http.ResponseWriter, r *http.Request, err error) {
	recordAppleFailure(r.Context(), err)
	if errors.Is(err, context.Canceled) || r.Context().Err() != nil {
		return
	}
	switch {
	case errors.Is(err, federated.ErrChallengeInvalid):
		writeValidationProblem(w)
	case errors.Is(err, federated.ErrEmailConflict):
		writeProblem(w, http.StatusConflict, "Could not sign in with this access method")
	case errors.Is(err, federated.ErrAppleUnavailable):
		writeProblem(w, http.StatusServiceUnavailable, "Sign-in is not available")
	default:
		writeProblem(w, http.StatusInternalServerError, "Could not sign in")
	}
}
