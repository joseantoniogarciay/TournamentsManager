package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/accounts"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// ErrAccountHasOwnedTournaments indicates that the account still owns one or more tournaments.
var ErrAccountHasOwnedTournaments = errors.New("account has owned tournaments")

// AccountTournamentRepository persists sessions and league relationships.
type AccountTournamentRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewAccountTournamentRepository builds the account-relationship adapter.
func NewAccountTournamentRepository(pool *pgxpool.Pool) AccountTournamentRepository {
	return AccountTournamentRepository{pool: pool, queries: sqlc.New(pool)}
}

// PurgeExpired removes a batch of accounts whose deletion window has expired.
// The CTE locks only selected accounts, and SKIP LOCKED avoids waiting for a
// concurrent transaction on one of them.
func (r AccountTournamentRepository) PurgeExpired(ctx context.Context, limit int) (int64, error) {
	if limit < 1 {
		return 0, fmt.Errorf("límite de purga debe ser positivo")
	}
	if _, err := r.pool.Exec(ctx, `DELETE FROM google_risc_events WHERE expires_at <= now()`); err != nil {
		return 0, err
	}
	command, err := r.pool.Exec(ctx, `
		WITH candidates AS (
			SELECT id
			FROM accounts
			WHERE state = 'deletion_pending'
				AND deletion_requested_at <= now() - interval '30 days'
				AND NOT EXISTS (
					SELECT 1
					FROM tournaments
					WHERE organizer_account_id = accounts.id
				)
			ORDER BY deletion_requested_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		, blocked_legal_evidence AS (
			UPDATE legal_account_acceptances
			SET retention_until = now() + interval '5 years', updated_at = now()
			WHERE account_id IN (SELECT id FROM candidates)
		)
		, expired_legal_evidence AS (
			DELETE FROM legal_account_acceptances
			WHERE retention_until IS NOT NULL
				AND retention_until <= now()
		)
		DELETE FROM accounts
		USING candidates
		WHERE accounts.id = candidates.id`, limit)
	if err != nil {
		return 0, err
	}
	return command.RowsAffected(), nil
}

var _ accounts.PurgeRepository = AccountTournamentRepository{}

// ScheduleAccountDeletion revokes access and personal relationships without deleting the account.
func (r AccountTournamentRepository) ScheduleAccountDeletion(ctx context.Context, accountID string) (time.Time, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var hasOwned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournaments WHERE organizer_account_id = $1)`, accountID).Scan(&hasOwned); err != nil {
		return time.Time{}, err
	}
	if hasOwned {
		return time.Time{}, ErrAccountHasOwnedTournaments
	}
	var requested time.Time
	if err := tx.QueryRow(ctx, `UPDATE accounts SET state = 'deletion_pending', deletion_requested_at = now() WHERE id = $1 AND state = 'verified' RETURNING deletion_requested_at`, accountID).Scan(&requested); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_followers WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return time.Time{}, err
	}
	return requested.AddDate(0, 0, 30).UTC(), nil
}

// Create creates a published league without a schedule yet.
func (r AccountTournamentRepository) Create(ctx context.Context, accountID string, input tournaments.CreateInput) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var league tournaments.Tournament
	if err := tx.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, state, published_at) VALUES ($1, $2, 'published', now()) RETURNING id::text, name, sport, format, state`, account, input.Name).Scan(&league.ID, &league.Name, &league.Sport, &league.Format, &league.State); err != nil {
		return tournaments.Tournament{}, err
	}
	league.Teams = make([]tournaments.Team, len(input.Teams))
	for i, team := range input.Teams {
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, league.ID, team.Name, i+1).Scan(&league.Teams[i].ID, &league.Teams[i].Name, &league.Teams[i].Position); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	league.Matches = []tournaments.Match{}
	return league, nil
}

// AddTeam adds a team to one of its owner's published tournaments.
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

