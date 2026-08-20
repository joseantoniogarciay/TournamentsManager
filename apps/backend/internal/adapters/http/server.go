// Package http exposes the API HTTP adapters.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const (
	usernameAvailabilityLimit  = 30
	usernameAvailabilityWindow = time.Minute
	userSearchLimit            = 60
	registrationLimit          = 5
)

// NewHandler builds the infrastructure routes available before business endpoints.
func NewHandler(registrationService registration.Service, federatedService *federated.Service, authenticator sessionAuthenticator, leagueService leagues.Service, corsAllowedOrigins []string, creationServices ...leagues.CreationService) http.Handler {
	return NewHandlerWithCookieSecurity(registrationService, federatedService, authenticator, leagueService, corsAllowedOrigins, true, creationServices...)
}

// NewHandlerWithCookieSecurity configures Secure cookies except on local HTTP loopback.
func NewHandlerWithCookieSecurity(registrationService registration.Service, federatedService *federated.Service, authenticator sessionAuthenticator, leagueService leagues.Service, corsAllowedOrigins []string, cookieSecure bool, creationServices ...leagues.CreationService) http.Handler {
	return NewHandlerWithCookieSecurityAndTrustedProxies(registrationService, federatedService, authenticator, leagueService, corsAllowedOrigins, cookieSecure, nil, creationServices...)
}

