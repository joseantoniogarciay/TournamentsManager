package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
)

// AccountLeagueRepository persiste sesiones y relaciones de ligas.
type AccountLeagueRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewAccountLeagueRepository construye el adaptador de relaciones de cuenta.
func NewAccountLeagueRepository(pool *pgxpool.Pool) AccountLeagueRepository {
	return AccountLeagueRepository{pool: pool, queries: sqlc.New(pool)}
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

// GetPublic devuelve la proyección visible de una liga ya creada.
func (r AccountLeagueRepository) GetPublic(ctx context.Context, leagueID string) (leagues.League, error) {
	var league leagues.League
	if err := r.pool.QueryRow(ctx, `SELECT id::text, name, sport, format, state FROM leagues WHERE id = $1 AND state <> 'draft'`, leagueID).Scan(&league.ID, &league.Name, &league.Sport, &league.Format, &league.State); errors.Is(err, pgx.ErrNoRows) {
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
	matches, err := r.pool.Query(ctx, `SELECT id::text, round_number, sequence, home_team_id::text, away_team_id::text, state FROM matches WHERE league_id = $1 ORDER BY round_number, sequence`, leagueID)
	if err != nil {
		return leagues.League{}, err
	}
	defer matches.Close()
	for matches.Next() {
		var match leagues.Match
		if err := matches.Scan(&match.ID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State); err != nil {
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