// RemoveTeam removes a team from a published league without reducing it below two entrants.
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
	if count <= 2 {
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
func (r AccountTournamentRepository) GetPublic(ctx context.Context, leagueID string) (tournaments.Tournament, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	value, err := readTournament(ctx, tx, leagueID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	return value, tx.Commit(ctx)
}

// WithdrawTeam keeps the roster for historical standings and assigns each opponent a 3-0 win.
func (r AccountTournamentRepository) WithdrawTeam(ctx context.Context, accountID, leagueID, teamID string) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrTournamentWithdrawalConflict
	}
	var format string
	if err := tx.QueryRow(ctx, `SELECT format FROM tournaments WHERE id=$1`, leagueID).Scan(&format); err != nil {
		return tournaments.Tournament{}, err
	}
	if format != "league" {
		return tournaments.Tournament{}, tournaments.ErrTournamentWithdrawalConflict
	}
	var withdrawn bool
	if err := tx.QueryRow(ctx, `SELECT withdrawn_at IS NOT NULL FROM tournament_teams WHERE tournament_id = $1 AND id = $2 FOR UPDATE`, leagueID, teamID).Scan(&withdrawn); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if withdrawn {
		return tournaments.Tournament{}, tournaments.ErrTournamentWithdrawalConflict
	}
	rows, err := tx.Query(ctx, `SELECT id::text, home_team_id::text, home_score, away_score FROM matches WHERE tournament_id = $1 AND (home_team_id = $2 OR away_team_id = $2) FOR UPDATE`, leagueID, teamID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer rows.Close()
	type scoreChange struct {
		id                         string
		homeID                     string
		previousHome, previousAway *int
	}
	changes := []scoreChange{}
	for rows.Next() {
		var change scoreChange
		if err := rows.Scan(&change.id, &change.homeID, &change.previousHome, &change.previousAway); err != nil {
			return tournaments.Tournament{}, err
		}
		changes = append(changes, change)
	}
	if err := rows.Err(); err != nil {
		return tournaments.Tournament{}, err
	}
	for _, change := range changes {
		var homeScore, awayScore int
		if change.homeID == teamID {
			homeScore, awayScore = 0, 3
		} else {
			homeScore, awayScore = 3, 0
		}
		if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $2, away_score = $3 WHERE id = $1`, change.id, homeScore, awayScore); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score) VALUES ($1, $2, $3, $4, $5, $6)`, change.id, account, change.previousHome, change.previousAway, homeScore, awayScore); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_teams SET withdrawn_at = now() WHERE tournament_id = $1 AND id = $2`, leagueID, teamID); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// RecordResult stores the score and a history entry in a single transaction.
func (r AccountTournamentRepository) RecordResult(ctx context.Context, accountID, leagueID, matchID string, input tournaments.MatchResultInput) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if state != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrMatchResultConflict
	}
	var administers bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_administrators WHERE tournament_id = $1 AND account_id = $2)`, leagueID, account).Scan(&administers); err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() && !administers {
		return tournaments.Tournament{}, tournaments.ErrMatchResultForbidden
	}
	var format string
	if err := tx.QueryRow(ctx, `SELECT format FROM tournaments WHERE id=$1`, leagueID).Scan(&format); err != nil {
		return tournaments.Tournament{}, err
	}
	if format == "single_elimination" {
		if err := recordBracketResult(ctx, tx, account.String(), leagueID, matchID, input); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at=now() WHERE id=$1`, leagueID); err != nil {
			return tournaments.Tournament{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return tournaments.Tournament{}, err
		}
		return r.GetPublic(ctx, leagueID)
	}
	if input.HomePenalties != nil || input.AwayPenalties != nil {
		return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
	}
	var previousHome, previousAway *int
	if err := tx.QueryRow(ctx, `SELECT home_score, away_score FROM matches WHERE id = $1 AND tournament_id = $2 FOR UPDATE`, matchID, leagueID).Scan(&previousHome, &previousAway); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $3, away_score = $4 WHERE id = $1 AND tournament_id = $2`, matchID, leagueID, input.HomeScore, input.AwayScore); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score) VALUES ($1, $2, $3, $4, $5, $6)`, matchID, account, previousHome, previousAway, input.HomeScore, input.AwayScore); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// Complete closes a complete league and retains its co-champions in the same transaction.
func (r AccountTournamentRepository) Complete(ctx context.Context, accountID, leagueID string) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	var legs int
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state, round_robin_legs FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state, &legs); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrTournamentCompletionConflict
	}
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM matches WHERE tournament_id = $1 AND state NOT IN ('completed','bye'))`, leagueID).Scan(&pending); err != nil {
		return tournaments.Tournament{}, err
	}
	if pending {
		return tournaments.Tournament{}, tournaments.ErrTournamentCompletionConflict
	}
	var format string
	if err := tx.QueryRow(ctx, `SELECT format FROM tournaments WHERE id=$1`, leagueID).Scan(&format); err != nil {
		return tournaments.Tournament{}, err
	}
	if format == "single_elimination" {
		var winner string
		if err := tx.QueryRow(ctx, `SELECT winner_team_id::text FROM matches WHERE tournament_id=$1 ORDER BY round_number DESC LIMIT 1`, leagueID).Scan(&winner); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_champions(tournament_id,team_id) VALUES ($1,$2)`, leagueID, winner); err != nil {
			return tournaments.Tournament{}, err
		}
	} else {
		league, err := loadTournamentForCompletion(ctx, tx, leagueID, legs)
		if err != nil {
			return tournaments.Tournament{}, err
		}
		standings := tournaments.CalculateStandings(league)
		for _, standing := range standings {
			if standing.Position != 1 {
				continue
			}
			if _, err := tx.Exec(ctx, `INSERT INTO tournament_champions (tournament_id, team_id) VALUES ($1, $2)`, leagueID, standing.TeamID); err != nil {
				return tournaments.Tournament{}, err
			}
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state='completed' WHERE tournament_id=$1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET state = 'completed', last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

func loadTournamentForCompletion(ctx context.Context, tx pgx.Tx, leagueID string, legs int) (tournaments.Tournament, error) {
	league := tournaments.Tournament{RoundRobinLegs: legs, Teams: []tournaments.Team{}, Matches: []tournaments.Match{}}
	teams, err := tx.Query(ctx, `SELECT id::text, name, position FROM tournament_teams WHERE tournament_id = $1 ORDER BY position`, leagueID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer teams.Close()
	for teams.Next() {
		var team tournaments.Team
		if err := teams.Scan(&team.ID, &team.Name, &team.Position); err != nil {
			return tournaments.Tournament{}, err
		}
		league.Teams = append(league.Teams, team)
	}
	if err := teams.Err(); err != nil {
		return tournaments.Tournament{}, err
	}
	matches, err := tx.Query(ctx, `SELECT id::text, round_number, sequence, home_team_id::text, away_team_id::text, state, home_score, away_score FROM matches WHERE tournament_id = $1 ORDER BY round_number, sequence`, leagueID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer matches.Close()
	for matches.Next() {
		var match tournaments.Match
		if err := matches.Scan(&match.ID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore); err != nil {
			return tournaments.Tournament{}, err
		}
		league.Matches = append(league.Matches, match)
	}
	return league, matches.Err()
}

// Start freezes configuration and generates a full round for each leg.
func (r AccountTournamentRepository) Start(ctx context.Context, accountID, leagueID string, input tournaments.StartInput) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer string
	var state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "published" {
		return tournaments.Tournament{}, tournaments.ErrTournamentConflict
	}
	rows, err := tx.Query(ctx, `SELECT id::text FROM tournament_teams WHERE tournament_id = $1 ORDER BY position`, leagueID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return tournaments.Tournament{}, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if rows.Err() != nil {
		return tournaments.Tournament{}, rows.Err()
	}
	if len(ids) < 2 || len(ids) > 64 {
		return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
	}
	if input.Format == "" {
		input.Format = "league"
	}
	var stageID string
	var stageLegs *int
	if input.Format == "league" {
		stageLegs = &input.RoundRobinLegs
	}
	if err := tx.QueryRow(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state,round_robin_legs) VALUES ($1,1,$2,'in_progress',$3) ON CONFLICT (tournament_id,position) DO UPDATE SET type=EXCLUDED.type,state=EXCLUDED.state,round_robin_legs=EXCLUDED.round_robin_legs RETURNING id::text`, leagueID, input.Format, stageLegs).Scan(&stageID); err != nil {
		return tournaments.Tournament{}, err
	}
	if input.Format == "single_elimination" {
		if err := insertBracket(ctx, tx, leagueID, stageID, ids); err != nil {
			return tournaments.Tournament{}, err
		}
	} else {
		for _, fixture := range fixtures(ids, input.RoundRobinLegs) {
			if _, err := tx.Exec(ctx, `INSERT INTO matches (tournament_id, round_number, sequence, home_team_id, away_team_id,stage_id) VALUES ($1, $2, $3, $4, $5,$6)`, leagueID, fixture.round, fixture.sequence, fixture.home, fixture.away, stageID); err != nil {
				return tournaments.Tournament{}, err
			}
		}
	}
	legs := input.RoundRobinLegs
	if legs == 0 {
		legs = 1
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET state = 'in_progress', round_robin_legs = $2,format=$3, last_activity_at = now() WHERE id = $1`, leagueID, legs, input.Format); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// Cancel retains the league and its data but removes it from the active sporting lifecycle.
func (r AccountTournamentRepository) Cancel(ctx context.Context, accountID, leagueID string) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "published" && state != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrTournamentCancellationConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state='cancelled' WHERE tournament_id=$1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET state = 'cancelled', last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// AssignAdministrator directly assigns a verified account by its public username.
func (r AccountTournamentRepository) AssignAdministrator(ctx context.Context, accountID, leagueID, username string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	var administrator string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM accounts WHERE username = $1 AND state = 'verified'`, username).Scan(&administrator); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if administrator == organizer {
		return tournaments.ErrTournamentAdministratorConflict
	}
	result, err := tx.Exec(ctx, `INSERT INTO tournament_administrators (tournament_id, account_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, leagueID, administrator)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `INSERT INTO account_notifications (account_id, kind, tournament_id) VALUES ($1, 'tournament_administrator_assigned', $2)`, administrator, leagueID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListNotifications returns the account's internal inbox.
func (r AccountTournamentRepository) ListNotifications(ctx context.Context, accountID string) ([]notifications.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT account_notifications.id::text, account_notifications.kind, tournaments.id::text, tournaments.name, account_notifications.created_at, account_notifications.read_at FROM account_notifications JOIN tournaments ON tournaments.id = account_notifications.tournament_id WHERE account_notifications.account_id = $1 ORDER BY account_notifications.created_at DESC, account_notifications.id DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []notifications.Item{}
	for rows.Next() {
		var item notifications.Item
		var createdAt time.Time
		var readAt *time.Time
		if err := rows.Scan(&item.ID, &item.Kind, &item.TournamentID, &item.TournamentName, &createdAt, &readAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(time.RFC3339Nano)
		if readAt != nil {
			formattedReadAt := readAt.Format(time.RFC3339Nano)
			item.ReadAt = &formattedReadAt
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UnreadCount counts the account's unread notifications.
func (r AccountTournamentRepository) UnreadCount(ctx context.Context, accountID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM account_notifications WHERE account_id = $1 AND read_at IS NULL`, accountID).Scan(&count)
	return count, err
}

