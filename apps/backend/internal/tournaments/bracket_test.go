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
