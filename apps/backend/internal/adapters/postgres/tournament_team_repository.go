package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// AddTeam adds a team while the tournament still accepts composition changes.
func (r AccountTournamentRepository) AddTeam(ctx context.Context, accountID, leagueID string, input tournaments.TeamInput) (tournaments.Team, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Team{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Team{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Team{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Team{}, err
	}
	if organizer != account.String() {
		return tournaments.Team{}, tournaments.ErrTournamentForbidden
	}
	if state != "published" {
		return tournaments.Team{}, tournaments.ErrTournamentTeamConflict
	}
	var position, count int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0), COUNT(*) FROM tournament_teams WHERE tournament_id = $1`, leagueID).Scan(&position, &count); err != nil {
		return tournaments.Team{}, err
	}
	if count >= 64 {
		return tournaments.Team{}, tournaments.ErrTournamentTeamConflict
	}
	var team tournaments.Team
	if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, leagueID, input.Name, position+1).Scan(&team.ID, &team.Name, &team.Position); err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return tournaments.Team{}, tournaments.ErrTournamentTeamConflict
		}
		return tournaments.Team{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Team{}, err
	}
	return team, nil
}

// RemoveTeam removes a team from a published tournament while retaining one entrant.
func (r AccountTournamentRepository) RemoveTeam(ctx context.Context, accountID, leagueID, teamID string) error {
	account, err := uuidValue(accountID)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	if state != "published" {
		return tournaments.ErrTournamentTeamConflict
	}
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_teams WHERE tournament_id = $1 AND id = $2)`, leagueID, teamID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return tournaments.ErrTournamentNotFound
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM tournament_teams WHERE tournament_id = $1`, leagueID).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return tournaments.ErrTournamentTeamConflict
	}
	command, err := tx.Exec(ctx, `DELETE FROM tournament_teams WHERE tournament_id = $1 AND id = $2`, leagueID, teamID)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return tournaments.ErrTournamentNotFound
	}
	return tx.Commit(ctx)
}

// GetPublic returns the visible projection of an existing league.
