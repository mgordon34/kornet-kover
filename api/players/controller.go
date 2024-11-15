package players

import (
	"context"
	"log"
	// "log"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayers(players []Player) {
    db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
        context.Background(),
        ` CREATE TEMP TABLE players_temp
        ON COMMIT DROP
        AS SELECT * FROM players
        WITH NO DATA`,
    )
	if err != nil {
		panic(err)
	}
    var playersInterface [][]interface{}
    for _, player := range players {
        playersInterface = append(
            playersInterface,
            []interface{}{player.Index, player.Sport, player.Name},
        )
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"players_temp"},
        []string{"index", "sport", "name"},
        pgx.CopyFromRows(playersInterface),
    )
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO players (index, sport, name)
        SELECT index, sport, name FROM players_temp
        ON CONFLICT DO NOTHING`,
    )
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
    log.Printf("Added %d players to DB", len(players))
}

// func AddPlayerGames(pGames []PlayerGame) {
//     db := storage.GetDB()
// 	txn, _ := db.Begin()
// 	_, err := txn.Exec(`
// 	CREATE TEMP TABLE player_games_temp
// 	ON COMMIT DROP
// 	AS SELECT * FROM nba_player_games
// 	WITH NO DATA`)
// 	if err != nil {
// 		panic(err)
// 	}
// 
// 	stmt, err := txn.Prepare(pq.CopyIn("player_games_temp", "player_index", "game", "team_index",
//             "minutes", "points", "rebounds", "assists", "usg", "ortg", "drtg", ))
// 	if err != nil {
// 		panic(err)
// 	}
// 
// 	for _, p := range pGames {
// 		if _, err := stmt.Exec(p.PlayerIndex, p.Game, p.TeamIndex, p.Minutes, p.Points, p.Rebounds,
//                 p.Assists, p.Usg, p.Ortg, p.Drtg); err != nil {
// 			panic(err)
// 		}
// 	}
// 	if _, err := stmt.Exec(); err != nil {
// 		panic(err)
// 	}
// 	if err := stmt.Close(); err != nil {
// 		panic(err)
// 	}
// 
// 	_, err = txn.Exec(`
// 	INSERT INTO nba_player_games (player_index, game, team_index, minutes, points, rebounds, assists, usg, ortg, drtg)
// 	SELECT player_index, game, team_index, minutes, points, rebounds, assists, usg, ortg, drtg FROM player_games_temp
// 	ON CONFLICT DO NOTHING`)
// 	if err != nil {
// 		panic(err)
// 	}
// 
// 	if err := txn.Commit(); err != nil {
// 		panic(err)
// 	}
// }
// 
func PlayerNameToIndex(playerName string) (string, error) {
    db := storage.GetDB()
    sql := `SELECT index FROM players WHERE UPPER(name) LIKE UPPER($1);`

    var index string
    row := db.QueryRow(context.Background(), sql, playerName)
    if err := row.Scan(&index); err != nil {
        log.Printf("Error finding player index for %s", playerName)
        return "", err
    }
    return index, nil
}