// NewHandlerWithCookieSecurityAndTrustedProxies accepts the IP forwarded by Caddy
// only when the immediate connection originates from a configured proxy network.
func NewHandlerWithCookieSecurityAndTrustedProxies(registrationService registration.Service, federatedService *federated.Service, authenticator sessionAuthenticator, leagueService leagues.Service, corsAllowedOrigins []string, cookieSecure bool, trustedProxyCIDRs []netip.Prefix, creationServices ...leagues.CreationService) http.Handler {
	mux := http.NewServeMux()
	resolveClientIP := newClientIPResolver(trustedProxyCIDRs)
	cookies := sessionCookies(cookieSecure)
	var accessService access.Service
	if repository, ok := authenticator.(access.Repository); ok {
		accessService = access.NewService(repository)
	}
	availabilityLimiter := newRequestLimiter(usernameAvailabilityLimit, usernameAvailabilityWindow)
	userSearchLimiter := newRequestLimiter(userSearchLimit, usernameAvailabilityWindow)
	registrationLimiter := newRequestLimiter(registrationLimit, time.Minute)
	localLoginLimiter := newLoginLimiter(10, time.Minute)
	cookieCSRF := func(next http.Handler) http.Handler {
		return requireCookieCSRF(corsAllowedOrigins, next)
	}
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /v1/usernames/{username}/availability", usernameAvailability(registrationService, availabilityLimiter, resolveClientIP))
	mux.HandleFunc("GET /v1/users", searchUsers(registrationService, userSearchLimiter, resolveClientIP))
	mux.HandleFunc("POST /v1/registrations", register(registrationService, registrationLimiter, resolveClientIP))
	mux.HandleFunc("POST /v1/sessions", createLocalSession(registrationService, localLoginLimiter, cookies, resolveClientIP))
	if federatedService != nil {
		mux.HandleFunc("POST /v1/google-login-challenges", createGoogleChallenge(*federatedService))
		mux.HandleFunc("POST /v1/google-sessions", createGoogleSession(*federatedService, cookies))
		mux.Handle("POST /v1/me/google-identities", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createGoogleIdentity(*federatedService)))))
		mux.Handle("DELETE /v1/me/google-identities", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteGoogleIdentity(*federatedService)))))
	} else {
		mux.HandleFunc("POST /v1/google-login-challenges", unavailableFederatedLogin)
		mux.HandleFunc("POST /v1/google-sessions", unavailableFederatedLogin)
		mux.HandleFunc("POST /v1/me/google-identities", unavailableFederatedLogin)
		mux.HandleFunc("DELETE /v1/me/google-identities", unavailableFederatedLogin)
	}
	passwordResetLimiter := newRequestLimiter(10, time.Minute)
	mux.HandleFunc("POST /v1/password-resets", requestPasswordReset(registrationService, passwordResetLimiter, resolveClientIP))
	mux.HandleFunc("POST /v1/password-reset-links", inspectPasswordReset(registrationService))
	mux.HandleFunc("POST /v1/password-reset-confirmations", confirmPasswordReset(registrationService, cookies))
	mux.HandleFunc("POST /v1/registration-verifications", verifyRegistration(registrationService, cookies))
	mux.Handle("GET /v1/sessions", requireSession(authenticator)(http.HandlerFunc(getCurrentSession(authenticator))))
	mux.Handle("POST /v1/sessions/refresh", refreshCookieCSRF(corsAllowedOrigins, cookies, refreshSession(registrationService, cookies)))
	mux.Handle("DELETE /v1/sessions", cookieCSRF(http.HandlerFunc(revokeCurrentSession(authenticator, cookies))))
	mux.Handle("GET /v1/me/access-methods", requireSession(authenticator)(http.HandlerFunc(getAccessMethods(authenticator))))
	mux.Handle("POST /v1/me/reauthentication-tickets", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createReauthenticationTicket(accessService, federatedService)))))
	mux.Handle("PUT /v1/me/local-credential", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(putLocalCredential(accessService)))))
	mux.Handle("DELETE /v1/me/local-credential", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteLocalCredential(accessService)))))
	mux.Handle("DELETE /v1/me/account", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(scheduleAccountDeletion(authenticator, cookies)))))
	mux.Handle("GET /v1/me/leagues", requireSession(authenticator)(http.HandlerFunc(listAccountLeagues(leagueService))))
	mux.Handle("GET /v1/me/recent-leagues", requireSession(authenticator)(http.HandlerFunc(listRecentAccountLeagues(leagueService))))
	if repository, ok := authenticator.(notifications.Repository); ok {
		notificationService := notifications.NewService(repository)
		mux.Handle("GET /v1/me/notifications", requireSession(authenticator)(http.HandlerFunc(listNotifications(notificationService))))
		mux.Handle("GET /v1/me/notifications/unread-count", requireSession(authenticator)(http.HandlerFunc(unreadNotificationCount(notificationService))))
		mux.Handle("POST /v1/me/notifications/read", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(markAllNotificationsRead(notificationService)))))
		mux.Handle("DELETE /v1/me/notifications", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteAllNotifications(notificationService)))))
		mux.Handle("DELETE /v1/me/notifications/{notificationId}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteNotification(notificationService)))))
	}
	followHandler := requireSession(authenticator)(cookieCSRF(http.HandlerFunc(followLeague(leagueService))))
	mux.Handle("PUT /v1/me/leagues/{leagueId}/follow", followHandler)
	mux.Handle("DELETE /v1/me/leagues/{leagueId}/follow", followHandler)
	if len(creationServices) > 0 {
		creationService := creationServices[0]
		mux.Handle("POST /v1/leagues", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createLeague(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/teams", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(addLeagueTeam(creationService)))))
		mux.Handle("DELETE /v1/leagues/{leagueId}/teams/{teamId}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(removeLeagueTeam(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/teams/{teamId}/withdraw", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(withdrawLeagueTeam(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/start", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(startLeague(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/cancel", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(cancelLeague(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/complete", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(completeLeague(creationService)))))
		mux.Handle("PUT /v1/leagues/{leagueId}/matches/{matchId}/result", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(recordMatchResult(creationService)))))
		mux.Handle("PUT /v1/leagues/{leagueId}/administrators/{username}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(assignLeagueAdministrator(creationService)))))
		mux.Handle("GET /v1/leagues/{leagueId}/administrators", requireSession(authenticator)(http.HandlerFunc(listLeagueAdministrators(creationService))))
		mux.Handle("DELETE /v1/leagues/{leagueId}/administrators/{username}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(removeLeagueAdministrator(creationService)))))
		mux.Handle("POST /v1/leagues/{leagueId}/transfer", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(transferLeagueOwnership(creationService)))))
		mux.HandleFunc("GET /v1/leagues/{leagueId}", getPublicLeague(creationService))
	}
	withCookieName := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routedRequest := r.WithContext(context.WithValue(r.Context(), sessionCookieNameContextKey{}, cookies.name))
		mux.ServeHTTP(w, routedRequest)
		// ServeMux writes the matched template on the request it serves. Preserve
		// it for the outer observability middleware after adding request context.
		r.Pattern = routedRequest.Pattern
	})
	return requireAllowedOrigin(corsAllowedOrigins, withCookieName)
}

type accountDeletionScheduler interface {
	ScheduleAccountDeletion(context.Context, string) (time.Time, error)
}

func scheduleAccountDeletion(authenticator sessionAuthenticator, cookies sessionCookieSettings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		scheduler, ok := authenticator.(accountDeletionScheduler)
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not schedule account deletion")
			return
		}
		effectiveAt, err := scheduler.ScheduleAccountDeletion(r.Context(), accountID)
		if errors.Is(err, postgres.ErrAccountHasOwnedLeagues) {
			writeProblem(w, http.StatusConflict, "Account cannot be deleted while it owns leagues")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not schedule account deletion")
			return
		}
		if transport, _ := currentSessionTransport(r.Context()); transport == cookieSession {
			cookies.clear(w)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"deletionEffectiveAt": effectiveAt})
	}
}

