package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// recordResultChange retains incident transitions alongside the existing score history.
func recordResultChange(ctx context.Context, tx pgx.Tx, accountID, matchID string, previousHome, previousAway, previousHP, previousAP *int, previousType tournaments.ResultType, previousIncident *tournaments.MatchIncident, input tournaments.MatchResultInput) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO match_result_changes(match_id,changed_by_account_id,previous_home_score,previous_away_score,home_score,away_score,previous_home_penalties,previous_away_penalties,home_penalties,away_penalties,result_type,incident,previous_result_type,previous_incident) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NULLIF($13,''),$14) RETURNING id::text`, matchID, accountID, previousHome, previousAway, input.HomeScore, input.AwayScore, previousHP, previousAP, input.HomePenalties, input.AwayPenalties, tournaments.MatchResultType(input), input.Incident, previousType, previousIncident).Scan(&id)
	return id, err
}
