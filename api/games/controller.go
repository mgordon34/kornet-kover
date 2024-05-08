package games

import (
	"log"
    "errors"

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

    res, err := txn.Exec(`
	INSERT INTO games (sport, home_index, away_index, home_score, away_score, date)
	SELECT sport, home_index, away_index, home_score, away_score, date FROM games_temp
	ON CONFLICT DO NOTHING
    RETURNING ID`)
	if err != nil {
		panic(err)
	}
    id, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
    log.Printf("ID from game %v", id)

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}

func AddGame(game Game) (int, error) {
    db := storage.GetDB()

    sqlStmt := `
	INSERT INTO games (sport, home_index, away_index, home_score, away_score, date)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING
    RETURNING ID`
    var resId int
    err := db.QueryRow(sqlStmt, game.Sport, game.HomeIndex, game.AwayIndex, game.HomeScore, game.AwayScore, game.Date).Scan(&resId)
	if err != nil {
        return 0, errors.New("Row not written")
	}
    return resId, nil
}
