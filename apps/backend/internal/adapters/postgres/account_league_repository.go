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
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
)

// ErrAccountHasOwnedLeagues indicates that the account still owns one or more leagues.
var ErrAccountHasOwnedLeagues = errors.New("account has owned leagues")

// AccountLeagueRepository persiste sesiones y relaciones de ligas.
type AccountLeagueRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewAccountLeagueRepository construye el adaptador de relaciones de cuenta.
func NewAccountLeagueRepository(pool *pgxpool.Pool) AccountLeagueRepository {
	return AccountLeagueRepository{pool: pool, queries: sqlc.New(pool)}
}

// PurgeExpired elimina un lote de cuentas cuya ventana de baja ha vencido.
// El CTE bloquea únicamente las cuentas seleccionadas y SKIP LOCKED evita esperar
// a una transacción concurrente sobre una de ellas.
func (r AccountLeagueRepository) PurgeExpired(ctx context.Context, limit int) (int64, error) {
	if limit < 1 {
		return 0, fmt.Errorf("límite de purga debe ser positivo")
	}
	command, err := r.pool.Exec(ctx, `
		WITH candidates AS (
			SELECT id
			FROM accounts
			WHERE state = 'deletion_pending'
				AND deletion_requested_at <= now() - interval '30 days'
				AND NOT EXISTS (
					SELECT 1
					FROM leagues
					WHERE organizer_account_id = accounts.id
				)
			ORDER BY deletion_requested_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		DELETE FROM accounts
		USING candidates
		WHERE accounts.id = candidates.id`, limit)
	if err != nil {
		return 0, err
	}
	return command.RowsAffected(), nil
}

var _ accounts.PurgeRepository = AccountLeagueRepository{}