// MarkAllRead marks the account's pending notifications as read.
func (r AccountTournamentRepository) MarkAllRead(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE account_notifications SET read_at = now() WHERE account_id = $1 AND read_at IS NULL`, accountID)
	return err
}

// Delete removes a notification owned by the account.
func (r AccountTournamentRepository) Delete(ctx context.Context, accountID, notificationID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE id = $1 AND account_id = $2`, notificationID, accountID)
	return err
}

// DeleteAll removes all account notifications.
func (r AccountTournamentRepository) DeleteAll(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE account_id = $1`, accountID)
	return err
}

// ListAdministrators returns delegated administrators exclusively to the league owner.
func (r AccountTournamentRepository) ListAdministrators(ctx context.Context, accountID, leagueID string) ([]string, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return nil, err
	}
	var organizer string
	if err := r.pool.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return nil, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return nil, err
	}
	if organizer != account.String() {
		return nil, tournaments.ErrTournamentForbidden
	}
	rows, err := r.pool.Query(ctx, `SELECT accounts.username FROM tournament_administrators JOIN accounts ON accounts.id = tournament_administrators.account_id WHERE tournament_administrators.tournament_id = $1 ORDER BY accounts.username`, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	administrators := []string{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		administrators = append(administrators, username)
	}
	return administrators, rows.Err()
}

// RemoveAdministrator removes a delegated administrator exclusively for the league owner.
func (r AccountTournamentRepository) RemoveAdministrator(ctx context.Context, accountID, leagueID, username string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE tournament_id = $1 AND account_id = (SELECT id FROM accounts WHERE username = $2)`, leagueID, username); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// TransferOwnership atomically replaces the organizer while preserving the competition.
func (r AccountTournamentRepository) TransferOwnership(ctx context.Context, accountID, leagueID, username string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	var recipient string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM accounts WHERE username = $1 AND state = 'verified'`, username).Scan(&recipient); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if recipient == organizer {
		return tournaments.ErrTournamentOwnershipTransferConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET organizer_account_id = $2, last_activity_at = now() WHERE id = $1`, leagueID, recipient); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE tournament_id = $1 AND account_id = $2`, leagueID, recipient); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO account_notifications (account_id, kind, tournament_id) VALUES ($1, 'tournament_ownership_transferred', $2)`, recipient, leagueID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type fixture struct {
	round, sequence int
	home, away      string
}

