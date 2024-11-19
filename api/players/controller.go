package players

import (
	"context"
	"strings"
    "time"
	"log"

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

func PlayerNameToIndex(playerName string) (string, error) {
    db := storage.GetDB()
    sql := `SELECT index FROM players WHERE UPPER(name) LIKE UPPER($1);`
    playerName = strings.ReplaceAll(playerName, ".", "")
    if playerName == "Alexandre Sarr" {
        playerName = "Alex Sarr"
    }

    var index string
    row := db.QueryRow(context.Background(), sql, playerName)
    if err := row.Scan(&index); err != nil {
        log.Printf("Error finding player index for %s", playerName)
        return "", err
    }
    return index, nil
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

func GetPlayerStats(player Player, startDate time.Time, endDate time.Time) (NBAAvg, error) {
    db := storage.GetDB()
    sql := `SELECT count(*) as num_games, avg(minutes) as minutes, avg(points) as points, avg(rebounds) as rebounds, 
            avg(assists) as assists, avg(usg) as usg, avg(ortg) as ortg, avg(drtg) as drtg FROM nba_player_games
                left join games on games.id = nba_player_games.game
                where nba_player_games.player_index = ($1) and games.date between ($2) and ($3)`

    rows, err := db.Query(context.Background(), sql, player.Index, startDate, endDate)
    if err != nil {
        log.Fatal("Error querying for player lines: ", err)
    }
    stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NBAAvg])
    if err != nil {
        log.Fatal("Error converting rows to playerLines: ", err)
    }

    return stats, nil
}
