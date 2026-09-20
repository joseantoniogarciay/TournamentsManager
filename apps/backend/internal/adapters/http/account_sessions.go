package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/accounts"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

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
		if errors.Is(err, accounts.ErrAccountHasOwnedTournaments) {
			observability.RecordEndpointFailure(r.Context(), "account.deletion_owned_tournaments")
			writeProblem(w, http.StatusConflict, "Account cannot be deleted while it owns tournaments")
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

func sessionUserResponse(accountID, username, lastTeamName string) map[string]string {
	user := map[string]string{"id": accountID, "username": username}
	if lastTeamName != "" {
		user["lastTeamName"] = lastTeamName
	}
	return user
}

func getCurrentSession(authenticator sessionAuthenticator) http.HandlerFunc {
	type currentSessionReader interface {
		GetCurrentSession(context.Context, string) (tournaments.CurrentSession, error)
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
		if errors.Is(err, tournaments.ErrUnauthenticated) {
			observability.RecordEndpointFailure(r.Context(), "session.invalid")
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
			"user":              sessionUserResponse(session.AccountID, session.Username, session.LastTeamName),
			"idleExpiresAt":     session.IdleExpiresAt,
			"absoluteExpiresAt": session.AbsoluteExpiresAt,
		})
	}
}