type sessionCookieSettings struct {
	name        string
	refreshName string
	secure      bool
}

func sessionCookies(secure bool) sessionCookieSettings {
	if !secure {
		return sessionCookieSettings{name: "tm_session", refreshName: "tm_refresh"}
	}
	return sessionCookieSettings{name: "__Host-tm_session", refreshName: "__Host-tm_refresh", secure: true}
}

func (cookies sessionCookieSettings) set(w http.ResponseWriter, name, value, expiresAt string) {
	expires, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return
	}
	maxAge := int(time.Until(expires).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	// #nosec G124 -- cookieSecure is false only for a loopback HTTP PUBLIC_BASE_URL.
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", Secure: cookies.secure, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: maxAge, Expires: expires})
}

func (cookies sessionCookieSettings) setSession(w http.ResponseWriter, access, refresh string, session registration.Session) {
	cookies.set(w, cookies.name, access, session.IdleExpiresAt)
	cookies.set(w, cookies.refreshName, refresh, session.RefreshExpiresAt)
}

func (cookies sessionCookieSettings) clear(w http.ResponseWriter) {
	for _, name := range []string{cookies.name, cookies.refreshName} {
		// #nosec G124 -- cookieSecure is false only for a loopback HTTP PUBLIC_BASE_URL.
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", Secure: cookies.secure, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	}
}

func getCurrentSession(authenticator sessionAuthenticator) http.HandlerFunc {
	type currentSessionReader interface {
		GetCurrentSession(context.Context, string) (leagues.CurrentSession, error)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		reader, ok := authenticator.(currentSessionReader)
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve session")
			return
		}
		token, ok := currentSessionToken(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		session, err := reader.GetCurrentSession(r.Context(), token)
		if errors.Is(err, leagues.ErrUnauthenticated) {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve session")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user":              map[string]string{"id": session.AccountID, "username": session.Username},
			"idleExpiresAt":     session.IdleExpiresAt,
			"absoluteExpiresAt": session.AbsoluteExpiresAt,
		})
	}
}

type leagueInput struct {
	Name  string `json:"name"`
	Teams []struct {
		Name string `json:"name"`
	} `json:"teams"`
}
type startLeagueInput struct {
	RoundRobinLegs int `json:"roundRobinLegs"`
}
type teamInput struct {
	Name string `json:"name"`
}
type matchResultInput struct {
	HomeScore *int `json:"homeScore"`
	AwayScore *int `json:"awayScore"`
}

