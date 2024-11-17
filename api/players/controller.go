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
}

func AddPlayerGames(pGames []PlayerGame) {
    db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
        context.Background(),
        `CREATE TEMP TABLE player_games_temp
        ON COMMIT DROP
        AS SELECT * FROM nba_player_games
        WITH NO DATA`,
    )
	if err != nil {
		panic(err)
	}
    var playersInterface [][]interface{}
    for _, pGame := range pGames {
        playersInterface = append(
            playersInterface,
            []interface{}{
                pGame.PlayerIndex,
                pGame.Game,
                pGame.TeamIndex,
                pGame.Minutes,
                pGame.Points,
                pGame.Rebounds,
                pGame.Assists,
                pGame.Usg,
                pGame.Ortg,
                pGame.Drtg,
            },
        )
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"player_games_temp"},
        []string{
            "player_index",
            "game",
            "team_index",
            "minutes",
            "points",
            "rebounds",
            "assists",
            "usg",
            "ortg",
            "drtg",
        },
        pgx.CopyFromRows(playersInterface),
    )
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
        context.Background(),
        ` INSERT INTO nba_player_games (player_index, game, team_index, minutes, points, rebounds, assists, usg, ortg, drtg)
        SELECT player_index, game, team_index, minutes, points, rebounds, assists, usg, ortg, drtg FROM player_games_temp
        ON CONFLICT DO NOTHING`,
    )
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

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
