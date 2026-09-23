package postgres

import (
	"context"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// Create persists a published tournament and its initial teams atomically.
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
	if err := tx.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, name, sport, best_of_sets, state, published_at) VALUES ($1, $2, $3, NULLIF($4, 0), 'published', now()) RETURNING id::text, name, sport, COALESCE(best_of_sets, 0), format, state`, account, input.Name, input.Sport, input.BestOfSets).Scan(&league.ID, &league.Name, &league.Sport, &league.BestOfSets, &league.Format, &league.State); err != nil {
		return tournaments.Tournament{}, err
	}
	league.Teams = make([]tournaments.Team, len(input.Teams))
	for i, team := range input.Teams {
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text, name, position`, league.ID, team.Name, i+1).Scan(&league.Teams[i].ID, &league.Teams[i].Name, &league.Teams[i].Position); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_team_accounts (tournament_id, team_id, account_id) VALUES ($1, $2, $3)`, league.ID, league.Teams[0].ID, account); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET last_team_name = $2 WHERE id = $1`, account, league.Teams[0].Name); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	league.Matches = []tournaments.Match{}
	return league, nil
}

// SaveTeamInvitation replaces the single active invitation for an owner's unstarted tournament.
