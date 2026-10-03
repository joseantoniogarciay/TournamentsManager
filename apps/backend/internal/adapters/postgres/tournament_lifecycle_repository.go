package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// GetPublic returns the public projection of a tournament.
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

// WithdrawTeam keeps the roster for historical standings and applies the domain's administrative score.
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
	var sport tournaments.Sport
	var activeLeague bool
	if err := tx.QueryRow(ctx, `SELECT t.sport,EXISTS(SELECT 1 FROM tournament_stages s WHERE s.tournament_id=t.id AND s.type='league' AND s.state='in_progress') FROM tournaments t WHERE t.id=$1`, leagueID).Scan(&sport, &activeLeague); err != nil {
		return tournaments.Tournament{}, err
	}
	if !activeLeague {
		return tournaments.Tournament{}, tournaments.ErrTournamentWithdrawalConflict
	}
	winningScore, err := tournaments.AdministrativeWinningScore(sport)
	if err != nil {
		return tournaments.Tournament{}, err
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
	rows, err := tx.Query(ctx, `SELECT m.id::text,m.home_team_id::text,m.home_score,m.away_score,m.result_type,m.incident FROM matches m JOIN tournament_stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 AND s.type='league' AND s.state='in_progress' AND (m.home_team_id=$2 OR m.away_team_id=$2) FOR UPDATE OF m`, leagueID, teamID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer rows.Close()
	type scoreChange struct {
		id                         string
		homeID                     string
		previousHome, previousAway *int
		previousType               *string
		previousIncident           *tournaments.MatchIncident
	}
	changes := []scoreChange{}
	for rows.Next() {
		var change scoreChange
		if err := rows.Scan(&change.id, &change.homeID, &change.previousHome, &change.previousAway, &change.previousType, &change.previousIncident); err != nil {
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
			homeScore, awayScore = 0, winningScore
		} else {
			homeScore, awayScore = winningScore, 0
		}
		if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $2, away_score = $3, result_type='administrative',incident=NULL,home_penalties=NULL,away_penalties=NULL WHERE id = $1`, change.id, homeScore, awayScore); err != nil {
			return tournaments.Tournament{}, err
		}
		var changeID string
		if err := tx.QueryRow(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score, result_type,previous_result_type,previous_incident) VALUES ($1, $2, $3, $4, $5, $6, 'administrative',$7,$8) RETURNING id::text`, change.id, account, change.previousHome, change.previousAway, homeScore, awayScore, change.previousType, change.previousIncident).Scan(&changeID); err != nil {
			return tournaments.Tournament{}, err
		}
		if sport == tournaments.SportVolleyball {
			if err := replaceMatchSets(ctx, tx, change.id, changeID, tournaments.AdministrativeVolleyballSets(homeScore > awayScore)); err != nil {
				return tournaments.Tournament{}, err
			}
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
	var sport tournaments.Sport
	var stageType, stageState string
	if err := tx.QueryRow(ctx, `SELECT s.type,s.state,t.sport FROM matches m JOIN tournament_stages s ON s.id=m.stage_id JOIN tournaments t ON t.id=m.tournament_id WHERE m.id=$1 AND m.tournament_id=$2`, matchID, leagueID).Scan(&stageType, &stageState, &sport); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if stageState != "in_progress" {
		return tournaments.Tournament{}, tournaments.ErrMatchResultConflict
	}
	if stageType == tournaments.StageTypeQualificationTieBreak {
		if err := recordQualificationTieBreakResult(ctx, tx, account.String(), leagueID, matchID, sport, input); err != nil {
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
	if stageType == "single_elimination" {
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
	if input.Incident != nil {
		best := 0
		if sport == tournaments.SportVolleyball {
			best = 5
		}
		input, err = tournaments.NormalizeIncidentResult(sport, best, 0, input)
		if err != nil {
			return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
		}
	} else if sport == tournaments.SportVolleyball {
		input, err = tournaments.NormalizeSetResult(sport, 5, 0, input)
		if err != nil {
			return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
		}
	}
	if err := tournaments.ValidateLeagueResult(sport, input); err != nil {
		return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
	}
	var previousHome, previousAway *int
	var administrative bool
	var previousType tournaments.ResultType
	var previousIncident *tournaments.MatchIncident
	if err := tx.QueryRow(ctx, `SELECT home_score,away_score,COALESCE(result_type='administrative',false),COALESCE(result_type,''),incident FROM matches WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, matchID, leagueID).Scan(&previousHome, &previousAway, &administrative, &previousType, &previousIncident); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if sport == tournaments.SportVolleyball && administrative {
		return tournaments.Tournament{}, tournaments.ErrMatchResultConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $3, away_score = $4, result_type=$5,incident=$6 WHERE id = $1 AND tournament_id = $2`, matchID, leagueID, input.HomeScore, input.AwayScore, tournaments.MatchResultType(input), input.Incident); err != nil {
		return tournaments.Tournament{}, err
	}
	changeID, err := recordResultChange(ctx, tx, account.String(), matchID, previousHome, previousAway, nil, nil, previousType, previousIncident, input)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	if tournaments.SetSport(sport) {
		if err := replaceMatchSets(ctx, tx, matchID, changeID, input.Sets); err != nil {
			return tournaments.Tournament{}, err
		}
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
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state); errors.Is(err, pgx.ErrNoRows) {
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
	if format == tournaments.FormatLeagueThenSingleElimination {
		var eliminationInProgress bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_stages WHERE tournament_id=$1 AND type='single_elimination' AND state='in_progress')`, leagueID).Scan(&eliminationInProgress); err != nil {
			return tournaments.Tournament{}, err
		}
		if !eliminationInProgress {
			return tournaments.Tournament{}, tournaments.ErrTournamentCompletionConflict
		}
	}
	if format != "league" {
		var winner string
		if err := tx.QueryRow(ctx, `SELECT m.winner_team_id::text FROM matches m JOIN tournament_stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 AND s.type='single_elimination' ORDER BY s.position DESC,m.round_number DESC LIMIT 1`, leagueID).Scan(&winner); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_champions(tournament_id,team_id) VALUES ($1,$2)`, leagueID, winner); err != nil {
			return tournaments.Tournament{}, err
		}
	} else {
		league, err := readTournament(ctx, tx, leagueID)
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
	var sport tournaments.Sport
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state, sport FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state, &sport); errors.Is(err, pgx.ErrNoRows) {
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
	if tournaments.RacketSport(sport) && input.Format != "single_elimination" {
		return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
	}
	var stageID string
	if input.Format == tournaments.FormatLeagueThenSingleElimination {
		if err := tournaments.ValidateMixedConfiguration(len(ids), input); err != nil {
			return tournaments.Tournament{}, tournaments.ErrInvalidMixedConfiguration
		}
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state,round_robin_legs,league_structure,qualifier_count,group_count,qualifiers_per_group) VALUES ($1,1,'league','in_progress',$2,$3,NULLIF($4,0),NULLIF($5,0),NULLIF($6,0)) RETURNING id::text`, leagueID, input.RoundRobinLegs, input.LeagueStructure, input.QualifierCount, input.GroupCount, input.QualifiersPerGroup).Scan(&stageID); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state) VALUES ($1,2,'qualification_tiebreak','pending'),($1,3,'single_elimination','pending')`, leagueID); err != nil {
			return tournaments.Tournament{}, err
		}
		assignments := make([]tournaments.StageTeam, len(ids))
		if input.LeagueStructure == tournaments.LeagueStructureGroups {
			assignments, err = tournaments.AssignGroups(ids, input.GroupCount)
			if err != nil {
				return tournaments.Tournament{}, err
			}
		} else {
			for index, teamID := range ids {
				assignments[index] = tournaments.StageTeam{TeamID: teamID, SeedPosition: index + 1}
			}
		}
		for _, assignment := range assignments {
			if _, err := tx.Exec(ctx, `INSERT INTO tournament_stage_teams(tournament_id,stage_id,team_id,seed_position,group_number) VALUES ($1,$2,$3,$4,NULLIF($5,0))`, leagueID, stageID, assignment.TeamID, assignment.SeedPosition, assignment.GroupNumber); err != nil {
				return tournaments.Tournament{}, err
			}
		}
		if input.LeagueStructure == tournaments.LeagueStructureGroups {
			for group := 1; group <= input.GroupCount; group++ {
				groupTeams := []string{}
				for _, assignment := range assignments {
					if assignment.GroupNumber == group {
						groupTeams = append(groupTeams, assignment.TeamID)
					}
				}
				if err := insertLeagueFixtures(ctx, tx, leagueID, stageID, groupTeams, input.RoundRobinLegs, group); err != nil {
					return tournaments.Tournament{}, err
				}
			}
		} else if err := insertLeagueFixtures(ctx, tx, leagueID, stageID, ids, input.RoundRobinLegs, 0); err != nil {
			return tournaments.Tournament{}, err
		}
	} else {
		var stageLegs *int
		if input.Format == "league" {
			stageLegs = &input.RoundRobinLegs
		}
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state,round_robin_legs) VALUES ($1,1,$2,'in_progress',$3) ON CONFLICT (tournament_id,position) DO UPDATE SET type=EXCLUDED.type,state=EXCLUDED.state,round_robin_legs=EXCLUDED.round_robin_legs RETURNING id::text`, leagueID, input.Format, stageLegs).Scan(&stageID); err != nil {
			return tournaments.Tournament{}, err
		}
		if input.Format == "single_elimination" {
			if err := insertBracket(ctx, tx, leagueID, stageID, ids, false); err != nil {
				return tournaments.Tournament{}, err
			}
		} else if err := insertLeagueFixtures(ctx, tx, leagueID, stageID, ids, input.RoundRobinLegs, 0); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	legs := input.RoundRobinLegs
	if legs == 0 {
		legs = 1
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET state = 'in_progress', round_robin_legs = $2,format=$3, last_activity_at = now() WHERE id = $1`, leagueID, legs, input.Format); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_team_invitations WHERE tournament_id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

func insertLeagueFixtures(ctx context.Context, tx pgx.Tx, tournamentID, stageID string, teamIDs []string, legs, groupNumber int) error {
	sequenceOffset := 0
	if groupNumber > 0 {
		sequenceOffset = (groupNumber - 1) * ((len(teamIDs) + 1) / 2)
	}
	for _, fixture := range fixtures(teamIDs, legs) {
		if _, err := tx.Exec(ctx, `INSERT INTO matches (tournament_id,round_number,sequence,home_team_id,away_team_id,stage_id,group_number) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,0))`, tournamentID, fixture.round, fixture.sequence+sequenceOffset, fixture.home, fixture.away, stageID, groupNumber); err != nil {
			return err
		}
	}
	return nil
}

func insertQualificationTieBreakCycle(ctx context.Context, tx pgx.Tx, tournamentID, stageID string, poolNumber, cycle int, teamIDs []string) error {
	if len(teamIDs) < 2 {
		return tournaments.ErrTournamentStageTransitionConflict
	}
	var sequence int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(sequence),0) FROM matches WHERE stage_id=$1 AND round_number=$2`, stageID, cycle).Scan(&sequence); err != nil {
		return err
	}
	for home := 0; home < len(teamIDs); home++ {
		for away := home + 1; away < len(teamIDs); away++ {
			sequence++
			if _, err := tx.Exec(ctx, `INSERT INTO matches(tournament_id,stage_id,round_number,sequence,group_number,home_team_id,away_team_id) VALUES ($1,$2,$3,$4,$5,$6,$7)`, tournamentID, stageID, cycle, sequence, poolNumber, teamIDs[home], teamIDs[away]); err != nil {
				return err
			}
		}
	}
	return nil
}

func recordQualificationTieBreakResult(ctx context.Context, tx pgx.Tx, accountID, tournamentID, matchID string, sport tournaments.Sport, input tournaments.MatchResultInput) error {
	var stageID, homeTeamID, awayTeamID string
	var poolNumber, cycle, currentCycle int
	var poolState string
	var previousHome, previousAway, previousHomePenalties, previousAwayPenalties *int
	var previousType tournaments.ResultType
	var previousIncident *tournaments.MatchIncident
	err := tx.QueryRow(ctx, `SELECT m.stage_id::text,m.group_number,m.round_number,m.home_team_id::text,m.away_team_id::text,m.home_score,m.away_score,m.home_penalties,m.away_penalties,p.current_cycle,p.state,COALESCE(m.result_type,''),m.incident FROM matches m JOIN tournament_tiebreak_pools p ON p.stage_id=m.stage_id AND p.pool_number=m.group_number WHERE m.id=$1 AND m.tournament_id=$2 FOR UPDATE OF m,p`, matchID, tournamentID).Scan(&stageID, &poolNumber, &cycle, &homeTeamID, &awayTeamID, &previousHome, &previousAway, &previousHomePenalties, &previousAwayPenalties, &currentCycle, &poolState, &previousType, &previousIncident)
	if errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	}
	if err != nil {
		return err
	}
	if poolState != "in_progress" || cycle != currentCycle {
		return tournaments.ErrMatchResultConflict
	}
	bestOfSets := 0
	if sport == tournaments.SportVolleyball {
		bestOfSets = 5
	}
	if input.Incident != nil {
		input, err = tournaments.NormalizeIncidentResult(sport, bestOfSets, 0, input)
		if err != nil {
			return tournaments.ErrInvalidTournamentInput
		}
	} else if sport == tournaments.SportVolleyball {
		input, err = tournaments.NormalizeSetResult(sport, bestOfSets, 0, input)
		if err != nil {
			return tournaments.ErrInvalidTournamentInput
		}
	}
	winnerTeamID, err := tournaments.DecisiveWinnerTeamID(sport, bestOfSets, 0, homeTeamID, awayTeamID, input)
	if err != nil {
		return tournaments.ErrInvalidTournamentInput
	}
	if _, err := tx.Exec(ctx, `UPDATE matches SET state='completed',home_score=$2,away_score=$3,home_penalties=$4,away_penalties=$5,winner_team_id=$6,result_type=$7,incident=$8 WHERE id=$1`, matchID, input.HomeScore, input.AwayScore, input.HomePenalties, input.AwayPenalties, winnerTeamID, tournaments.MatchResultType(input), input.Incident); err != nil {
		return err
	}
	changeID, err := recordResultChange(ctx, tx, accountID, matchID, previousHome, previousAway, previousHomePenalties, previousAwayPenalties, previousType, previousIncident, input)
	if err != nil {
		return err
	}
	if tournaments.SetSport(sport) {
		if err := replaceMatchSets(ctx, tx, matchID, changeID, input.Sets); err != nil {
			return err
		}
	}
	value, err := readTournament(ctx, tx, tournamentID)
	if err != nil {
		return err
	}
	var pool tournaments.QualificationTieBreakPool
	found := false
	for _, candidate := range value.TieBreakPools {
		if candidate.StageID == stageID && candidate.PoolNumber == poolNumber {
			pool, found = candidate, true
			break
		}
	}
	if !found {
		return tournaments.ErrTournamentStageTransitionConflict
	}
	resolution, err := tournaments.ResolveQualificationTieBreak(value, pool)
	if err != nil {
		return err
	}
	if resolution.NeedsNextCycle {
		nextCycle := pool.CurrentCycle + 1
		if _, err := tx.Exec(ctx, `UPDATE tournament_tiebreak_pools SET current_cycle=$3 WHERE stage_id=$1 AND pool_number=$2 AND state='in_progress'`, stageID, poolNumber, nextCycle); err != nil {
			return err
		}
		return insertQualificationTieBreakCycle(ctx, tx, tournamentID, stageID, poolNumber, nextCycle, resolution.PendingTeamIDs)
	}
	if !resolution.Complete {
		return nil
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_tiebreak_pools SET state='completed' WHERE stage_id=$1 AND pool_number=$2`, stageID, poolNumber); err != nil {
		return err
	}
	var pendingPools bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tournament_tiebreak_pools WHERE stage_id=$1 AND state='in_progress')`, stageID).Scan(&pendingPools); err != nil {
		return err
	}
	if !pendingPools {
		if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state='completed' WHERE id=$1`, stageID); err != nil {
			return err
		}
	}
	return nil
}

