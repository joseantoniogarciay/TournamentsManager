package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// SaveTeamInvitation persists the current invitation secret hash.
func (r AccountTournamentRepository) SaveTeamInvitation(ctx context.Context, accountID, tournamentID string, tokenHash tournaments.InvitationTokenHash) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, tournamentID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	if state != "published" {
		return tournaments.ErrTournamentInvitationConflict
	}
	var teamCount int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM tournament_teams WHERE tournament_id = $1`, tournamentID).Scan(&teamCount); err != nil {
		return err
	}
	if teamCount >= 64 {
		return tournaments.ErrTournamentInvitationConflict
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_team_invitations (tournament_id, token_hash) VALUES ($1, $2) ON CONFLICT (tournament_id) DO UPDATE SET token_hash = EXCLUDED.token_hash, created_at = now()`, tournamentID, tokenHash[:]); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RevokeTeamInvitation idempotently removes an owner's active invitation.
func (r AccountTournamentRepository) RevokeTeamInvitation(ctx context.Context, accountID, tournamentID string) error {
	account, err := uuidValue(accountID)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, tournamentID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_team_invitations WHERE tournament_id = $1`, tournamentID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// InspectTeamInvitation returns only the safe tournament projection for an active invitation.
func (r AccountTournamentRepository) InspectTeamInvitation(ctx context.Context, tokenHash tournaments.InvitationTokenHash) (tournaments.TeamInvitation, error) {
	var invitation tournaments.TeamInvitation
	err := r.pool.QueryRow(ctx, `SELECT tournaments.id::text, tournaments.name FROM tournament_team_invitations JOIN tournaments ON tournaments.id = tournament_team_invitations.tournament_id WHERE tournament_team_invitations.token_hash = $1 AND tournaments.state = 'published' AND (SELECT COUNT(*) FROM tournament_teams WHERE tournament_id = tournaments.id) < 64`, tokenHash[:]).Scan(&invitation.TournamentID, &invitation.TournamentName)
	if errors.Is(err, pgx.ErrNoRows) {
		return tournaments.TeamInvitation{}, tournaments.ErrTournamentInvitationNotFound
	}
	return invitation, err
}

// JoinTeamInvitation creates the team, account link, and follow relationship atomically.
func (r AccountTournamentRepository) JoinTeamInvitation(ctx context.Context, accountID string, tokenHash tournaments.InvitationTokenHash, input tournaments.TeamInput) (tournaments.TeamRegistration, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.TeamRegistration{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.TeamRegistration{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var registration tournaments.TeamRegistration
	var state, organizer string
	if err := tx.QueryRow(ctx, `SELECT tournaments.id::text, tournaments.state, tournaments.organizer_account_id::text FROM tournament_team_invitations JOIN tournaments ON tournaments.id = tournament_team_invitations.tournament_id WHERE tournament_team_invitations.token_hash = $1 FOR UPDATE OF tournaments`, tokenHash[:]).Scan(&registration.TournamentID, &state, &organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationNotFound
	} else if err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if state != "published" {
		return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationNotFound
	}
	// Historical tournaments predate the account-team relation, so the explicit
	// ownership check also prevents their organizer from registering twice.
	if organizer == account.String() {
		return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationConflict
	}
	var alreadyJoined bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_team_accounts WHERE tournament_id = $1 AND account_id = $2)`, registration.TournamentID, account).Scan(&alreadyJoined); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if alreadyJoined {
		return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationConflict
	}
	var position, count int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0), COUNT(*) FROM tournament_teams WHERE tournament_id = $1`, registration.TournamentID).Scan(&position, &count); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if count >= 64 {
		return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationConflict
	}
	if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, registration.TournamentID, input.Name, position+1).Scan(&registration.Team.ID, &registration.Team.Name, &registration.Team.Position); err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationConflict
		}
		return tournaments.TeamRegistration{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_team_accounts (tournament_id, team_id, account_id) VALUES ($1, $2, $3)`, registration.TournamentID, registration.Team.ID, account); err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return tournaments.TeamRegistration{}, tournaments.ErrTournamentInvitationConflict
		}
		return tournaments.TeamRegistration{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_followers (tournament_id, account_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, registration.TournamentID, account); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET last_team_name = $2 WHERE id = $1`, account, registration.Team.Name); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at = now() WHERE id = $1`, registration.TournamentID); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.TeamRegistration{}, err
	}
	return registration, nil
}

// AddTeam adds a team to one of its owner's published tournaments.
