package games

import (
	"github.com/lib/pq"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func AddGames(games []Game) {
    db := storage.GetDB()
	txn, _ := db.Begin()
	_, err := txn.Exec(`
	CREATE TEMP TABLE games_temp
	ON COMMIT DROP
	AS SELECT * FROM games
	WITH NO DATA`)
	if err != nil {
		panic(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("games_temp", "sport", "home_index", "away_index", "home_score", "away_score", "date"))
	if err != nil {
		panic(err)
	}

	for _, g := range games {
		if _, err := stmt.Exec(g.Sport, g.HomeIndex, g.AwayIndex, g.HomeScore, g.AwayScore, g.Date); err != nil {
			panic(err)
		}
	}
	if _, err := stmt.Exec(); err != nil {
		panic(err)
	}
	if err := stmt.Close(); err != nil {
		panic(err)
	}

	_, err = txn.Exec(`
	INSERT INTO games (sport, home_index, away_index, home_score, away_score, date)
	SELECT sport, home_index, away_index, home_score, away_score, date FROM games_temp
	ON CONFLICT DO NOTHING`)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}
