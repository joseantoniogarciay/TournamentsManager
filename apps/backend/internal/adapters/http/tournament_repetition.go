package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func repeatTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 32768)
		account, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var input tournaments.RepeatInput
		if !uuidPattern.MatchString(r.PathValue("tournamentId")) || decodeBody(r, &input) != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		value, err := service.Repeat(r.Context(), account, r.PathValue("tournamentId"), input)
		recordTournamentFailure(r.Context(), err)
		switch {
		case errors.Is(err, tournaments.ErrInvalidTournamentInput):
			writeTournamentValidationProblem(w, r)
		case errors.Is(err, tournaments.ErrTournamentForbidden):
			writeProblem(w, http.StatusForbidden, "You cannot repeat this tournament")
		case errors.Is(err, tournaments.ErrTournamentNotFound):
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
		case errors.Is(err, tournaments.ErrTournamentRepetitionConflict):
			writeProblem(w, http.StatusConflict, "Tournament is not completed")
		case err != nil:
			writeProblem(w, http.StatusInternalServerError, "Could not repeat tournament")
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(value)
		}
	}
}
