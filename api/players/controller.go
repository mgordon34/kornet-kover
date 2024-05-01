package players

import (
	"github.com/lib/pq"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayers(players []Player) {
    db := storage.GetDB()
	txn, _ := db.Begin()
	_, err := txn.Exec(`
	CREATE TEMP TABLE players_temp
	ON COMMIT DROP
	AS SELECT * FROM players
	WITH NO DATA`)
	if err != nil {
		panic(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("players_temp", "index", "sport", "name"))
	if err != nil {
		panic(err)
	}

	for _, p := range players {
		if _, err := stmt.Exec(p.Index, p.Sport, p.Name); err != nil {
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
	INSERT INTO players (index, sport, name)
	SELECT * FROM players_temp
	ON CONFLICT DO NOTHING`)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}
