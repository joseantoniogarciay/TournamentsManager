package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

type tournamentReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func readTournament(ctx context.Context, db tournamentReader, id string) (tournaments.Tournament, error) {
	value := tournaments.Tournament{Teams: []tournaments.Team{}, Matches: []tournaments.Match{}, Stages: []tournaments.Stage{}, ChampionTeamIDs: []string{}}
	err := db.QueryRow(ctx, `SELECT id::text,name,sport,format,state,round_robin_legs FROM tournaments WHERE id=$1`, id).Scan(&value.ID, &value.Name, &value.Sport, &value.Format, &value.State, &value.RoundRobinLegs)
	if errors.Is(err, pgx.ErrNoRows) {
		return value, tournaments.ErrTournamentNotFound
	}
	if err != nil {
		return value, err
	}
	rows, err := db.Query(ctx, `SELECT id::text,name,position,withdrawn_at IS NOT NULL FROM tournament_teams WHERE tournament_id=$1 ORDER BY position`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var team tournaments.Team
		if err = rows.Scan(&team.ID, &team.Name, &team.Position, &team.Withdrawn); err != nil {
			rows.Close()
			return value, err
		}
		value.Teams = append(value.Teams, team)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT id::text,position,type,state,round_robin_legs FROM tournament_stages WHERE tournament_id=$1 ORDER BY position`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var stage tournaments.Stage
		if err = rows.Scan(&stage.ID, &stage.Position, &stage.Type, &stage.State, &stage.RoundRobinLegs); err != nil {
			rows.Close()
			return value, err
		}
		value.Stages = append(value.Stages, stage)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT id::text,stage_id::text,round_number,sequence,COALESCE(home_team_id::text,''),COALESCE(away_team_id::text,''),state,home_score,away_score,home_source_kind,away_source_kind,COALESCE(home_source_match_id::text,''),COALESCE(away_source_match_id::text,''),COALESCE(winner_team_id::text,''),home_penalties,away_penalties FROM matches WHERE tournament_id=$1 ORDER BY stage_id,round_number,sequence`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var match tournaments.Match
		if err = rows.Scan(&match.ID, &match.StageID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore, &match.HomeSourceKind, &match.AwaySourceKind, &match.HomeSourceMatchID, &match.AwaySourceMatchID, &match.WinnerTeamID, &match.HomePenalties, &match.AwayPenalties); err != nil {
			rows.Close()
			return value, err
		}
		value.Matches = append(value.Matches, match)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT team_id::text FROM tournament_champions WHERE tournament_id=$1 ORDER BY team_id`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var team string
		if err = rows.Scan(&team); err != nil {
			rows.Close()
			return value, err
		}
		value.ChampionTeamIDs = append(value.ChampionTeamIDs, team)
	}
	rows.Close()
	return value, rows.Err()
}

func insertBracket(ctx context.Context, tx pgx.Tx, tournamentID, stageID string, teams []string) error {
	bracket, err := tournaments.GenerateSingleElimination(teams)
	if err != nil {
		return err
	}
	resolved, err := bracket.Resolve()
	if err != nil {
		return err
	}
	ids := map[[2]int]string{}
	for _, match := range resolved {
		state := "pending"
		if match.AutoWinnerTeamID != "" {
			state = "bye"
		}
		homeSource := ids[[2]int{match.Home.SourceRound, match.Home.SourceSequence}]
		awaySource := ids[[2]int{match.Away.SourceRound, match.Away.SourceSequence}]
		var id string
		err = tx.QueryRow(ctx, `INSERT INTO matches (tournament_id,stage_id,round_number,sequence,home_team_id,away_team_id,home_source_kind,away_source_kind,home_source_match_id,away_source_match_id,winner_team_id,state) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,NULLIF($6,'')::uuid,$7,$8,NULLIF($9,'')::uuid,NULLIF($10,'')::uuid,NULLIF($11,'')::uuid,$12) RETURNING id::text`, tournamentID, stageID, match.Round, match.Sequence, match.HomeTeamID, match.AwayTeamID, match.Home.Kind, match.Away.Kind, homeSource, awaySource, match.WinnerTeamID, state).Scan(&id)
		if err != nil {
			return err
		}
		ids[[2]int{match.Round, match.Sequence}] = id
	}
	return nil
}

func recordBracketResult(ctx context.Context, tx pgx.Tx, accountID, tournamentID, matchID string, input tournaments.MatchResultInput) error {
	value, err := readTournament(ctx, tx, tournamentID)
	if err != nil {
		return err
	}
	ids := map[string]tournaments.Match{}
	for _, match := range value.Matches {
		ids[match.ID] = match
	}
	target, exists := ids[matchID]
	if !exists {
		return tournaments.ErrTournamentNotFound
	}
	bracket := tournaments.Bracket{Size: len(value.Matches) + 1}
	for _, match := range value.Matches {
		source := func(kind tournaments.SlotSourceKind, team, sourceID string) tournaments.SlotSource {
			if kind == tournaments.SeededTeam {
				return tournaments.SlotSource{Kind: kind, TeamID: team}
			}
			previous := ids[sourceID]
			return tournaments.SlotSource{Kind: kind, SourceRound: previous.RoundNumber, SourceSequence: previous.Sequence}
		}
		structural := tournaments.BracketMatch{Round: match.RoundNumber, Sequence: match.Sequence, Home: source(match.HomeSourceKind, match.HomeTeamID, match.HomeSourceMatchID), Away: source(match.AwaySourceKind, match.AwayTeamID, match.AwaySourceMatchID)}
		if match.State == "bye" {
			structural.AutoWinnerTeamID = match.WinnerTeamID
		}
		if match.State == "completed" {
			if match.HomeScore == nil || match.AwayScore == nil {
				return tournaments.ErrInvalidBracketStructure
			}
			structural.Result = &tournaments.BracketResult{HomeScore: *match.HomeScore, AwayScore: *match.AwayScore, HomePenalties: match.HomePenalties, AwayPenalties: match.AwayPenalties}
		}
		bracket.Matches = append(bracket.Matches, structural)
	}
	next, err := bracket.RecordResult(target.RoundNumber, target.Sequence, tournaments.BracketResult(input))
	if err != nil {
		return err
	}
	resolved, err := next.Resolve()
	if err != nil {
		return err
	}
	for i, match := range resolved {
		state := "pending"
		var home, away, hp, ap *int
		if match.AutoWinnerTeamID != "" {
			state = "bye"
		}
		if match.Result != nil {
			state = "completed"
			home = &match.Result.HomeScore
			away = &match.Result.AwayScore
			hp = match.Result.HomePenalties
			ap = match.Result.AwayPenalties
		}
		_, err = tx.Exec(ctx, `UPDATE matches SET home_team_id=NULLIF($2,'')::uuid,away_team_id=NULLIF($3,'')::uuid,winner_team_id=NULLIF($4,'')::uuid,state=$5,home_score=$6,away_score=$7,home_penalties=$8,away_penalties=$9 WHERE id=$1`, value.Matches[i].ID, match.HomeTeamID, match.AwayTeamID, match.WinnerTeamID, state, home, away, hp, ap)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO match_result_changes(match_id,changed_by_account_id,previous_home_score,previous_away_score,home_score,away_score,previous_home_penalties,previous_away_penalties,home_penalties,away_penalties) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, matchID, accountID, target.HomeScore, target.AwayScore, input.HomeScore, input.AwayScore, target.HomePenalties, target.AwayPenalties, input.HomePenalties, input.AwayPenalties)
	return err
}