// StartElimination advances a mixed tournament through any required tiebreak and into its bracket.
func (r AccountTournamentRepository) StartElimination(ctx context.Context, accountID, tournamentID string) (tournaments.Tournament, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer, state, format string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text,state,format FROM tournaments WHERE id=$1 FOR UPDATE`, tournamentID).Scan(&organizer, &state, &format); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if organizer != account.String() {
		return tournaments.Tournament{}, tournaments.ErrTournamentForbidden
	}
	if state != "in_progress" || format != tournaments.FormatLeagueThenSingleElimination {
		return tournaments.Tournament{}, tournaments.ErrTournamentStageTransitionConflict
	}
	value, err := readTournament(ctx, tx, tournamentID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	var leagueStage, tieBreakStage, eliminationStage *tournaments.Stage
	for index := range value.Stages {
		switch value.Stages[index].Type {
		case "league":
			leagueStage = &value.Stages[index]
		case tournaments.StageTypeQualificationTieBreak:
			tieBreakStage = &value.Stages[index]
		case "single_elimination":
			eliminationStage = &value.Stages[index]
		}
	}
	if leagueStage == nil || tieBreakStage == nil || eliminationStage == nil || eliminationStage.State != "pending" {
		return tournaments.Tournament{}, tournaments.ErrTournamentStageTransitionConflict
	}
	var qualified []string
	switch {
	case leagueStage.State == "in_progress" && tieBreakStage.State == "pending":
		plan, err := tournaments.PlanQualification(value, *leagueStage)
		if err != nil {
			return tournaments.Tournament{}, err
		}
		if len(plan.Pools) > 0 {
			seedPosition := 0
			for _, pool := range plan.Pools {
				if _, err := tx.Exec(ctx, `INSERT INTO tournament_tiebreak_pools(tournament_id,stage_id,pool_number,source_group_number,qualifier_count) VALUES ($1,$2,$3,$4,$5)`, tournamentID, tieBreakStage.ID, pool.PoolNumber, pool.SourceGroupNumber, pool.QualifierCount); err != nil {
					return tournaments.Tournament{}, err
				}
				teamIDs := make([]string, len(pool.Standings))
				for index, standing := range pool.Standings {
					teamIDs[index] = standing.TeamID
				}
				for _, teamID := range teamIDs {
					seedPosition++
					if _, err := tx.Exec(ctx, `INSERT INTO tournament_stage_teams(tournament_id,stage_id,team_id,seed_position,group_number) VALUES ($1,$2,$3,$4,$5)`, tournamentID, tieBreakStage.ID, teamID, seedPosition, pool.PoolNumber); err != nil {
						return tournaments.Tournament{}, err
					}
				}
				if err := insertQualificationTieBreakCycle(ctx, tx, tournamentID, tieBreakStage.ID, pool.PoolNumber, 1, teamIDs); err != nil {
					return tournaments.Tournament{}, err
				}
			}
			if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state=CASE id WHEN $2 THEN 'completed' WHEN $3 THEN 'in_progress' ELSE state END WHERE tournament_id=$1`, tournamentID, leagueStage.ID, tieBreakStage.ID); err != nil {
				return tournaments.Tournament{}, err
			}
			if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at=now() WHERE id=$1`, tournamentID); err != nil {
				return tournaments.Tournament{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return tournaments.Tournament{}, err
			}
			return r.GetPublic(ctx, tournamentID)
		}
		qualified, err = tournaments.QualifiedTeamIDs(value, *leagueStage)
		if err != nil {
			return tournaments.Tournament{}, err
		}
	case leagueStage.State == "completed" && tieBreakStage.State == "completed":
		qualified, err = tournaments.ResolvedQualifiedTeamIDs(value, *leagueStage, *tieBreakStage)
		if err != nil {
			return tournaments.Tournament{}, err
		}
	default:
		return tournaments.Tournament{}, tournaments.ErrTournamentStageTransitionConflict
	}
	for index, teamID := range qualified {
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_stage_teams(tournament_id,stage_id,team_id,seed_position) VALUES ($1,$2,$3,$4)`, tournamentID, eliminationStage.ID, teamID, index+1); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	if err := insertBracket(ctx, tx, tournamentID, eliminationStage.ID, qualified, true); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state=CASE WHEN id=$2 THEN 'in_progress' WHEN type IN ('league','qualification_tiebreak') THEN 'completed' ELSE state END WHERE tournament_id=$1`, tournamentID, eliminationStage.ID); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET last_activity_at=now() WHERE id=$1`, tournamentID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, tournamentID)
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
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_team_invitations WHERE tournament_id = $1`, leagueID); err != nil {
		return tournaments.Tournament{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tournaments.Tournament{}, err
	}
	return r.GetPublic(ctx, leagueID)
}

// AssignAdministrator directly assigns a verified account by its public username.