func fixtures(ids []string, legs int) []fixture {
	players := append([]string(nil), ids...)
	if len(players)%2 != 0 {
		players = append(players, "")
	}
	var result []fixture
	half := len(players) / 2
	for leg := 0; leg < legs; leg++ {
		for round := 0; round < len(players)-1; round++ {
			for i := 0; i < half; i++ {
				home, away := players[i], players[len(players)-1-i]
				if home == "" || away == "" {
					continue
				}
				if leg == 1 {
					home, away = away, home
				}
				result = append(result, fixture{round: leg*(len(players)-1) + round + 1, sequence: i + 1, home: home, away: away})
			}
			players = append([]string{players[0], players[len(players)-1]}, players[1:len(players)-1]...)
		}
	}
	return result
}

// Authenticate resolves a valid opaque session to its account.
func (r AccountTournamentRepository) Authenticate(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	accountID, err := r.queries.FindAuthenticatedAccountID(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", tournaments.ErrUnauthenticated
		}
		return "", fmt.Errorf("buscar sesión: %w", err)
	}
	return uuidString(accountID), nil
}

// GetCurrentSession returns the identity and validity of the presented session.
func (r AccountTournamentRepository) GetCurrentSession(ctx context.Context, token string) (tournaments.CurrentSession, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	row, err := r.queries.GetCurrentSession(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tournaments.CurrentSession{}, tournaments.ErrUnauthenticated
		}
		return tournaments.CurrentSession{}, fmt.Errorf("consultar sesión actual: %w", err)
	}
	return tournaments.CurrentSession{
		AccountID:         uuidString(row.ID),
		Username:          row.Username,
		IdleExpiresAt:     timestamp(row.IdleExpiresAt.Time),
		AbsoluteExpiresAt: timestamp(row.AbsoluteExpiresAt.Time),
	}, nil
}

