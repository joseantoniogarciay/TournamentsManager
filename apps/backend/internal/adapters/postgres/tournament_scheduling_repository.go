package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// SetVenue atomically authorizes and replaces the common tournament location.
func (r AccountTournamentRepository) SetVenue(ctx context.Context, accountID, tournamentID, venue string) (tournaments.Tournament, error) {
	return r.saveScheduling(ctx, accountID, tournamentID, "", tournaments.ScheduleInput{Venue: venue})
}

// SetMatchSchedule replaces the informational metadata without touching sporting data.
func (r AccountTournamentRepository) SetMatchSchedule(ctx context.Context, accountID, tournamentID, matchID string, input tournaments.ScheduleInput) (tournaments.Tournament, error) {
	return r.saveScheduling(ctx, accountID, tournamentID, matchID, input)
}

func (r AccountTournamentRepository) saveScheduling(ctx context.Context, accountID, tournamentID, matchID string, input tournaments.ScheduleInput) (tournaments.Tournament, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	err = tx.QueryRow(ctx, `SELECT organizer_account_id::text,state FROM tournaments WHERE id=$1 FOR UPDATE`, tournamentID).Scan(&organizer, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	}
	if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != accountID {
		var delegated bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tournament_administrators WHERE tournament_id=$1 AND account_id=$2)`, tournamentID, accountID).Scan(&delegated); err != nil {
			return tournaments.Tournament{}, err
		}
		if !delegated {
			return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
		}
	}
	if state != "published" && state != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrSchedulingConflict
	}
	if matchID == "" {
		_, err = tx.Exec(ctx, `UPDATE tournaments SET venue=$2,last_activity_at=now() WHERE id=$1`, tournamentID, input.Venue)
	} else {
		var matchState string
		err = tx.QueryRow(ctx, `SELECT state FROM matches WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, matchID, tournamentID).Scan(&matchState)
		if errors.Is(err, pgx.ErrNoRows) {
			return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
		}
		if err != nil {
			return tournaments.Tournament{}, err
		}
		if matchState == "bye" {
			return tournaments.Tournament{}, tournaments.ErrSchedulingConflict
		}
		_, err = tx.Exec(ctx, `UPDATE matches SET scheduled_date=$3,scheduled_time=$4,time_zone=$5,venue=$6 WHERE id=$1 AND tournament_id=$2`, matchID, tournamentID, input.ScheduledDate, input.ScheduledTime, input.TimeZone, input.Venue)
		if err == nil {
			_, err = tx.Exec(ctx, `UPDATE tournaments SET last_activity_at=now() WHERE id=$1`, tournamentID)
		}
	}
	if err != nil {
		return tournaments.Tournament{}, err
	}
	value, err := readTournament(ctx, tx, tournamentID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return value, nil
}
