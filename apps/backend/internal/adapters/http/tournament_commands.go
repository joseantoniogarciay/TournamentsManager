package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func createTournamentTeamInvitation(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		tournamentID := r.PathValue("tournamentId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(tournamentID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		token, err := service.CreateTeamInvitation(r.Context(), accountID, tournamentID)
		recordTournamentFailure(r.Context(), err)
		switch {
		case errors.Is(err, tournaments.ErrTournamentForbidden):
			writeProblem(w, http.StatusForbidden, "You cannot create invitations for this tournament")
			return
		case errors.Is(err, tournaments.ErrTournamentNotFound):
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		case errors.Is(err, tournaments.ErrTournamentInvitationConflict):
			writeProblem(w, http.StatusConflict, "Tournament no longer accepts invitations")
			return
		case err != nil:
			writeProblem(w, http.StatusInternalServerError, "Could not create invitation")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func revokeTournamentTeamInvitation(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		tournamentID := r.PathValue("tournamentId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(tournamentID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		err := service.RevokeTeamInvitation(r.Context(), accountID, tournamentID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot revoke invitations for this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not revoke invitation")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func inspectTournamentTeamInvitation(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"`
		}
		if decodeBody(r, &body) != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		invitation, err := service.InspectTeamInvitation(r.Context(), body.Token)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
			writeProblem(w, http.StatusNotFound, "Invitation is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not inspect invitation")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(invitation)
	}
}

func joinTournamentTeamInvitation(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var request struct {
			Token string `json:"token"`
			Name  string `json:"name"`
		}
		if decodeBody(r, &request) != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		registration, err := service.JoinTeamInvitation(r.Context(), accountID, request.Token, tournaments.TeamInput{Name: request.Name})
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
			writeProblem(w, http.StatusNotFound, "Invitation is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
			writeProblem(w, http.StatusConflict, "Invitation cannot accept this team")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not join tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(registration)
	}
}

type leagueInput struct {
	DraftID    string            `json:"draftId"`
	Name       string            `json:"name"`
	Sport      tournaments.Sport `json:"sport"`
	BestOfSets int               `json:"bestOfSets"`
	Teams      []struct {
		Name string `json:"name"`
	} `json:"teams"`
}
type startTournamentInput struct {
	Format             string `json:"format"`
	RoundRobinLegs     int    `json:"roundRobinLegs"`
	LeagueStructure    string `json:"leagueStructure"`
	QualifierCount     int    `json:"qualifierCount"`
	GroupCount         int    `json:"groupCount"`
	QualifiersPerGroup int    `json:"qualifiersPerGroup"`
}
type teamInput struct {
	Name string `json:"name"`
}
type matchResultInput struct {
	HomePenalties *int                    `json:"homePenalties"`
	AwayPenalties *int                    `json:"awayPenalties"`
	HomeScore     *int                    `json:"homeScore"`
	AwayScore     *int                    `json:"awayScore"`
	Sets          *[]tournaments.SetScore `json:"sets"`
}

func createTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var body leagueInput
		if err := decodeBody(r, &body); err != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		teams := make([]tournaments.TeamInput, len(body.Teams))
		for i, team := range body.Teams {
			teams[i] = tournaments.TeamInput{Name: team.Name}
		}
		league, err := service.Create(r.Context(), accountID, tournaments.CreateInput{Name: body.Name, Sport: body.Sport, BestOfSets: body.BestOfSets, Teams: teams})
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not create tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(league)
	}
}
func addTournamentTeam(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("tournamentId")
		var body teamInput
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || decodeBody(r, &body) != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		team, err := service.AddTeam(r.Context(), accountID, leagueID, tournaments.TeamInput{Name: body.Name})
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot modify this tournament's teams")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentTeamConflict) {
			writeProblem(w, http.StatusConflict, "Tournament cannot accept this team")
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
func removeTournamentTeam(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, teamID := r.PathValue("tournamentId"), r.PathValue("teamId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(teamID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		err := service.RemoveTeam(r.Context(), accountID, leagueID, teamID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot modify this tournament's teams")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament or team is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentTeamConflict) {
			writeProblem(w, http.StatusConflict, "An unstarted tournament must keep at least one team")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not remove team")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func withdrawTournamentTeam(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, teamID := r.PathValue("tournamentId"), r.PathValue("teamId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(teamID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		league, err := service.WithdrawTeam(r.Context(), accountID, leagueID, teamID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot withdraw teams from this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament or team is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentWithdrawalConflict) {
			writeProblem(w, http.StatusConflict, "Team cannot be withdrawn from this tournament")
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
func getPublicTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("tournamentId")
		if !uuidPattern.MatchString(leagueID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		league, err := service.GetPublic(r.Context(), leagueID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}
func startTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		leagueID := r.PathValue("tournamentId")
		var body startTournamentInput
		if !uuidPattern.MatchString(leagueID) || decodeBody(r, &body) != nil || (body.Format != "league" && body.Format != "single_elimination" && body.Format != tournaments.FormatLeagueThenSingleElimination) {
			writeTournamentValidationProblem(w, r)
			return
		}
		league, err := service.Start(r.Context(), accountID, leagueID, tournaments.StartInput{Format: body.Format, RoundRobinLegs: body.RoundRobinLegs, LeagueStructure: body.LeagueStructure, QualifierCount: body.QualifierCount, GroupCount: body.GroupCount, QualifiersPerGroup: body.QualifiersPerGroup})
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot start this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentConflict) {
			writeProblem(w, http.StatusConflict, "Tournament is no longer unstarted")
			return
		}
		if errors.Is(err, tournaments.ErrInvalidMixedConfiguration) {
			writeProblem(w, http.StatusUnprocessableEntity, "Tournament configuration does not match its teams")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not start tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func startTournamentElimination(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		tournamentID := r.PathValue("tournamentId")
		if !uuidPattern.MatchString(tournamentID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		tournament, err := service.StartElimination(r.Context(), accountID, tournamentID)
		recordTournamentFailure(r.Context(), err)
		switch {
		case errors.Is(err, tournaments.ErrTournamentForbidden):
			writeProblem(w, http.StatusForbidden, "You cannot start this tournament stage")
			return
		case errors.Is(err, tournaments.ErrTournamentNotFound):
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		case errors.Is(err, tournaments.ErrTournamentStageTransitionConflict):
			writeProblem(w, http.StatusConflict, "The qualifying stage is not ready")
			return
		case err != nil:
			writeProblem(w, http.StatusInternalServerError, "Could not start tournament stage")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tournament)
	}
}

func cancelTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		leagueID := r.PathValue("tournamentId")
		if !uuidPattern.MatchString(leagueID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		league, err := service.Cancel(r.Context(), accountID, leagueID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot cancel this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentCancellationConflict) {
			writeProblem(w, http.StatusConflict, "Tournament cannot be cancelled from its current state")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not cancel tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func completeTournament(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("tournamentId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		league, err := service.Complete(r.Context(), accountID, leagueID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot complete this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentCompletionConflict) {
			writeProblem(w, http.StatusConflict, "Tournament cannot be completed yet")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not complete tournament")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(league)
	}
}

func recordMatchResult(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, matchID := r.PathValue("tournamentId"), r.PathValue("matchId")
		var body matchResultInput
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !uuidPattern.MatchString(matchID) || decodeBody(r, &body) != nil {
			writeTournamentValidationProblem(w, r)
			return
		}
		hasSetResult := body.Sets != nil && len(*body.Sets) > 0
		hasScoreResult := body.HomeScore != nil && body.AwayScore != nil
		mixedResult := body.Sets != nil && (body.HomeScore != nil || body.AwayScore != nil || body.HomePenalties != nil || body.AwayPenalties != nil)
		if mixedResult || hasSetResult == hasScoreResult {
			writeTournamentValidationProblem(w, r)
			return
		}
		input := tournaments.MatchResultInput{HomePenalties: body.HomePenalties, AwayPenalties: body.AwayPenalties}
		if body.Sets != nil {
			input.Sets = *body.Sets
		}
		if body.HomeScore != nil {
			input.HomeScore = *body.HomeScore
		}
		if body.AwayScore != nil {
			input.AwayScore = *body.AwayScore
		}
		league, err := service.RecordResult(r.Context(), accountID, leagueID, matchID, input)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrInvalidTournamentInput) || errors.Is(err, tournaments.ErrInvalidBracketResult) {
			writeTournamentValidationProblem(w, r)
			return
		}
		if errors.Is(err, tournaments.ErrMatchResultForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot record results for this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament or match is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrMatchResultConflict) || errors.Is(err, tournaments.ErrBracketResultDependency) || errors.Is(err, tournaments.ErrBracketMatchNotReady) {
			writeProblem(w, http.StatusConflict, "Tournament is not in progress")
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

func assignTournamentAdministrator(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, username := r.PathValue("tournamentId"), r.PathValue("username")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !usernamePattern.MatchString(username) {
			writeTournamentValidationProblem(w, r)
			return
		}
		err := service.AssignAdministrator(r.Context(), accountID, leagueID, username)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot assign administrators for this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament or account is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentAdministratorConflict) {
			writeProblem(w, http.StatusConflict, "Tournament owner cannot be a delegated administrator")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not assign administrator")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listTournamentAdministrators(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("tournamentId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeTournamentValidationProblem(w, r)
			return
		}
		usernames, err := service.ListAdministrators(r.Context(), accountID, leagueID)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot view administrators for this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
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

func removeTournamentAdministrator(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID, username := r.PathValue("tournamentId"), r.PathValue("username")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		if !uuidPattern.MatchString(leagueID) || !usernamePattern.MatchString(username) {
			writeTournamentValidationProblem(w, r)
			return
		}
		err := service.RemoveAdministrator(r.Context(), accountID, leagueID, username)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot remove administrators for this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament is unavailable")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not remove administrator")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func transferTournamentOwnership(service tournaments.CreationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		leagueID := r.PathValue("tournamentId")
		if !ok {
			writeProblem(w, http.StatusInternalServerError, "Could not resolve session")
			return
		}
		var body struct {
			Username string `json:"username"`
		}
		if !uuidPattern.MatchString(leagueID) || json.NewDecoder(r.Body).Decode(&body) != nil || !usernamePattern.MatchString(body.Username) {
			writeTournamentValidationProblem(w, r)
			return
		}
		err := service.TransferOwnership(r.Context(), accountID, leagueID, body.Username)
		recordTournamentFailure(r.Context(), err)
		if errors.Is(err, tournaments.ErrTournamentForbidden) {
			writeProblem(w, http.StatusForbidden, "You cannot transfer this tournament")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentNotFound) {
			writeProblem(w, http.StatusNotFound, "Tournament or account is unavailable")
			return
		}
		if errors.Is(err, tournaments.ErrTournamentOwnershipTransferConflict) {
			writeProblem(w, http.StatusConflict, "Tournament owner cannot receive the transfer")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not transfer tournament")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeTournamentValidationProblem(w http.ResponseWriter, r *http.Request) {
	observability.RecordEndpointFailure(r.Context(), "validation.rejected")
	writeValidationProblem(w)
}

// recordTournamentFailure preserves the feature's closed vocabulary on the HTTP
// span. It intentionally carries neither resource identifiers nor usernames.
func recordTournamentFailure(ctx context.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, tournaments.ErrInvalidTournamentInput), errors.Is(err, tournaments.ErrInvalidRelationship), errors.Is(err, tournaments.ErrInvalidPage):
		observability.RecordEndpointFailure(ctx, "validation.rejected")
	case errors.Is(err, tournaments.ErrTournamentNotFound):
		observability.RecordEndpointFailure(ctx, "tournament.not_found")
	case errors.Is(err, tournaments.ErrTournamentForbidden), errors.Is(err, tournaments.ErrMatchResultForbidden):
		observability.RecordEndpointFailure(ctx, "tournament.forbidden")
	case errors.Is(err, tournaments.ErrTournamentConflict):
		observability.RecordEndpointFailure(ctx, "tournament.start_conflict")
	case errors.Is(err, tournaments.ErrInvalidMixedConfiguration):
		observability.RecordEndpointFailure(ctx, "tournament.configuration_rejected")
	case errors.Is(err, tournaments.ErrTournamentStageTransitionConflict):
		observability.RecordEndpointFailure(ctx, "tournament.stage_transition_conflict")
	case errors.Is(err, tournaments.ErrTournamentCancellationConflict):
		observability.RecordEndpointFailure(ctx, "tournament.cancellation_conflict")
	case errors.Is(err, tournaments.ErrTournamentCompletionConflict):
		observability.RecordEndpointFailure(ctx, "tournament.completion_conflict")
	case errors.Is(err, tournaments.ErrTournamentTeamConflict):
		observability.RecordEndpointFailure(ctx, "tournament.team_conflict")
	case errors.Is(err, tournaments.ErrTournamentInvitationNotFound):
		observability.RecordEndpointFailure(ctx, "tournament.invitation_not_found")
	case errors.Is(err, tournaments.ErrTournamentInvitationConflict):
		observability.RecordEndpointFailure(ctx, "tournament.invitation_conflict")
	case errors.Is(err, tournaments.ErrTournamentWithdrawalConflict):
		observability.RecordEndpointFailure(ctx, "tournament.withdrawal_conflict")
	case errors.Is(err, tournaments.ErrBracketResultDependency):
		observability.RecordEndpointFailure(ctx, "bracket.result_dependency")
	case errors.Is(err, tournaments.ErrBracketMatchNotReady):
		observability.RecordEndpointFailure(ctx, "bracket.match_not_ready")
	case errors.Is(err, tournaments.ErrInvalidBracketResult):
		observability.RecordEndpointFailure(ctx, "validation.rejected")
	case errors.Is(err, tournaments.ErrMatchResultConflict):
		observability.RecordEndpointFailure(ctx, "tournament.result_conflict")
	case errors.Is(err, tournaments.ErrTournamentAdministratorConflict):
		observability.RecordEndpointFailure(ctx, "tournament.administrator_conflict")
	case errors.Is(err, tournaments.ErrTournamentOwnershipTransferConflict):
		observability.RecordEndpointFailure(ctx, "tournament.ownership_transfer_conflict")
	default:
		observability.RecordDatabaseEndpointFailure(ctx, err)
	}
}
