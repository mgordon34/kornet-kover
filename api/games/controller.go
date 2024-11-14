package games

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddGame(game Game) (int, error) {
    db := storage.GetDB()

    sqlStmt := `
	INSERT INTO games (sport, home_index, away_index, home_score, away_score, date)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING
    RETURNING ID`
    var resId int
    err := db.QueryRow(context.Background(), sqlStmt, game.Sport, game.HomeIndex, game.AwayIndex, game.HomeScore, game.AwayScore, game.Date).Scan(&resId)
	if err != nil {
        return 0, err
	}
    log.Printf("Added game: %v", game)
    return resId, nil
}

func GetLastGame() (Game, error) {
    db := storage.GetDB()

    sql := `
	SELECT * from games
    ORDER BY date DESC
    LIMIT 1`

    var game Game
    row, _ := db.Query(context.Background(), sql)
    game, err := pgx.CollectOneRow(row, pgx.RowToStructByName[Game])
    if err != nil {
        return Game{}, errors.New(fmt.Sprintf("Error getting last game: %v", err))
    }

    return game, nil
}
