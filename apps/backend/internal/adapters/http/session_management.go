package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/sessions"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// Categories are advisory and never store raw User-Agent or client hints.
func sessionClient(r *http.Request, transport sessionTransport) sessions.Client {
	client := sessions.Client{Platform: "unknown", Kind: "unknown"}
	ua := strings.ToLower(r.UserAgent())
	switch {
	case strings.Contains(ua, "android"):
		client.Platform = "android"
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"):
		client.Platform = "ios"
	case strings.Contains(ua, "windows"):
		client.Platform = "windows"
	case strings.Contains(ua, "macintosh"), strings.Contains(ua, "mac os"):
		client.Platform = "macos"
	case strings.Contains(ua, "linux"):
		client.Platform = "linux"
	}
	if transport == bearerSession {
		client.Kind = "app"
		switch r.Header.Get("X-TM-Client-Platform") {
		case "ios":
			client.Platform = "ios"
		case "android":
			client.Platform = "android"
		}
		return client
	}
	switch {
	case strings.Contains(ua, "edg/"):
		client.Kind = "edge"
	case strings.Contains(ua, "firefox/"), strings.Contains(ua, "fxios/"):
		client.Kind = "firefox"
	case strings.Contains(ua, "chrome/"), strings.Contains(ua, "crios/"):
		client.Kind = "chrome"
	case strings.Contains(ua, "safari/"):
		client.Kind = "safari"
	default:
		client.Kind = "web"
	}
	return client
}

func manageSessions(authenticator sessionAuthenticator, revoke, all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cursor, target := r.URL.Query().Get("cursor"), ""
		if revoke && !all {
			target = r.PathValue("sessionId")
		}
		if (!revoke && (r.URL.Query().Has("cursor") && !uuidPattern.MatchString(cursor))) || (revoke && !all && !uuidPattern.MatchString(target)) {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeValidationProblem(w)
			return
		}
		repository, ok := authenticator.(sessions.Repository)
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not manage sessions")
			return
		}
		token, _ := currentSessionToken(r.Context())
		w.Header().Set("Cache-Control", "no-store")
		if revoke {
			if err := repository.RevokeOtherSessions(r.Context(), token, target); err != nil {
				sessionManagementFailure(w, r, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		page, err := repository.ListSessions(r.Context(), token, cursor)
		if err != nil {
			sessionManagementFailure(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}
}

func sessionManagementFailure(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, tournaments.ErrUnauthenticated) {
		observability.RecordEndpointFailure(r.Context(), "session.invalid")
		writeProblem(w, http.StatusUnauthorized, "Invalid session")
		return
	}
	observability.RecordDatabaseEndpointFailure(r.Context(), err)
	if errors.Is(r.Context().Err(), context.Canceled) {
		return
	}
	writeProblem(w, http.StatusInternalServerError, "Could not manage sessions")
}
