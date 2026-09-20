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
	rows, err := tx.Query(ctx, `SELECT m.id::text,m.home_team_id::text,m.home_score,m.away_score FROM matches m JOIN tournament_stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 AND s.type='league' AND s.state='in_progress' AND (m.home_team_id=$2 OR m.away_team_id=$2) FOR UPDATE OF m`, leagueID, teamID)
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
			homeScore, awayScore = 0, winningScore
		} else {
			homeScore, awayScore = winningScore, 0
		}
		if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $2, away_score = $3, result_type='administrative' WHERE id = $1`, change.id, homeScore, awayScore); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score, result_type) VALUES ($1, $2, $3, $4, $5, $6, 'administrative')`, change.id, account, change.previousHome, change.previousAway, homeScore, awayScore); err != nil {
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
	if err := tournaments.ValidateLeagueResult(sport, input); err != nil {
		return tournaments.Tournament{}, tournaments.ErrInvalidTournamentInput
	}
	var previousHome, previousAway *int
	if err := tx.QueryRow(ctx, `SELECT home_score, away_score FROM matches WHERE id = $1 AND tournament_id = $2 FOR UPDATE`, matchID, leagueID).Scan(&previousHome, &previousAway); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.Tournament{}, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE matches SET state = 'completed', home_score = $3, away_score = $4, result_type='played' WHERE id = $1 AND tournament_id = $2`, matchID, leagueID, input.HomeScore, input.AwayScore); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO match_result_changes (match_id, changed_by_account_id, previous_home_score, previous_away_score, home_score, away_score, result_type) VALUES ($1, $2, $3, $4, $5, $6, 'played')`, matchID, account, previousHome, previousAway, input.HomeScore, input.AwayScore); err != nil {
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
	var sport tournaments.Sport
	var legs int
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text, state, round_robin_legs, sport FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer, &state, &legs, &sport); errors.Is(err, pgx.ErrNoRows) {
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
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_stages WHERE tournament_id=$1 AND position=2 AND type='single_elimination' AND state='in_progress')`, leagueID).Scan(&eliminationInProgress); err != nil {
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
		league, err := loadTournamentForCompletion(ctx, tx, leagueID, legs, sport)
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

func loadTournamentForCompletion(ctx context.Context, tx pgx.Tx, leagueID string, legs int, sport tournaments.Sport) (tournaments.Tournament, error) {
	league := tournaments.Tournament{RoundRobinLegs: legs, Sport: sport, Teams: []tournaments.Team{}, Matches: []tournaments.Match{}}
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
	matches, err := tx.Query(ctx, `SELECT id::text, round_number, sequence, home_team_id::text, away_team_id::text, state, home_score, away_score, result_type FROM matches WHERE tournament_id = $1 ORDER BY round_number, sequence`, leagueID)
	if err != nil {
		return tournaments.Tournament{}, err
	}
	defer matches.Close()
	for matches.Next() {
		var match tournaments.Match
		if err := matches.Scan(&match.ID, &match.RoundNumber, &match.Sequence, &match.HomeTeamID, &match.AwayTeamID, &match.State, &match.HomeScore, &match.AwayScore, &match.ResultType); err != nil {
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
	if input.Format == tournaments.FormatLeagueThenSingleElimination {
		if err := tournaments.ValidateMixedConfiguration(len(ids), input); err != nil {
			return tournaments.Tournament{}, tournaments.ErrInvalidMixedConfiguration
		}
		if err := tx.QueryRow(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state,round_robin_legs,league_structure,qualifier_count,group_count,qualifiers_per_group) VALUES ($1,1,'league','in_progress',$2,$3,NULLIF($4,0),NULLIF($5,0),NULLIF($6,0)) RETURNING id::text`, leagueID, input.RoundRobinLegs, input.LeagueStructure, input.QualifierCount, input.GroupCount, input.QualifiersPerGroup).Scan(&stageID); err != nil {
			return tournaments.Tournament{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_stages (tournament_id,position,type,state) VALUES ($1,2,'single_elimination','pending')`, leagueID); err != nil {
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

// StartElimination freezes the completed qualifying stage and creates its seeded bracket.
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
	if len(value.Stages) != 2 || value.Stages[0].State != "in_progress" || value.Stages[1].State != "pending" {
		return tournaments.Tournament{}, tournaments.ErrTournamentStageTransitionConflict
	}
	qualified, err := tournaments.QualifiedTeamIDs(value, value.Stages[0])
	if err != nil {
		return tournaments.Tournament{}, err
	}
	for index, teamID := range qualified {
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_stage_teams(tournament_id,stage_id,team_id,seed_position) VALUES ($1,$2,$3,$4)`, tournamentID, value.Stages[1].ID, teamID, index+1); err != nil {
			return tournaments.Tournament{}, err
		}
	}
	if err := insertBracket(ctx, tx, tournamentID, value.Stages[1].ID, qualified, true); err != nil {
		return tournaments.Tournament{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE tournament_stages SET state=CASE position WHEN 1 THEN 'completed' ELSE 'in_progress' END WHERE tournament_id=$1`, tournamentID); err != nil {
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
