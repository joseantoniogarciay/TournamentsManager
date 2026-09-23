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
	value := tournaments.Tournament{Teams: []tournaments.Team{}, Matches: []tournaments.Match{}, Stages: []tournaments.Stage{}, StageTeams: []tournaments.StageTeam{}, TieBreakPools: []tournaments.QualificationTieBreakPool{}, ChampionTeamIDs: []string{}}
	err := db.QueryRow(ctx, `SELECT id::text,name,sport,COALESCE(best_of_sets,0),format,state,round_robin_legs FROM tournaments WHERE id=$1`, id).Scan(&value.ID, &value.Name, &value.Sport, &value.BestOfSets, &value.Format, &value.State, &value.RoundRobinLegs)
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
	rows, err = db.Query(ctx, `SELECT id::text,position,type,state,round_robin_legs,COALESCE(league_structure,''),COALESCE(qualifier_count,0),COALESCE(group_count,0),COALESCE(qualifiers_per_group,0) FROM tournament_stages WHERE tournament_id=$1 ORDER BY position`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var stage tournaments.Stage
		if err = rows.Scan(&stage.ID, &stage.Position, &stage.Type, &stage.State, &stage.RoundRobinLegs, &stage.LeagueStructure, &stage.QualifierCount, &stage.GroupCount, &stage.QualifiersPerGroup); err != nil {
			rows.Close()
			return value, err
		}
		value.Stages = append(value.Stages, stage)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT stage_id::text,team_id::text,seed_position,COALESCE(group_number,0) FROM tournament_stage_teams WHERE tournament_id=$1 ORDER BY stage_id,seed_position`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var stageTeam tournaments.StageTeam
		if err = rows.Scan(&stageTeam.StageID, &stageTeam.TeamID, &stageTeam.SeedPosition, &stageTeam.GroupNumber); err != nil {
			rows.Close()
			return value, err
		}
		value.StageTeams = append(value.StageTeams, stageTeam)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT stage_id::text,pool_number,source_group_number,qualifier_count,current_cycle,state FROM tournament_tiebreak_pools WHERE tournament_id=$1 ORDER BY stage_id,pool_number`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var pool tournaments.QualificationTieBreakPool
		if err = rows.Scan(&pool.StageID, &pool.PoolNumber, &pool.SourceGroupNumber, &pool.QualifierCount, &pool.CurrentCycle, &pool.State); err != nil {
			rows.Close()
			return value, err
		}
		value.TieBreakPools = append(value.TieBreakPools, pool)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	rows, err = db.Query(ctx, `SELECT m.id::text,m.stage_id::text,m.round_number,m.sequence,COALESCE(m.group_number,0),COALESCE(m.home_team_id::text,''),COALESCE(m.away_team_id::text,''),m.state,m.home_score,m.away_score,m.home_source_kind,m.away_source_kind,COALESCE(m.home_source_match_id::text,''),COALESCE(m.away_source_match_id::text,''),COALESCE(m.winner_team_id::text,''),m.home_penalties,m.away_penalties,COALESCE(m.result_type,'') FROM matches m JOIN tournament_stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 ORDER BY s.position,m.round_number,m.sequence`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var match tournaments.Match
		match.Sets = []tournaments.SetScore{}
		if err = rows.Scan(&match.ID, &match.StageID, &match.RoundNumber, &match.Sequence, &match.GroupNumber, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore, &match.HomeSourceKind, &match.AwaySourceKind, &match.HomeSourceMatchID, &match.AwaySourceMatchID, &match.WinnerTeamID, &match.HomePenalties, &match.AwayPenalties, &match.ResultType); err != nil {
			rows.Close()
			return value, err
		}
		value.Matches = append(value.Matches, match)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	setsByMatch := make(map[string][]tournaments.SetScore)
	rows, err = db.Query(ctx, `SELECT ms.match_id::text,ms.home_score,ms.away_score FROM match_sets ms JOIN matches m ON m.id=ms.match_id WHERE m.tournament_id=$1 ORDER BY ms.match_id,ms.set_number`, id)
	if err != nil {
		return value, err
	}
	for rows.Next() {
		var matchID string
		var set tournaments.SetScore
		if err = rows.Scan(&matchID, &set.HomeScore, &set.AwayScore); err != nil {
			rows.Close()
			return value, err
		}
		setsByMatch[matchID] = append(setsByMatch[matchID], set)
	}
	rows.Close()
	if rows.Err() != nil {
		return value, rows.Err()
	}
	for index := range value.Matches {
		if sets, ok := setsByMatch[value.Matches[index].ID]; ok {
			value.Matches[index].Sets = sets
		}
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

func insertBracket(ctx context.Context, tx pgx.Tx, tournamentID, stageID string, teams []string, seeded bool) error {
	var bracket tournaments.Bracket
	var err error
	if seeded {
		bracket, err = tournaments.GenerateSeededSingleElimination(teams)
	} else {
		bracket, err = tournaments.GenerateSingleElimination(teams)
	}
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
	stageMatches := make([]tournaments.Match, 0)
	for _, match := range value.Matches {
		if match.StageID == target.StageID {
			stageMatches = append(stageMatches, match)
		}
	}
	bracket := tournaments.Bracket{Size: len(stageMatches) + 1, Sport: value.Sport, BestOfSets: value.BestOfSets}
	for _, match := range stageMatches {
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
			structural.Result = &tournaments.BracketResult{HomeScore: *match.HomeScore, AwayScore: *match.AwayScore, HomePenalties: match.HomePenalties, AwayPenalties: match.AwayPenalties, Sets: match.Sets}
		}
		bracket.Matches = append(bracket.Matches, structural)
	}
	if tournaments.RacketSport(value.Sport) {
		input, err = tournaments.NormalizeRacketResult(value.Sport, value.BestOfSets, input)
		if err != nil {
			return err
		}
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
		_, err = tx.Exec(ctx, `UPDATE matches SET home_team_id=NULLIF($2,'')::uuid,away_team_id=NULLIF($3,'')::uuid,winner_team_id=NULLIF($4,'')::uuid,state=$5,home_score=$6,away_score=$7,home_penalties=$8,away_penalties=$9,result_type=CASE WHEN $5='completed' THEN 'played' ELSE NULL END WHERE id=$1`, stageMatches[i].ID, match.HomeTeamID, match.AwayTeamID, match.WinnerTeamID, state, home, away, hp, ap)
		if err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, `DELETE FROM match_sets WHERE match_id=$1`, matchID); err != nil {
		return err
	}
	for index, set := range input.Sets {
		if _, err = tx.Exec(ctx, `INSERT INTO match_sets(match_id,set_number,home_score,away_score) VALUES ($1,$2,$3,$4)`, matchID, index+1, set.HomeScore, set.AwayScore); err != nil {
			return err
		}
	}
	var changeID string
	err = tx.QueryRow(ctx, `INSERT INTO match_result_changes(match_id,changed_by_account_id,previous_home_score,previous_away_score,home_score,away_score,previous_home_penalties,previous_away_penalties,home_penalties,away_penalties,result_type) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'played') RETURNING id::text`, matchID, accountID, target.HomeScore, target.AwayScore, input.HomeScore, input.AwayScore, target.HomePenalties, target.AwayPenalties, input.HomePenalties, input.AwayPenalties).Scan(&changeID)
	if err != nil {
		return err
	}
	for index, set := range input.Sets {
		if _, err = tx.Exec(ctx, `INSERT INTO match_result_change_sets(change_id,set_number,home_score,away_score) VALUES ($1,$2,$3,$4)`, changeID, index+1, set.HomeScore, set.AwayScore); err != nil {
			return err
		}
	}
	return nil
}