// ScheduleAccountDeletion revoca accesos y relaciones personales sin borrar la cuenta.
func (r AccountLeagueRepository) ScheduleAccountDeletion(ctx context.Context, accountID string) (time.Time, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var hasOwned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM leagues WHERE organizer_account_id = $1)`, accountID).Scan(&hasOwned); err != nil {
		return time.Time{}, err
	}
	if hasOwned {
		return time.Time{}, ErrAccountHasOwnedLeagues
	}
	var requested time.Time
	if err := tx.QueryRow(ctx, `UPDATE accounts SET state = 'deletion_pending', deletion_requested_at = now() WHERE id = $1 AND state = 'verified' RETURNING deletion_requested_at`, accountID).Scan(&requested); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM league_followers WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM league_administrators WHERE account_id = $1`, accountID); err != nil {
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

// Create crea una liga publicada, todavía sin calendario.
func (r AccountLeagueRepository) Create(ctx context.Context, accountID string, input leagues.CreateInput) (leagues.League, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.League{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.League{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var league leagues.League
	if err := tx.QueryRow(ctx, `INSERT INTO leagues (organizer_account_id, name, state, published_at) VALUES ($1, $2, 'published', now()) RETURNING id::text, name, sport, format, state`, account, input.Name).Scan(&league.ID, &league.Name, &league.Sport, &league.Format, &league.State); err != nil {
		return leagues.League{}, err
	}
	league.Teams = make([]leagues.Team, len(input.Teams))
	for i, team := range input.Teams {
		if err := tx.QueryRow(ctx, `INSERT INTO league_teams (league_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, league.ID, team.Name, i+1).Scan(&league.Teams[i].ID, &league.Teams[i].Name, &league.Teams[i].Position); err != nil {
			return leagues.League{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.League{}, err
	}
	league.Matches = []leagues.Match{}
	return league, nil
}

// AddTeam añade un equipo a una liga publicada de su organizadora.
func (r AccountLeagueRepository) AddTeam(ctx context.Context, accountID, leagueID string, input leagues.TeamInput) (leagues.Team, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.Team{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.Team{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return leagues.Team{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.Team{}, err
	}
	if organizer != account.String() {
		return leagues.Team{}, leagues.ErrLeagueForbidden
	}
	if state != "published" {
		return leagues.Team{}, leagues.ErrLeagueTeamConflict
	}
	var position, count int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0), COUNT(*) FROM league_teams WHERE league_id = $1`, leagueID).Scan(&position, &count); err != nil {
		return leagues.Team{}, err
	}
	if count >= 64 {
		return leagues.Team{}, leagues.ErrLeagueTeamConflict
	}
	var team leagues.Team
	if err := tx.QueryRow(ctx, `INSERT INTO league_teams (league_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, leagueID, input.Name, position+1).Scan(&team.ID, &team.Name, &team.Position); err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return leagues.Team{}, leagues.ErrLeagueTeamConflict
		}
		return leagues.Team{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.Team{}, err
	}
	return team, nil
}

// RemoveTeam elimina un equipo de una liga publicada sin rebajarla de dos inscritos.
func (r AccountLeagueRepository) RemoveTeam(ctx context.Context, accountID, leagueID, teamID string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return leagues.ErrLeagueNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return leagues.ErrLeagueForbidden
	}
	if state != "published" {
		return leagues.ErrLeagueTeamConflict
	}
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM league_teams WHERE league_id = $1 AND id = $2)`, leagueID, teamID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return leagues.ErrLeagueNotFound
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM league_teams WHERE league_id = $1`, leagueID).Scan(&count); err != nil {
		return err
	}
	if count <= 2 {
		return leagues.ErrLeagueTeamConflict
	}
	command, err := tx.Exec(ctx, `DELETE FROM league_teams WHERE league_id = $1 AND id = $2`, leagueID, teamID)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return leagues.ErrLeagueNotFound
	}
	return tx.Commit(ctx)
}

// GetPublic devuelve la proyección visible de una liga ya creada.
func (r AccountLeagueRepository) GetPublic(ctx context.Context, leagueID string) (leagues.League, error) {
	league := leagues.League{Teams: []leagues.Team{}, Matches: []leagues.Match{}, ChampionTeamIDs: []string{}}
	if err := r.pool.QueryRow(ctx, `SELECT id::text, name, sport, format, state, round_robin_legs FROM leagues WHERE id = $1 AND state <> 'draft'`, leagueID).Scan(&league.ID, &league.Name, &league.Sport, &league.Format, &league.State, &league.RoundRobinLegs); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	teams, err := r.pool.Query(ctx, `SELECT id::text, name, position FROM league_teams WHERE league_id = $1 ORDER BY position`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer teams.Close()
	for teams.Next() {
		var team leagues.Team
		if err := teams.Scan(&team.ID, &team.Name, &team.Position); err != nil {
			return leagues.League{}, err
		}
		league.Teams = append(league.Teams, team)
	}
	if err := teams.Err(); err != nil {
		return leagues.League{}, err
	}
	champions, err := r.pool.Query(ctx, `SELECT team_id::text FROM league_champions WHERE league_id = $1 ORDER BY team_id`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer champions.Close()
	for champions.Next() {
		var teamID string
		if err := champions.Scan(&teamID); err != nil {
			return leagues.League{}, err
		}
		league.ChampionTeamIDs = append(league.ChampionTeamIDs, teamID)
	}
	if err := champions.Err(); err != nil {
		return leagues.League{}, err
	}
	matches, err := r.pool.Query(ctx, `SELECT id::text, round_number, sequence, home_team_id::text, away_team_id::text, state, home_score, away_score FROM matches WHERE league_id = $1 ORDER BY round_number, sequence`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer matches.Close()
	for matches.Next() {
		var match leagues.Match
		if err := matches.Scan(&match.ID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore); err != nil {
			return leagues.League{}, err
		}
		league.Matches = append(league.Matches, match)
	}
	return league, matches.Err()
}

// RecordResult guarda el marcador y una entrada de historial dentro de una única transacción.
func (r AccountLeagueRepository) RecordResult(ctx context.Context, accountID, leagueID, matchID string, input leagues.MatchResultInput) (leagues.League, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.League{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.League{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	if state != "in_progress" {
		return leagues.League{}, leagues.ErrMatchResultConflict
	}
	var administers bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM league_administrators WHERE league_id = $1 AND account_id = $2)`, leagueID, account).Scan(&administers); err != nil {
		return leagues.League{}, err
	}
	if organizer != account.String() && !administers {
		return leagues.League{}, leagues.ErrMatchResultForbidden
	}
	var previousHome, previousAway *int
	if err := tx.QueryRow(ctx, `SELECT home_score, away_score FROM matches WHERE id = $1 AND league_id = $2 FOR UPDATE`, matchID, leagueID).Scan(&previousHome, &previousAway); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $3, away_score = $4 WHERE id = $1 AND league_id = $2`, matchID, leagueID, input.HomeScore, input.AwayScore); err != nil {
		return leagues.League{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score) VALUES ($1, $2, $3, $4, $5, $6)`, matchID, account, previousHome, previousAway, input.HomeScore, input.AwayScore); err != nil {
		return leagues.League{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE leagues SET last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return leagues.League{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.League{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// Complete cierra una liga completa y conserva en la misma transacción sus co-campeones.
func (r AccountLeagueRepository) Complete(ctx context.Context, accountID, leagueID string) (leagues.League, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.League{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.League{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	var legs int
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state, round_robin_legs FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state, &legs); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	if organizer != account.String() {
		return leagues.League{}, leagues.ErrLeagueForbidden
	}
	if state != "in_progress" {
		return leagues.League{}, leagues.ErrLeagueCompletionConflict
	}
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM matches WHERE league_id = $1 AND state <> 'completed')`, leagueID).Scan(&pending); err != nil {
		return leagues.League{}, err
	}
	if pending {
		return leagues.League{}, leagues.ErrLeagueCompletionConflict
	}
	league, err := loadLeagueForCompletion(ctx, tx, leagueID, legs)
	if err != nil {
		return leagues.League{}, err
	}
	standings := leagues.CalculateStandings(league)
	for _, standing := range standings {
		if standing.Position != 1 {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO league_champions (league_id, team_id) VALUES ($1, $2)`, leagueID, standing.TeamID); err != nil {
			return leagues.League{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE leagues SET state = 'completed', last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return leagues.League{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.League{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

func loadLeagueForCompletion(ctx context.Context, tx pgx.Tx, leagueID string, legs int) (leagues.League, error) {
	league := leagues.League{RoundRobinLegs: legs, Teams: []leagues.Team{}, Matches: []leagues.Match{}}
	teams, err := tx.Query(ctx, `SELECT id::text, name, position FROM league_teams WHERE league_id = $1 ORDER BY position`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer teams.Close()
	for teams.Next() {
		var team leagues.Team
		if err := teams.Scan(&team.ID, &team.Name, &team.Position); err != nil {
			return leagues.League{}, err
		}
		league.Teams = append(league.Teams, team)
	}
	if err := teams.Err(); err != nil {
		return leagues.League{}, err
	}
	matches, err := tx.Query(ctx, `SELECT id::text, round_number, sequence, home_team_id::text, away_team_id::text, state, home_score, away_score FROM matches WHERE league_id = $1 ORDER BY round_number, sequence`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer matches.Close()
	for matches.Next() {
		var match leagues.Match
		if err := matches.Scan(&match.ID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore); err != nil {
			return leagues.League{}, err
		}
		league.Matches = append(league.Matches, match)
	}
	return league, matches.Err()
}

// Start congela la configuración y genera una vuelta completa por cada leg.
func (r AccountLeagueRepository) Start(ctx context.Context, accountID, leagueID string, input leagues.StartInput) (leagues.League, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.League{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.League{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer string
	var state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	if organizer != account.String() {
		return leagues.League{}, leagues.ErrLeagueForbidden
	}
	if state != "published" {
		return leagues.League{}, leagues.ErrLeagueConflict
	}
	rows, err := tx.Query(ctx, `SELECT id::text FROM league_teams WHERE league_id = $1 ORDER BY position`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return leagues.League{}, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	for _, fixture := range fixtures(ids, input.RoundRobinLegs) {
		if _, err := tx.Exec(ctx, `INSERT INTO matches (league_id, round_number, sequence, home_team_id, away_team_id) VALUES ($1, $2, $3, $4, $5)`, leagueID, fixture.round, fixture.sequence, fixture.home, fixture.away); err != nil {
			return leagues.League{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE leagues SET state = 'in_progress', round_robin_legs = $2, last_activity_at = now() WHERE id = $1`, leagueID, input.RoundRobinLegs); err != nil {
		return leagues.League{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.League{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// Cancel conserva la liga y sus datos, pero la saca del ciclo deportivo activo.
func (r AccountLeagueRepository) Cancel(ctx context.Context, accountID, leagueID string) (leagues.League, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return leagues.League{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return leagues.League{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
		return leagues.League{}, leagues.ErrLeagueNotFound
	} else if err != nil {
		return leagues.League{}, err
	}
	if organizer != account.String() {
		return leagues.League{}, leagues.ErrLeagueForbidden
	}
	if state != "published" && state != "in_progress" {
		return leagues.League{}, leagues.ErrLeagueCancellationConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE leagues SET state = 'cancelled', last_activity_at = now() WHERE id = $1`, leagueID); err != nil {
		return leagues.League{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return leagues.League{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// AssignAdministrator asigna directamente una cuenta verificada por su username público.
func (r AccountLeagueRepository) AssignAdministrator(ctx context.Context, accountID, leagueID, username string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return leagues.ErrLeagueNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return leagues.ErrLeagueForbidden
	}
	var administrator string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM accounts WHERE username = $1 AND state = 'verified'`, username).Scan(&administrator); errors.Is(err, pgx.ErrNoRows) {
		return leagues.ErrLeagueNotFound
	} else if err != nil {
		return err
	}
	if administrator == organizer {
		return leagues.ErrLeagueAdministratorConflict
	}
	result, err := tx.Exec(ctx, `INSERT INTO league_administrators (league_id, account_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, leagueID, administrator)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `INSERT INTO account_notifications (account_id, kind, league_id) VALUES ($1, 'league_administrator_assigned', $2)`, administrator, leagueID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListNotifications devuelve el buzón interno de la cuenta.
func (r AccountLeagueRepository) ListNotifications(ctx context.Context, accountID string) ([]notifications.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT account_notifications.id::text, account_notifications.kind, leagues.id::text, leagues.name, account_notifications.created_at::text, account_notifications.read_at::text FROM account_notifications JOIN leagues ON leagues.id = account_notifications.league_id WHERE account_notifications.account_id = $1 ORDER BY account_notifications.created_at DESC, account_notifications.id DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []notifications.Item{}
	for rows.Next() {
		var item notifications.Item
		if err := rows.Scan(&item.ID, &item.Kind, &item.LeagueID, &item.LeagueName, &item.CreatedAt, &item.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UnreadCount cuenta los avisos no leídos de la cuenta.
func (r AccountLeagueRepository) UnreadCount(ctx context.Context, accountID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM account_notifications WHERE account_id = $1 AND read_at IS NULL`, accountID).Scan(&count)
	return count, err
}

// MarkAllRead marca como leídos los avisos pendientes de la cuenta.
func (r AccountLeagueRepository) MarkAllRead(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE account_notifications SET read_at = now() WHERE account_id = $1 AND read_at IS NULL`, accountID)
	return err
}

// Delete elimina un aviso que pertenezca a la cuenta.
func (r AccountLeagueRepository) Delete(ctx context.Context, accountID, notificationID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE id = $1 AND account_id = $2`, notificationID, accountID)
	return err
}

// DeleteAll elimina todos los avisos de la cuenta.
func (r AccountLeagueRepository) DeleteAll(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE account_id = $1`, accountID)
	return err
}

// ListAdministrators devuelve las administradoras delegadas, exclusivamente a la organizadora.
func (r AccountLeagueRepository) ListAdministrators(ctx context.Context, accountID, leagueID string) ([]string, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return nil, err
	}
	var organizer string
	if err := r.pool.QueryRow(ctx, `SELECT organizer_account_id::text FROM leagues WHERE id = $1`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return nil, leagues.ErrLeagueNotFound
	} else if err != nil {
		return nil, err
	}
	if organizer != account.String() {
		return nil, leagues.ErrLeagueForbidden
	}
	rows, err := r.pool.Query(ctx, `SELECT accounts.username FROM league_administrators JOIN accounts ON accounts.id = league_administrators.account_id WHERE league_administrators.league_id = $1 ORDER BY accounts.username`, leagueID)
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

// RemoveAdministrator retira una administradora delegada exclusivamente a la organizadora.
func (r AccountLeagueRepository) RemoveAdministrator(ctx context.Context, accountID, leagueID, username string) error {
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return leagues.ErrLeagueNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return leagues.ErrLeagueForbidden
	}
	if _, err := tx.Exec(ctx, `DELETE FROM league_administrators WHERE league_id = $1 AND account_id = (SELECT id FROM accounts WHERE username = $2)`, leagueID, username); err != nil {
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

// Authenticate resuelve una sesión opaca válida en su cuenta.
func (r AccountLeagueRepository) Authenticate(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	accountID, err := r.queries.FindAuthenticatedAccountID(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", leagues.ErrUnauthenticated
		}
		return "", fmt.Errorf("buscar sesión: %w", err)
	}
	return uuidString(accountID), nil
}

// GetCurrentSession devuelve la identidad y vigencia de la sesión presentada.
func (r AccountLeagueRepository) GetCurrentSession(ctx context.Context, token string) (leagues.CurrentSession, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	row, err := r.queries.GetCurrentSession(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return leagues.CurrentSession{}, leagues.ErrUnauthenticated
		}
		return leagues.CurrentSession{}, fmt.Errorf("consultar sesión actual: %w", err)
	}
	return leagues.CurrentSession{
		AccountID:         uuidString(row.ID),
		Username:          row.Username,
		IdleExpiresAt:     timestamp(row.IdleExpiresAt.Time),
		AbsoluteExpiresAt: timestamp(row.AbsoluteExpiresAt.Time),
	}, nil
}

// RevokeSession revoca la sesión presentada y sus refresh tokens de forma idempotente.
func (r AccountLeagueRepository) RevokeSession(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte("session:" + token))
	_, err := r.queries.RevokeSession(ctx, hash[:])
	return err
}

// GetAccessMethods devuelve los métodos de acceso configurados para una cuenta.
func (r AccountLeagueRepository) GetAccessMethods(ctx context.Context, accountID string) (leagues.AccessMethods, error) {
	id, err := uuidValue(accountID)
	if err != nil {
		return leagues.AccessMethods{}, err
	}
	row, err := r.queries.GetAccessMethods(ctx, id)
	if err != nil {
		return leagues.AccessMethods{}, err
	}
	return leagues.AccessMethods{Email: row.Email, Username: row.Username, HasPassword: row.HasPassword, HasGoogle: row.HasGoogle}, nil
}

// CurrentPasswordHash obtiene el verificador asociado a una sesión activa.
func (r AccountLeagueRepository) CurrentPasswordHash(ctx context.Context, sessionToken string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + sessionToken))
	return r.queries.GetCurrentPasswordHash(ctx, hash[:])
}

// CreateReauthenticationTicket guarda un ticket de reautenticación para una sesión.
func (r AccountLeagueRepository) CreateReauthenticationTicket(ctx context.Context, sessionToken string, ticketHash []byte) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	_, err := r.queries.CreateReauthenticationTicket(ctx, sqlc.CreateReauthenticationTicketParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	return err
}

// ConsumeReauthenticationTicketAndSetPassword consume el ticket y cambia la contraseña.
func (r AccountLeagueRepository) ConsumeReauthenticationTicketAndSetPassword(ctx context.Context, sessionToken string, ticketHash []byte, passwordHash string) error {
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

// List devuelve la página solicitada de relaciones con ligas.
func (r AccountLeagueRepository) List(ctx context.Context, accountID string, relationship leagues.Relationship, cursor string, limit int) ([]leagues.Item, error) {
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
	params := sqlc.ListAdministeredLeaguesParams{AccountID: accountUUID, CursorID: cursorUUID, PageSize: pageSize}
	if relationship == leagues.Followed {
		rows, queryErr := r.queries.ListFollowedLeagues(ctx, sqlc.ListFollowedLeaguesParams(params))
		return followedItems(rows), queryErr
	}
	rows, err := r.queries.ListAdministeredLeagues(ctx, params)
	return administeredItems(rows), err
}

// ListRecent devuelve el resumen fijo de relaciones con actividad reciente.
func (r AccountLeagueRepository) ListRecent(ctx context.Context, accountID string) ([]leagues.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return nil, fmt.Errorf("convertir cuenta: %w", err)
	}
	rows, err := r.queries.ListRecentAccountLeagues(ctx, accountUUID)
	return recentItems(rows), err
}

// Follow crea la relación de seguimiento cuando la liga es visible.
func (r AccountLeagueRepository) Follow(ctx context.Context, accountID, leagueID string) (bool, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return false, fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return false, fmt.Errorf("convertir liga: %w", err)
	}
	visible, err := r.queries.FollowVisibleLeague(ctx, sqlc.FollowVisibleLeagueParams{AccountID: accountUUID, LeagueID: leagueUUID})
	if err != nil {
		return false, fmt.Errorf("seguir liga: %w", err)
	}
	return visible, nil
}

// Unfollow elimina la relación de seguimiento de forma idempotente.
func (r AccountLeagueRepository) Unfollow(ctx context.Context, accountID, leagueID string) error {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return fmt.Errorf("convertir liga: %w", err)
	}
	if err := r.queries.UnfollowLeague(ctx, sqlc.UnfollowLeagueParams{AccountID: accountUUID, LeagueID: leagueUUID}); err != nil {
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

func administeredItems(rows []sqlc.ListAdministeredLeaguesRow) []leagues.Item {
	items := make([]leagues.Item, len(rows))
	for index, row := range rows {
		items[index] = leagues.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func followedItems(rows []sqlc.ListFollowedLeaguesRow) []leagues.Item {
	items := make([]leagues.Item, len(rows))
	for index, row := range rows {
		items[index] = leagues.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func recentItems(rows []sqlc.ListRecentAccountLeaguesRow) []leagues.Item {
	items := make([]leagues.Item, len(rows))
	for index, row := range rows {
		items[index] = leagues.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

var _ access.Repository = AccountLeagueRepository{}