// RevokeSession idempotently revokes the presented session and its refresh tokens.
func (r AccountTournamentRepository) RevokeSession(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte("session:" + token))
	_, err := r.queries.RevokeSession(ctx, hash[:])
	return err
}

// GetAccessMethods returns the access methods configured for an account.
func (r AccountTournamentRepository) GetAccessMethods(ctx context.Context, accountID string) (tournaments.AccessMethods, error) {
	id, err := uuidValue(accountID)
	if err != nil {
		return tournaments.AccessMethods{}, err
	}
	row, err := r.queries.GetAccessMethods(ctx, id)
	if err != nil {
		return tournaments.AccessMethods{}, err
	}
	return tournaments.AccessMethods{Email: row.Email, Username: row.Username, HasPassword: row.HasPassword, HasGoogle: row.HasGoogle}, nil
}

// CurrentPasswordHash gets the verifier associated with an active session.
func (r AccountTournamentRepository) CurrentPasswordHash(ctx context.Context, sessionToken string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + sessionToken))
	return r.queries.GetCurrentPasswordHash(ctx, hash[:])
}

// CreateReauthenticationTicket stores a reauthentication ticket for a session.
func (r AccountTournamentRepository) CreateReauthenticationTicket(ctx context.Context, sessionToken string, ticketHash []byte) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	_, err := r.queries.CreateReauthenticationTicket(ctx, sqlc.CreateReauthenticationTicketParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	return err
}

