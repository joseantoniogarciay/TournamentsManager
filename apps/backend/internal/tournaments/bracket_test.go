package tournaments

import "testing"

func TestGenerateSingleEliminationMaterializesFutureWinnerSlots(t *testing.T) {
	bracket, err := GenerateSingleElimination([]string{"a", "b", "c", "d"})
	if err != nil {
		t.Fatal(err)
	}
	if bracket.Size != 4 || len(bracket.Matches) != 3 {
		t.Fatalf("bracket = %#v", bracket)
	}
	final := bracket.Matches[2]
	if final.Round != 2 || final.Home != (SlotSource{Kind: Winner, SourceRound: 1, SourceSequence: 1}) || final.Away != (SlotSource{Kind: Winner, SourceRound: 1, SourceSequence: 2}) {
		t.Fatalf("final = %#v", final)
	}
}

func TestGenerateSingleEliminationCreatesDeterministicBye(t *testing.T) {
	bracket, err := GenerateSingleElimination([]string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if got := bracket.Matches[1].AutoWinnerTeamID; got != "b" {
		t.Fatalf("bye winner = %q, want b", got)
	}
	if bracket.Matches[1].Away.Kind != Bye {
		t.Fatalf("bye source = %#v", bracket.Matches[1].Away)
	}
}

func TestGenerateSingleEliminationNeverCreatesAByeOnlyMatch(t *testing.T) {
	bracket, err := GenerateSingleElimination([]string{"a", "b", "c", "d", "e"})
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range bracket.Matches {
		if match.Home.Kind == Bye && match.Away.Kind == Bye {
			t.Fatalf("match with two byes: %#v", match)
		}
	}
}

func TestGenerateSingleEliminationRejectsInvalidEntrants(t *testing.T) {
	for _, teams := range [][]string{{"a"}, {"a", "a"}, {"a", ""}} {
		if _, err := GenerateSingleElimination(teams); err != ErrInvalidBracketEntrants {
			t.Fatalf("GenerateSingleElimination(%#v) error = %v", teams, err)
		}
	}
}

func TestGenerateSeededSingleEliminationPairsExtremesAndSeparatesTopSeeds(t *testing.T) {
	bracket, err := GenerateSeededSingleElimination([]string{"1", "2", "3", "4", "5", "6", "7", "8"})
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]string{{"1", "8"}, {"4", "5"}, {"2", "7"}, {"3", "6"}}
	for index, pair := range want {
		match := bracket.Matches[index]
		if match.Home.TeamID != pair[0] || match.Away.TeamID != pair[1] {
			t.Fatalf("match %d = %s-%s, want %s-%s", index+1, match.Home.TeamID, match.Away.TeamID, pair[0], pair[1])
		}
	}
}

func TestGenerateSeededSingleEliminationRequiresACompleteField(t *testing.T) {
	if _, err := GenerateSeededSingleElimination([]string{"1", "2", "3"}); err != ErrInvalidBracketEntrants {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateSeededSingleEliminationSupportsEveryAcceptedFieldSize(t *testing.T) {
	for _, size := range []int{2, 4, 8, 16, 32, 64} {
		teams := make([]string, size)
		for index := range teams {
			teams[index] = string(rune('A' + index))
		}
		bracket, err := GenerateSeededSingleElimination(teams)
		if err != nil {
			t.Fatalf("size %d: %v", size, err)
		}
		if bracket.Size != size || len(bracket.Matches) != size-1 {
			t.Fatalf("size %d: bracket = %#v", size, bracket)
		}
	}
}
