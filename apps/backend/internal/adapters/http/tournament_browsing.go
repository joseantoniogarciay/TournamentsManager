package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func browseTournaments(service tournaments.BrowsingService, personal bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		limit, valid := listLimit(query.Get("limit"))
		cursor := r.URL.Query().Get("cursor")
		if !valid || (query.Has("limit") && limit < 1) || (cursor != "" && !uuidPattern.MatchString(cursor)) || (!personal && !uuidPattern.MatchString(r.PathValue("tournamentId"))) {
			writeTournamentValidationProblem(w, r)
			return
		}
		var value any
		var err error
		if personal {
			account, ok := currentAccountID(r.Context())
			if !ok {
				writeProblem(w, 500, "Could not resolve session")
				return
			}
			value, err = service.PersonalMatches(r.Context(), account, r.URL.Query().Get("view"), cursor, limit)
		} else {
			value, err = service.Editions(r.Context(), r.PathValue("tournamentId"), cursor, limit)
		}
		recordTournamentFailure(r.Context(), err)
		switch {
		case errors.Is(err, tournaments.ErrInvalidPage):
			writeTournamentValidationProblem(w, r)
		case errors.Is(err, tournaments.ErrTournamentNotFound):
			writeProblem(w, 404, "Tournament is unavailable")
		case err != nil:
			if r.Context().Err() == nil {
				writeProblem(w, 500, "Could not retrieve collection")
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			_ = json.NewEncoder(w).Encode(value)
		}
	}
}