func createLeague(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var body leagueInput
		if err := decodeBody(r, &body); err != nil {
			writeValidationProblem(w)
			return
		}
		teams := make([]leagues.TeamInput, len(body.Teams))
		for i, team := range body.Teams {
			teams[i] = leagues.TeamInput{Name: team.Name}
		}
		league, err := service.Create(r.Context(), accountID, leagues.CreateInput{Name: body.Name, Teams: teams})
		if errors.Is(err, leagues.ErrInvalidLeagueInput) {
			writeValidationProblem(w)
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not create league")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(league)
	}
}
func addLeagueTeam(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("leagueId")
		var body teamInput
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || decodeBody(r, &body) != nil {
			writeValidationProblem(w)
			return
		}
		team, err := service.AddTeam(r.Context(), accountID, leagueID, leagues.TeamInput{Name: body.Name})
		if errors.Is(err, leagues.ErrInvalidLeagueInput) {
			writeValidationProblem(w)
			return
		}
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot modify this league's teams")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueTeamConflict) {
			writeProblem(w, http.StatusConflict, "League cannot accept this team")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not add team")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(team)
	}
}
func removeLeagueTeam(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, teamID := r.PathValue("leagueId"), r.PathValue("teamId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(teamID) {
			writeValidationProblem(w)
			return
		}
		err := service.RemoveTeam(r.Context(), accountID, leagueID, teamID)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot modify this league's teams")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League or team is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueTeamConflict) {
			writeProblem(w, http.StatusConflict, "An unstarted league must keep at least two teams")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not remove team")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func withdrawLeagueTeam(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, teamID := r.PathValue("leagueId"), r.PathValue("teamId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(teamID) {
			writeValidationProblem(w)
			return
		}
		league, err := service.WithdrawTeam(r.Context(), accountID, leagueID, teamID)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot withdraw teams from this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League or team is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueWithdrawalConflict) {
			writeProblem(w, http.StatusConflict, "Team cannot be withdrawn from this league")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not withdraw team")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}
func getPublicLeague(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("leagueId")
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(w)
			return
		}
		league, err := service.GetPublic(r.Context(), leagueID)
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve league")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}
func startLeague(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		leagueID := r.PathValue("leagueId")
		var body startLeagueInput
		if !uuidPattern.MatchString(leagueID) || decodeBody(r, &body) != nil {
			writeValidationProblem(w)
			return
		}
		league, err := service.Start(r.Context(), accountID, leagueID, leagues.StartInput{RoundRobinLegs: body.RoundRobinLegs})
		if errors.Is(err, leagues.ErrInvalidLeagueInput) {
			writeValidationProblem(w)
			return
		}
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot start this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueConflict) {
			writeProblem(w, http.StatusConflict, "League is no longer unstarted")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not start league")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func cancelLeague(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		leagueID := r.PathValue("leagueId")
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(w)
			return
		}
		league, err := service.Cancel(r.Context(), accountID, leagueID)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot cancel this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueCancellationConflict) {
			writeProblem(w, http.StatusConflict, "League cannot be cancelled from its current state")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not cancel league")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func completeLeague(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("leagueId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(w)
			return
		}
		league, err := service.Complete(r.Context(), accountID, leagueID)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot complete this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueCompletionConflict) {
			writeProblem(w, http.StatusConflict, "League cannot be completed yet")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not complete league")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func recordMatchResult(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, matchID := r.PathValue("leagueId"), r.PathValue("matchId")
		var body matchResultInput
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(matchID) || decodeBody(r, &body) != nil || body.HomeScore == nil || body.AwayScore == nil {
			writeValidationProblem(w)
			return
		}
		league, err := service.RecordResult(r.Context(), accountID, leagueID, matchID, leagues.MatchResultInput{HomeScore: *body.HomeScore, AwayScore: *body.AwayScore})
		if errors.Is(err, leagues.ErrInvalidLeagueInput) {
			writeValidationProblem(w)
			return
		}
		if errors.Is(err, leagues.ErrMatchResultForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot record results for this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League or match is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrMatchResultConflict) {
			writeProblem(w, http.StatusConflict, "League is not in progress")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not save result")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func assignLeagueAdministrator(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, username := r.PathValue("leagueId"), r.PathValue("username")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !usernamePattern.MatchString(username) {
			writeValidationProblem(w)
			return
		}
		err := service.AssignAdministrator(r.Context(), accountID, leagueID, username)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot assign administrators for this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League or account is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueAdministratorConflict) {
			writeProblem(w, http.StatusConflict, "League owner cannot be a delegated administrator")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not assign administrator")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listLeagueAdministrators(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("leagueId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(w)
			return
		}
		usernames, err := service.ListAdministrators(r.Context(), accountID, leagueID)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot view administrators for this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve administrators")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Usernames []string `json:"usernames"`
		}{Usernames: usernames})
	}
}

