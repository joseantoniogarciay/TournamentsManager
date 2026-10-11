package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func setTournamentScheduling(service tournaments.SchedulingService, match bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		id, matchID := r.PathValue("tournamentId"), r.PathValue("matchId")
		if !uuidPattern.MatchString(id) || match && !uuidPattern.MatchString(matchID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		var value tournaments.Tournament
		var err error
		if match {
			var body struct {
				ScheduledDate *string `json:"scheduledDate"`
				ScheduledTime *string `json:"scheduledTime"`
				TimeZone      *string `json:"timeZone"`
				Venue         *string `json:"venue"`
			}
			if decodeBody(r, &body) != nil || body.ScheduledDate == nil || body.ScheduledTime == nil || body.TimeZone == nil || body.Venue == nil {
				writeTournamentValidationProblem(w, r)
				return
			}
			value, err = service.SetMatchSchedule(r.Context(), accountID, id, matchID, tournaments.ScheduleInput{ScheduledDate: *body.ScheduledDate, ScheduledTime: *body.ScheduledTime, TimeZone: *body.TimeZone, Venue: *body.Venue})
		} else {
			var body struct {
				Venue *string `json:"venue"`
			}
			if decodeBody(r, &body) != nil || body.Venue == nil {
				writeTournamentValidationProblem(w, r)
				return
			}
			value, err = service.SetVenue(r.Context(), accountID, id, *body.Venue)
		}
		recordTournamentFailure(r.Context(), err)
		switch {
		case errors.Is(err, tournaments.ErrInvalidTournamentInput):
			writeTournamentValidationProblem(w, r)
		case errors.Is(err, tournaments.ErrTournamentForbidden):
			writeProblem(w, http.StatusForbidden, "You cannot edit scheduling for this tournament")
		case errors.Is(err, tournaments.ErrTournamentNotFound):
			writeProblem(w, http.StatusNotFound, "Tournament or match is unavailable")
		case errors.Is(err, tournaments.ErrSchedulingConflict):
			writeProblem(w, http.StatusConflict, "Scheduling is read-only")
		case err != nil:
			writeProblem(w, http.StatusInternalServerError, "Could not save scheduling")
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(value)
		}
	}
}
