package odds

import (
	"github.com/lib/pq"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayerLines(playerLines []PlayerLine) {
    db := storage.GetDB()
	txn, _ := db.Begin()
	_, err := txn.Exec(`
	CREATE TEMP TABLE player_lines_temp
	ON COMMIT DROP
	AS SELECT * FROM player_lines
	WITH NO DATA`)
	if err != nil {
		panic(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("player_lines_temp", "sport", "player_index", "timestamp", "stat", "side", "line", "odds"))
	if err != nil {
		panic(err)
	}

	for _, p := range playerLines {
		if _, err := stmt.Exec(p.Sport, p.PlayerIndex, p.Timestamp, p.Stat, p.Side, p.Line, p.Odds); err != nil {
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
	INSERT INTO player_lines (sport, player_index, timestamp, stat, side, line, odds)
	SELECT sport, player_index, timestamp, stat, side, line, odds FROM player_lines_temp
	ON CONFLICT DO NOTHING`)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}
