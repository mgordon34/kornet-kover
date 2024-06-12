package odds

import (
	"github.com/lib/pq"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayerOdds(playerOdds []PlayerOdds) {
    db := storage.GetDB()
	txn, _ := db.Begin()
	_, err := txn.Exec(`
	CREATE TEMP TABLE player_odds_temp
	ON COMMIT DROP
	AS SELECT * FROM player_odds
	WITH NO DATA`)
	if err != nil {
		panic(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("player_odds_temp", "player_index", "date", "stat", "line", "over_odds", "under_odds"))
	if err != nil {
		panic(err)
	}

	for _, p := range playerOdds {
		if _, err := stmt.Exec(p.PlayerIndex, p.Date, p.Stat, p.Line, p.OverOdds, p.UnderOdds); err != nil {
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
	INSERT INTO player_odds (player_index, date, stat, line, over_odds, under_odds)
	SELECT player_index, date, stat, line, over_odds, under_odds FROM player_odds_temp
	ON CONFLICT DO NOTHING`)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}
