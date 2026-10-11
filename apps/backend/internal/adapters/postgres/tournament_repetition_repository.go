package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// Repeat serializes with source ownership changes and creates an independent edition.
func (r AccountTournamentRepository) Repeat(ctx context.Context, accountID, sourceID string, input tournaments.RepeatInput) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owner, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text,state FROM tournaments WHERE id=$1 FOR UPDATE`, sourceID).Scan(&owner, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if owner != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "completed" {
		return tournaments.Tournament{}, tournaments.ErrTournamentRepetitionConflict
	}
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM tournaments WHERE repetition_source_id=$1 AND repetition_account_id=$2 AND repetition_request_id=$3`, sourceID, account, input.RequestID).Scan(&id)
	if err == nil {
		value, err := readTournament(ctx, tx, id)
		if err != nil {
			return tournaments.Tournament{}, err
		}
		return value, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, err
	}
	source, err := readTournament(ctx, tx, sourceID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	config := tournaments.RepetitionStartConfiguration(source)
	if err := tx.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id,name,sport,best_of_sets,points_per_game,state,published_at,repetition_source_id,repetition_account_id,repetition_request_id,suggested_start) VALUES ($1,$2,$3,NULLIF($4,0),NULLIF($5,0),'published',now(),$6,$1,$7,$8) RETURNING id::text`, account, input.Name, source.Sport, source.BestOfSets, source.PointsPerGame, sourceID, input.RequestID, config).Scan(&id); err != nil {
		return tournaments.Tournament{}, err
	}
	var firstTeamID string
	for i, team := range input.Teams {
		var teamID string
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id,name,name_normalized,position) VALUES ($1,$2,lower($2),$3) RETURNING id::text`, id, team.Name, i+1).Scan(&teamID); err != nil {
			return tournaments.Tournament{}, err
		}
		if i == 0 {
			firstTeamID = teamID
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_team_accounts (tournament_id,team_id,account_id) VALUES ($1,$2,$3)`, id, firstTeamID, account); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET last_team_name=$2 WHERE id=$1`, account, input.Teams[0].Name); err != nil {
		return tournaments.Tournament{}, err
	}
	value, err := readTournament(ctx, tx, id)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	return value, tx.Commit(ctx)
}
