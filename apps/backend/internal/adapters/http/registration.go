package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

type registerRequest struct {
	Email        string       `json:"email"`
	Locale       string       `json:"locale"`
	Password     string       `json:"password"`
	Username     string       `json:"username"`
	TermsVersion string       `json:"termsVersion"`
	Draft        *leagueInput `json:"draft"`
}

type loginRequest struct {
	Email            string       `json:"email"`
	Password         string       `json:"password"`
	SessionTransport string       `json:"sessionTransport"`
	Draft            *leagueInput `json:"draft"`
}

// createLocalSession authenticates without disclosing whether email, password, or state failed.
func createLocalSession(service registration.Service, limiter *loginLimiter, cookies sessionCookieSettings, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body loginRequest
		if err := decodeBody(request, &body); err != nil || !validEmail(strings.TrimSpace(body.Email)) || len(body.Password) < 8 || len(body.Password) > 1024 || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}
		draft := toRegistrationDraft(body.Draft)
		if !validRegistrationDraft(draft) {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}
		if allowed, retryAfter := limiter.allow(resolveClientIP(request), strings.ToLower(strings.TrimSpace(body.Email))); !allowed {
			observability.RecordEndpointFailure(request.Context(), "rate_limit.exceeded")
			writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(writer, http.StatusTooManyRequests, "Too many sign-in attempts")
			return
		}
		result, err := service.Login(request.Context(), body.Email, body.Password, draft)
		if errors.Is(err, registration.ErrLoginInvalid) {
			observability.RecordEndpointFailure(request.Context(), "authentication.credentials_rejected")
			writeProblem(writer, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		if err != nil {
			recordAuthenticationTechnicalFailure(request.Context(), err)
			writeProblem(writer, http.StatusInternalServerError, "Could not sign in")
			return
		}
		if result.Pending {
			writer.WriteHeader(http.StatusAccepted)
			return
		}
		response := map[string]any{"user": sessionUserResponse(result.Session.AccountID, result.Session.Username, result.Session.LastTeamName), "delivery": body.SessionTransport, "expiresAt": result.Session.IdleExpiresAt, "refreshExpiresAt": result.Session.RefreshExpiresAt}
		if body.SessionTransport == "cookie" {
			cookies.setSession(writer, result.Access, result.Refresh, result.Session)
		} else {
			response["sessionToken"], response["refreshToken"] = result.Access, result.Refresh
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}
}

func toRegistrationDraft(draft *leagueInput) *registration.Draft {
	if draft == nil {
		return nil
	}
	teams := make([]string, len(draft.Teams))
	for index, team := range draft.Teams {
		teams[index] = team.Name
	}
	normalized := registration.NormalizeInput(registration.Input{Draft: &registration.Draft{ID: draft.DraftID, Name: draft.Name, Sport: draft.Sport, Teams: teams}})
	return normalized.Draft
}

func recordAuthenticationTechnicalFailure(ctx context.Context, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		observability.RecordEndpointFailure(ctx, "request.cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		observability.RecordEndpointFailure(ctx, "request.timeout")
	default:
		observability.RecordEndpointFailure(ctx, "request.failed")
	}
}

func register(service registration.Service, limiter *requestLimiter, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body registerRequest
		if err := decodeBody(request, &body); err != nil {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}

		input := registration.NormalizeInput(registration.Input{
			Email:        body.Email,
			Locale:       registration.Locale(body.Locale),
			Password:     body.Password,
			TermsVersion: body.TermsVersion,
			Username:     body.Username,
		})
		if body.Draft != nil {
			teams := make([]string, len(body.Draft.Teams))
			for index, team := range body.Draft.Teams {
				teams[index] = team.Name
			}
			input.Draft = &registration.Draft{ID: body.Draft.DraftID, Name: body.Draft.Name, Sport: body.Draft.Sport, Teams: teams}
		}
		input = registration.NormalizeInput(input)
		if !validRegistration(input) || input.TermsVersion != legal.CurrentTermsVersion || !validRegistrationDraft(input.Draft) {
			observability.RecordEndpointFailure(request.Context(), "validation.rejected")
			writeValidationProblem(writer)
			return
		}
		if allowed, retryAfter := limiter.allow(resolveClientIP(request)); !allowed {
			observability.RecordEndpointFailure(request.Context(), "rate_limit.exceeded")
			writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(writer, http.StatusTooManyRequests, "Too many registrations")
			return
		}
		if err := service.Register(request.Context(), input); err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not complete registration")
			return
		}
		writer.WriteHeader(http.StatusAccepted)
	}
}

func validRegistrationDraft(draft *registration.Draft) bool {
	if draft == nil {
		return true
	}
	if !uuidPattern.MatchString(draft.ID) || (draft.Sport != tournaments.SportFootball && draft.Sport != tournaments.SportBasketball) || len(strings.TrimSpace(draft.Name)) == 0 || utf8.RuneCountInString(draft.Name) > tournaments.MaximumTournamentNameLength || len(draft.Teams) < 1 || len(draft.Teams) > 64 {
		return false
	}
	seen := map[string]bool{}
	for _, team := range draft.Teams {
		name := strings.TrimSpace(team)
		if name == "" || utf8.RuneCountInString(name) > 100 || seen[strings.ToLower(name)] {
			return false
		}
		seen[strings.ToLower(name)] = true
	}
	return true
}

func validRegistration(input registration.Input) bool {
	if len(input.Email) == 0 || len(input.Email) > 254 || len(input.Password) < 8 || len(input.Password) > 1024 || !registration.IsSupportedLocale(input.Locale) || !usernamePattern.MatchString(input.Username) {
		return false
	}
	return validEmail(input.Email)
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(value, "@")
}

func writeValidationProblem(writer http.ResponseWriter) {
	writeProblem(writer, http.StatusBadRequest, "Invalid request")
}

func writeProblem(writer http.ResponseWriter, status int, title string) {
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"type":   "about:blank",
		"title":  title,
		"status": status,
	})
}
