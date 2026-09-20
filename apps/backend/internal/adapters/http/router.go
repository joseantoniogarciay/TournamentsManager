package http

import (
	"context"
	"net/http"
	"net/netip"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// HandlerConfig contains transport concerns that do not belong to a use case.
type HandlerConfig struct {
	CORSAllowedOrigins []string
	CookieSecure       bool
	TrustedProxyCIDRs  []netip.Prefix
	EdgeProxyAuthToken string
}

// HandlerDependencies contains the application services exposed over HTTP.
type HandlerDependencies struct {
	Registration       registration.Service
	Federated          *federated.Service
	Authenticator      sessionAuthenticator
	TournamentList     tournaments.Service
	TournamentCreation *tournaments.CreationService
	Suggestions        *suggestions.Service
	RISCReceiver       http.Handler
}

// NewHandler builds a handler with secure-cookie defaults. It is kept as a
// compact convenience for focused handler tests.
func NewHandler(registrationService registration.Service, federatedService *federated.Service, authenticator sessionAuthenticator, tournamentService tournaments.Service, corsAllowedOrigins []string, creationServices ...tournaments.CreationService) http.Handler {
	dependencies := HandlerDependencies{
		Registration:   registrationService,
		Federated:      federatedService,
		Authenticator:  authenticator,
		TournamentList: tournamentService,
	}
	if len(creationServices) > 0 {
		dependencies.TournamentCreation = &creationServices[0]
	}
	return NewHandlerWithConfig(HandlerConfig{CORSAllowedOrigins: corsAllowedOrigins, CookieSecure: true}, dependencies)
}

// NewHandlerWithConfig builds the complete HTTP adapter from explicit
// configuration and dependencies.
func NewHandlerWithConfig(config HandlerConfig, dependencies HandlerDependencies) http.Handler {
	registrationService := dependencies.Registration
	federatedService := dependencies.Federated
	authenticator := dependencies.Authenticator
	tournamentService := dependencies.TournamentList
	creationService := dependencies.TournamentCreation
	suggestionService := dependencies.Suggestions
	mux := http.NewServeMux()
	resolveClientIP := newClientIPResolver(config.TrustedProxyCIDRs, config.EdgeProxyAuthToken)
	cookies := sessionCookies(config.CookieSecure)
	var accessService access.Service
	if repository, ok := authenticator.(access.Repository); ok {
		accessService = access.NewService(repository)
	}
	availabilityLimiter := newRequestLimiter(usernameAvailabilityLimit, usernameAvailabilityWindow)
	userSearchLimiter := newRequestLimiter(userSearchLimit, usernameAvailabilityWindow)
	registrationLimiter := newRequestLimiter(registrationLimit, time.Minute)
	localLoginLimiter := newLoginLimiter(10, time.Minute)
	cookieCSRF := func(next http.Handler) http.Handler {
		return requireCookieCSRF(config.CORSAllowedOrigins, next)
	}
	mux.HandleFunc("GET /healthz", healthz)
	if dependencies.RISCReceiver != nil {
		mux.Handle("POST /v1/risc/events", dependencies.RISCReceiver)
	}
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
	mux.Handle("POST /v1/sessions/refresh", refreshCookieCSRF(config.CORSAllowedOrigins, cookies, refreshSession(registrationService, cookies)))
	mux.Handle("DELETE /v1/sessions", cookieCSRF(http.HandlerFunc(revokeCurrentSession(authenticator, cookies))))
	mux.Handle("GET /v1/me/access-methods", requireSession(authenticator)(http.HandlerFunc(getAccessMethods(authenticator))))
	mux.Handle("POST /v1/me/reauthentication-tickets", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createReauthenticationTicket(accessService, federatedService)))))
	mux.Handle("PUT /v1/me/local-credential", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(putLocalCredential(accessService)))))
	mux.Handle("DELETE /v1/me/local-credential", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteLocalCredential(accessService)))))
	mux.Handle("DELETE /v1/me/account", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(scheduleAccountDeletion(authenticator, cookies)))))
	mux.Handle("GET /v1/me/tournaments", requireSession(authenticator)(http.HandlerFunc(listAccountTournaments(tournamentService))))
	mux.Handle("GET /v1/me/recent-tournaments", requireSession(authenticator)(http.HandlerFunc(listRecentAccountTournaments(tournamentService))))
	if suggestionService != nil {
		suggestionLimiter := newRequestLimiter(suggestionLimit, suggestionWindow)
		mux.Handle("POST /v1/me/suggestions", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(submitSuggestion(*suggestionService, suggestionLimiter)))))
	}
	if repository, ok := authenticator.(notifications.Repository); ok {
		notificationService := notifications.NewService(repository)
		mux.Handle("GET /v1/me/notifications", requireSession(authenticator)(http.HandlerFunc(listNotifications(notificationService))))
		mux.Handle("GET /v1/me/notifications/unread-count", requireSession(authenticator)(http.HandlerFunc(unreadNotificationCount(notificationService))))
		mux.Handle("POST /v1/me/notifications/read", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(markAllNotificationsRead(notificationService)))))
		mux.Handle("DELETE /v1/me/notifications", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteAllNotifications(notificationService)))))
		mux.Handle("DELETE /v1/me/notifications/{notificationId}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(deleteNotification(notificationService)))))
	}
	followHandler := requireSession(authenticator)(cookieCSRF(http.HandlerFunc(followTournament(tournamentService))))
	mux.Handle("PUT /v1/me/tournaments/{tournamentId}/follow", followHandler)
	mux.Handle("DELETE /v1/me/tournaments/{tournamentId}/follow", followHandler)
	if creationService != nil {
		mux.Handle("POST /v1/tournaments", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createTournament(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/team-invitation", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(createTournamentTeamInvitation(*creationService)))))
		mux.Handle("DELETE /v1/tournaments/{tournamentId}/team-invitation", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(revokeTournamentTeamInvitation(*creationService)))))
		mux.HandleFunc("POST /v1/team-invitations/inspection", inspectTournamentTeamInvitation(*creationService))
		mux.Handle("POST /v1/team-invitations/registration", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(joinTournamentTeamInvitation(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/teams", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(addTournamentTeam(*creationService)))))
		mux.Handle("DELETE /v1/tournaments/{tournamentId}/teams/{teamId}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(removeTournamentTeam(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/teams/{teamId}/withdraw", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(withdrawTournamentTeam(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/start", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(startTournament(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/stages/elimination/start", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(startTournamentElimination(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/cancel", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(cancelTournament(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/complete", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(completeTournament(*creationService)))))
		mux.Handle("PUT /v1/tournaments/{tournamentId}/matches/{matchId}/result", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(recordMatchResult(*creationService)))))
		mux.Handle("PUT /v1/tournaments/{tournamentId}/administrators/{username}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(assignTournamentAdministrator(*creationService)))))
		mux.Handle("GET /v1/tournaments/{tournamentId}/administrators", requireSession(authenticator)(http.HandlerFunc(listTournamentAdministrators(*creationService))))
		mux.Handle("DELETE /v1/tournaments/{tournamentId}/administrators/{username}", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(removeTournamentAdministrator(*creationService)))))
		mux.Handle("POST /v1/tournaments/{tournamentId}/transfer", requireSession(authenticator)(cookieCSRF(http.HandlerFunc(transferTournamentOwnership(*creationService)))))
		mux.HandleFunc("GET /v1/tournaments/{tournamentId}", getPublicTournament(*creationService))
	}
	withCookieName := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routedRequest := r.WithContext(context.WithValue(r.Context(), sessionCookieNameContextKey{}, cookies.name))
		mux.ServeHTTP(w, routedRequest)
		// ServeMux writes the matched template on the request it serves. Preserve
		// it for the outer observability middleware after adding request context.
		r.Pattern = routedRequest.Pattern
	})
	return requireAllowedOrigin(config.CORSAllowedOrigins, withCookieName)
}