func removeLeagueAdministrator(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, username := r.PathValue("leagueId"), r.PathValue("username")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !usernamePattern.MatchString(username) {
			writeValidationProblem(w)
			return
		}
		err := service.RemoveAdministrator(r.Context(), accountID, leagueID, username)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot remove administrators for this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not remove administrator")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func transferLeagueOwnership(service leagues.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("leagueId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var body struct {
			Username string `json:"username"`
		}
		if !uuidPattern.MatchString(leagueID) || json.NewDecoder(r.Body).Decode(&body) != nil || !usernamePattern.MatchString(body.Username) {
			writeValidationProblem(w)
			return
		}
		err := service.TransferOwnership(r.Context(), accountID, leagueID, body.Username)
		if errors.Is(err, leagues.ErrLeagueForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot transfer this league")
			return
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(w, http.StatusNotFound, "League or account is unavailable")
			return
		}
		if errors.Is(err, leagues.ErrLeagueOwnershipTransferConflict) {
			writeProblem(w, http.StatusConflict, "League owner cannot receive the transfer")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not transfer league")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

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
			writeValidationProblem(w)
			return
		}
		var ticket, expiresAt string
		var err error
		if body.Password != "" {
			if len(body.Password) < 8 || len(body.Password) > 1024 {
				writeValidationProblem(w)
				return
			}
			if body.Purpose == access.RemoveLocalPassword {
				writeValidationProblem(w)
				return
			}
			ticket, expiresAt, err = service.ReauthenticateWithPassword(r.Context(), credential.token, body.Password, body.Purpose)
		} else if federatedService == nil || !uuidPattern.MatchString(body.ChallengeID) {
			writeValidationProblem(w)
			return
		} else {
			accountID, _ := currentAccountID(r.Context())
			if body.Purpose != access.SetLocalPassword && body.Purpose != access.RemoveLocalPassword {
				writeValidationProblem(w)
				return
			}
			ticket, expiresAt, err = federatedService.ReauthenticateGoogle(r.Context(), accountID, credential.token, body.ChallengeID, body.IDToken, string(body.Purpose))
		}
		if errors.Is(err, federated.ErrIdentityConflict) {
			writeProblem(w, http.StatusConflict, "Selected Google account is not linked to this account")
			return
		}
		if errors.Is(err, access.ErrReauthenticationInvalid) || errors.Is(err, federated.ErrChallengeInvalid) {
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
			writeValidationProblem(w)
			return
		}
		if err := service.AddGoogleWithTicket(r.Context(), credential.token, body.Ticket, body.ChallengeID, body.IDToken); errors.Is(err, federated.ErrIdentityConflict) {
			writeProblem(w, http.StatusConflict, "Could not link this access method")
			return
		} else if errors.Is(err, federated.ErrChallengeInvalid) {
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
			writeValidationProblem(w)
			return
		}
		if err := service.RemoveGoogleWithTicket(r.Context(), credential.token, body.Ticket); err != nil {
			writeProblem(w, http.StatusUnauthorized, "Invalid reauthentication")
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
			writeValidationProblem(w)
			return
		}
		if err := service.SetPassword(r.Context(), credential.token, body.Ticket, body.Password); errors.Is(err, access.ErrReauthenticationInvalid) {
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
			writeValidationProblem(w)
			return
		}
		if err := service.RemovePassword(r.Context(), credential.token, body.Ticket); err != nil {
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
			GetAccessMethods(context.Context, string) (leagues.AccessMethods, error)
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

func unavailableFederatedLogin(w http.ResponseWriter, _ *http.Request) {
	writeProblem(w, http.StatusServiceUnavailable, "Google sign-in is unavailable")
}

type googleSessionRequest struct {
	ChallengeID      string              `json:"challengeId"`
	IDToken          string              `json:"idToken"`
	SessionTransport string              `json:"sessionTransport"`
	Username         string              `json:"username"`
	Locale           string              `json:"locale"`
	Draft            *registration.Draft `json:"draft"`
}

func createGoogleChallenge(service federated.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		challenge, err := service.CreateChallenge(r.Context())
		if err != nil {
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
		if err := decodeBody(r, &body); err != nil || !uuidPattern.MatchString(body.ChallengeID) || body.IDToken == "" || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") || (body.Username != "" && !usernamePattern.MatchString(body.Username)) || (body.Locale != "" && !registration.IsSupportedLocale(registration.Locale(body.Locale))) || (body.Username == "") != (body.Locale == "") || (body.Draft != nil && (body.Username == "" || !validRegistrationDraft(body.Draft))) {
			writeValidationProblem(w)
			return
		}
		var registrationInput *federated.Registration
		if body.Username != "" {
			registrationInput = &federated.Registration{Username: body.Username, Locale: body.Locale, Draft: toFederatedDraft(body.Draft)}
		}
		established, err := service.Authenticate(r.Context(), body.ChallengeID, body.IDToken, registrationInput)
		if errors.Is(err, federated.ErrRegistration) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if errors.Is(err, federated.ErrEmailConflict) {
			writeProblem(w, http.StatusConflict, "Could not sign in with this access method")
			return
		}
		if errors.Is(err, federated.ErrChallengeInvalid) {
			writeValidationProblem(w)
			return
		}
		if err != nil {
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
	return &federated.Draft{Name: strings.TrimSpace(draft.Name), Teams: draft.Teams}
}

func writeFederatedSession(w http.ResponseWriter, transport string, established federated.EstablishedSession, cookies sessionCookieSettings) {
	response := map[string]any{"user": map[string]string{"id": established.AccountID, "username": established.Username}, "delivery": transport, "expiresAt": established.IdleExpiresAt, "refreshExpiresAt": established.RefreshExpiresAt}
	if transport == "cookie" {
		cookies.setSession(w, established.AccessToken, established.RefreshToken, registration.Session{AccountID: established.AccountID, Username: established.Username, IdleExpiresAt: established.IdleExpiresAt, RefreshExpiresAt: established.RefreshExpiresAt})
	} else {
		response["sessionToken"], response["refreshToken"] = established.AccessToken, established.RefreshToken
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

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
		response := map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
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
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		var token, transport string
		if authorization != "" {
			if !strings.HasPrefix(authorization, "Bearer ") || len(authorization) == len("Bearer ") {
				writeProblem(writer, http.StatusUnauthorized, "Invalid session")
				return
			}
			token, transport = strings.TrimPrefix(authorization, "Bearer "), "bearer"
		} else if cookieErr == nil && refreshCookie.Value != "" {
			token, transport = refreshCookie.Value, "cookie"
		} else {
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		session, accessToken, refreshToken, err := service.Refresh(request.Context(), token)
		if errors.Is(err, registration.ErrRefreshInvalid) {
			writeProblem(writer, http.StatusUnauthorized, "Invalid session")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not refresh session")
			return
		}
		response := map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": transport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
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

type clientIPResolver func(*http.Request) string

func newClientIPResolver(trustedProxyCIDRs []netip.Prefix) clientIPResolver {
	return func(request *http.Request) string {
		peer := remoteAddrIP(request.RemoteAddr)
		if peer.IsValid() {
			for _, trustedProxyCIDR := range trustedProxyCIDRs {
				if trustedProxyCIDR.Contains(peer) {
					if forwarded, err := netip.ParseAddr(strings.TrimSpace(request.Header.Get("X-Client-IP"))); err == nil {
						return forwarded.Unmap().String()
					}
					break
				}
			}
			return peer.String()
		}
		return request.RemoteAddr
	}
}

func remoteAddrIP(remoteAddr string) netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return address.Unmap()
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
		response := map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
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

func followLeague(service leagues.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		leagueID := request.PathValue("leagueId")
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(writer)
			return
		}
		var err error
		if request.Method == http.MethodPut {
			err = service.Follow(request.Context(), accountID, leagueID)
		} else {
			err = service.Unfollow(request.Context(), accountID, leagueID)
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(writer, http.StatusNotFound, "League is unavailable")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not update follow status")
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func listAccountLeagues(service leagues.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}

		relationship := leagues.Relationship(request.URL.Query().Get("relationship"))
		limit, valid := listLimit(request.URL.Query().Get("limit"))
		cursor := request.URL.Query().Get("cursor")
		if !valid || (cursor != "" && !uuidPattern.MatchString(cursor)) {
			writeValidationProblem(writer)
			return
		}
		page, err := service.List(request.Context(), accountID, relationship, cursor, limit)
		if errors.Is(err, leagues.ErrInvalidRelationship) || errors.Is(err, leagues.ErrInvalidPage) {
			writeValidationProblem(writer)
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not retrieve leagues")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(page)
	}
}

func listRecentAccountLeagues(service leagues.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		items, err := service.ListRecent(request.Context(), accountID)
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not retrieve recent leagues")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(items)
	}
}

func listLimit(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	limit, err := strconv.Atoi(raw)
	return limit, err == nil
}

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

type registerRequest struct {
	Email    string       `json:"email"`
	Locale   string       `json:"locale"`
	Password string       `json:"password"`
	Username string       `json:"username"`
	Draft    *leagueInput `json:"draft"`
}

type loginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	SessionTransport string `json:"sessionTransport"`
}

// createLocalSession authenticates without disclosing whether email, password, or state failed.
func createLocalSession(service registration.Service, limiter *loginLimiter, cookies sessionCookieSettings, resolveClientIP clientIPResolver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body loginRequest
		if err := decodeBody(request, &body); err != nil || !validEmail(strings.TrimSpace(body.Email)) || len(body.Password) < 8 || len(body.Password) > 1024 || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") {
			writeValidationProblem(writer)
			return
		}
		if allowed, retryAfter := limiter.allow(resolveClientIP(request), strings.ToLower(strings.TrimSpace(body.Email))); !allowed {
			writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(writer, http.StatusTooManyRequests, "Too many sign-in attempts")
			return
		}
		result, err := service.Login(request.Context(), body.Email, body.Password)
		if errors.Is(err, registration.ErrLoginInvalid) {
			writeProblem(writer, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not sign in")
			return
		}
		if result.Pending {
			writer.WriteHeader(http.StatusAccepted)
			return
		}
		response := map[string]any{"user": map[string]string{"id": result.Session.AccountID, "username": result.Session.Username}, "delivery": body.SessionTransport, "expiresAt": result.Session.IdleExpiresAt, "refreshExpiresAt": result.Session.RefreshExpiresAt}
		if body.SessionTransport == "cookie" {
			cookies.setSession(writer, result.Access, result.Refresh, result.Session)
		} else {
			response["sessionToken"], response["refreshToken"] = result.Access, result.Refresh
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
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
			Email:    body.Email,
			Locale:   registration.Locale(body.Locale),
			Password: body.Password,
			Username: body.Username,
		})
		if body.Draft != nil {
			teams := make([]string, len(body.Draft.Teams))
			for index, team := range body.Draft.Teams {
				teams[index] = team.Name
			}
			input.Draft = &registration.Draft{Name: body.Draft.Name, Teams: teams}
		}
		input = registration.NormalizeInput(input)
		if !validRegistration(input) || !validRegistrationDraft(input.Draft) {
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
	if len(strings.TrimSpace(draft.Name)) == 0 || utf8.RuneCountInString(draft.Name) > leagues.MaximumLeagueNameLength || len(draft.Teams) < 2 || len(draft.Teams) > 64 {
		return false
	}
	seen := map[string]bool{}
	for _, team := range draft.Teams {
		name := strings.TrimSpace(team)
		if name == "" || len(name) > 100 || seen[strings.ToLower(name)] {
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

type requestLimiter struct {
	entries map[string]requestLimit
	limit   int
	mu      sync.Mutex
	window  time.Duration
}

type loginLimiter struct {
	byEmail, byIP *requestLimiter
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{byEmail: newRequestLimiter(limit, window), byIP: newRequestLimiter(limit, window)}
}

func (l *loginLimiter) allow(ip, email string) (bool, int) {
	allowedIP, retryIP := l.byIP.allow(ip)
	allowedEmail, retryEmail := l.byEmail.allow(email)
	if allowedIP && allowedEmail {
		return true, 0
	}
	if retryIP > retryEmail {
		return false, retryIP
	}
	return false, retryEmail
}

type requestLimit struct {
	count   int
	resetAt time.Time
}

func newRequestLimiter(limit int, window time.Duration) *requestLimiter {
	return &requestLimiter{entries: make(map[string]requestLimit), limit: limit, window: window}
}

func (l *requestLimiter) allow(key string) (bool, int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	for entryKey, entry := range l.entries {
		if !entry.resetAt.After(now) {
			delete(l.entries, entryKey)
		}
	}

	entry := l.entries[key]
	if entry.resetAt.IsZero() {
		entry.resetAt = now.Add(l.window)
	}
	if entry.count >= l.limit {
		retryAfter := int(time.Until(entry.resetAt).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}
