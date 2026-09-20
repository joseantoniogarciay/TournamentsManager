package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func followTournament(service tournaments.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		leagueID := request.PathValue("tournamentId")
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeTournamentValidationProblem(writer, request)
			return
		}
		var err error
		if request.Method == http.MethodPut {
			err = service.Follow(request.Context(), accountID, leagueID)
		} else {
			err = service.Unfollow(request.Context(), accountID, leagueID)
		}
		recordTournamentFailure(request.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(writer, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not update follow status")
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func listAccountTournaments(service tournaments.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}

		relationship := tournaments.Relationship(request.URL.Query().Get("relationship"))
		limit, valid := listLimit(request.URL.Query().Get("limit"))
		cursor := request.URL.Query().Get("cursor")
		if !valid || (cursor != "" && !uuidPattern.MatchString(cursor)) {
			writeTournamentValidationProblem(writer, request)
			return
		}
		page, err := service.List(request.Context(), accountID, relationship, cursor, limit)
		recordTournamentFailure(request.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidRelationship) || errors.Is(err, tournaments.ErrInvalidPage) {
			writeTournamentValidationProblem(writer, request)
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not retrieve tournaments")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(page)
	}
}

func listRecentAccountTournaments(service tournaments.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		items, err := service.ListRecent(request.Context(), accountID)
		recordTournamentFailure(request.Context(), err)
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Could not retrieve recent tournaments")
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
