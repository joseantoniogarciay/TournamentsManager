package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const fixtureInsertBatchSize = 256

type fixture struct {
	round, sequence int
	home, away      string
}

// insertFixtures reduces round trips while preserving the caller's transaction.
// Arrays have equal lengths and only contain generated fixture data.
func insertFixtures(ctx context.Context, tx pgx.Tx, tournamentID, stageID string, values []fixture, groupNumber int) error {
	for start := 0; start < len(values); start += fixtureInsertBatchSize {
		batch := values[start:min(start+fixtureInsertBatchSize, len(values))]
		rounds := make([]int64, len(batch))
		sequences := make([]int64, len(batch))
		homes := make([]string, len(batch))
		aways := make([]string, len(batch))
		for i, value := range batch {
			rounds[i], sequences[i] = int64(value.round), int64(value.sequence)
			homes[i], aways[i] = value.home, value.away
		}
		if _, err := tx.Exec(ctx, `INSERT INTO matches
			(tournament_id, stage_id, round_number, sequence, home_team_id, away_team_id, group_number)
			SELECT $1::uuid, $2::uuid, f.round_number, f.sequence, f.home::uuid, f.away::uuid, NULLIF($7::integer, 0)
			FROM unnest($3::integer[], $4::integer[], $5::text[], $6::text[])
			AS f(round_number, sequence, home, away)`, tournamentID, stageID, rounds, sequences, homes, aways, groupNumber); err != nil {
			return err
		}
	}
	return nil
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
