package tournaments

import (
	"reflect"
	"testing"
)

func TestRepetitionStartConfigurationPreservesOnlyInitialRules(t *testing.T) {
	for _, test := range []struct {
		name   string
		source Tournament
		want   StartInput
	}{
		{"league", Tournament{Format: "league", RoundRobinLegs: 2}, StartInput{Format: "league", RoundRobinLegs: 2}},
		{"elimination", Tournament{Format: "single_elimination", RoundRobinLegs: 1}, StartInput{Format: "single_elimination"}},
		{"table", Tournament{Format: "league_then_single_elimination", RoundRobinLegs: 2, Stages: []Stage{{Type: "qualification_tiebreak"}, {Type: "league", LeagueStructure: "single_table", QualifierCount: 4}}}, StartInput{Format: "league_then_single_elimination", RoundRobinLegs: 2, LeagueStructure: "single_table", QualifierCount: 4}},
		{"groups", Tournament{Format: "league_then_single_elimination", RoundRobinLegs: 1, Stages: []Stage{{Type: "league", LeagueStructure: "groups", GroupCount: 4, QualifiersPerGroup: 2}}}, StartInput{Format: "league_then_single_elimination", RoundRobinLegs: 1, LeagueStructure: "groups", GroupCount: 4, QualifiersPerGroup: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := RepetitionStartConfiguration(test.source); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %#v want %#v", got, test.want)
			}
		})
	}
}