// ConsumeReauthenticationTicketAndSetPassword consumes the ticket and changes the password.
func (r AccountTournamentRepository) ConsumeReauthenticationTicketAndSetPassword(ctx context.Context, sessionToken string, ticketHash []byte, passwordHash string) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	rows, err := r.queries.ConsumeReauthenticationTicketAndSetPassword(ctx, sqlc.ConsumeReauthenticationTicketAndSetPasswordParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash, PasswordHash: passwordHash})
	if err != nil {
		return err
	}
	if rows != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

// ConsumeReauthenticationTicketAndRemovePassword removes the local method only
// when Google remains available for the same account.
func (r AccountTournamentRepository) ConsumeReauthenticationTicketAndRemovePassword(ctx context.Context, sessionToken string, ticketHash []byte) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	rows, err := r.queries.ConsumeReauthenticationTicketAndRemovePassword(ctx, sqlc.ConsumeReauthenticationTicketAndRemovePasswordParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	if err != nil {
		return err
	}
	if rows != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

// List returns the requested page of league relationships.
func (r AccountTournamentRepository) List(ctx context.Context, accountID string, relationship tournaments.Relationship, cursor string, limit int) ([]tournaments.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return nil, fmt.Errorf("convertir cuenta: %w", err)
	}
	cursorUUID, err := optionalUUID(cursor)
	if err != nil {
		return nil, fmt.Errorf("convertir cursor: %w", err)
	}
	if limit < 1 || limit > 51 {
		return nil, fmt.Errorf("límite inválido")
	}
	pageSize := int32(limit)
	params := sqlc.ListAdministeredTournamentsParams{AccountID: accountUUID, CursorID: cursorUUID, PageSize: pageSize}
	if relationship == tournaments.Followed {
		rows, queryErr := r.queries.ListFollowedTournaments(ctx, sqlc.ListFollowedTournamentsParams(params))
		return followedItems(rows), queryErr
	}
	rows, err := r.queries.ListAdministeredTournaments(ctx, params)
	return administeredItems(rows), err
}

// ListRecent returns the fixed summary of relationships with recent activity.
func (r AccountTournamentRepository) ListRecent(ctx context.Context, accountID string) ([]tournaments.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return nil, fmt.Errorf("convertir cuenta: %w", err)
	}
	rows, err := r.queries.ListRecentAccountTournaments(ctx, accountUUID)
	return recentItems(rows), err
}

// Follow creates the follow relationship when the league is visible.
func (r AccountTournamentRepository) Follow(ctx context.Context, accountID, leagueID string) (bool, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return false, fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return false, fmt.Errorf("convertir liga: %w", err)
	}
	visible, err := r.queries.FollowVisibleTournament(ctx, sqlc.FollowVisibleTournamentParams{AccountID: accountUUID, TournamentID: leagueUUID})
	if err != nil {
		return false, fmt.Errorf("seguir liga: %w", err)
	}
	return visible, nil
}

// Unfollow idempotently removes the follow relationship.
func (r AccountTournamentRepository) Unfollow(ctx context.Context, accountID, leagueID string) error {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return fmt.Errorf("convertir liga: %w", err)
	}
	if err := r.queries.UnfollowTournament(ctx, sqlc.UnfollowTournamentParams{AccountID: accountUUID, TournamentID: leagueUUID}); err != nil {
		return fmt.Errorf("dejar de seguir liga: %w", err)
	}
	return nil
}

func optionalUUID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{}, nil
	}
	return uuidValue(value)
}

func uuidValue(value string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}
	return uuid, nil
}

func uuidString(value pgtype.UUID) string { return value.String() }

func timestamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func administeredItems(rows []sqlc.ListAdministeredTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func followedItems(rows []sqlc.ListFollowedTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func recentItems(rows []sqlc.ListRecentAccountTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

var _ access.Repository = AccountTournamentRepository{}
