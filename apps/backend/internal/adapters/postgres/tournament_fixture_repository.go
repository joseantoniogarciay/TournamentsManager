package postgres

type fixture struct {
	round, sequence int
	home, away      string
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
