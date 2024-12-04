package players

import (
	"context"
	"log"
	"strings"
	"time"

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

func GetPlayer(index string) (Player, error) {
    db := storage.GetDB()
    sql := `SELECT index, sport, name from players where index = ($1)`

    rows, err := db.Query(context.Background(), sql, index)
    if err != nil {
        log.Fatal("Error querying for player index: ", err)
    }

    player, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Player])
    if err != nil {
        log.Println("Error converting player: ", err)
        return Player{}, err
    }

    return player, nil
}

func PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
    playerName = strings.ReplaceAll(playerName, ".", "")
    index, ok := nameMap[playerName]; if ok {
        return index, nil
    }

    db := storage.GetDB()
    sql := `SELECT index FROM players WHERE UPPER(name) LIKE UPPER($1);`
    if playerName == "Alexandre Sarr" {
        playerName = "Alex Sarr"
    }

    row := db.QueryRow(context.Background(), sql, playerName)
    if err := row.Scan(&index); err != nil {
        log.Printf("Error finding player index for %s", playerName)
        return "", err
    }
    nameMap[playerName] = index
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

func GetPlayerStats(player string, startDate time.Time, endDate time.Time) (PlayerAvg, error) {
    db := storage.GetDB()
    sql := `SELECT count(*) as num_games, avg(minutes) as minutes, avg(points) as points, avg(rebounds) as rebounds, 
            avg(assists) as assists, avg(usg) as usg, avg(ortg) as ortg, avg(drtg) as drtg FROM nba_player_games
                left join games on games.id = nba_player_games.game
                where nba_player_games.player_index = ($1) and nba_player_games.minutes > 10 and games.date between ($2) and ($3)`

    rows, err := db.Query(context.Background(), sql, player, startDate, endDate)
    if err != nil {
        log.Fatal("Error querying for player stats: ", err)
    }

    stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NBAAvg])
    if err != nil {
        return NBAAvg{}, err
    }

    return stats, nil
}

type Relationship int

const (
    Teammate Relationship = iota
    Opponent
)

func GetPlayersForGame(gameId int, homeIndex string) (map[string][]Player, error) {
    playerMap := make(map[string][]Player)
    db := storage.GetDB()
    sql := `SELECT pl.index, pl.name, pg.team_index FROM players pl
                 LEFT JOIN nba_player_games pg ON pg.player_index=pl.index
                 LEFT JOIN games gg ON gg.id=pg.game
                 WHERE gg.id=($1)
                 ORDER BY pg.minutes DESC`
    rows, err := db.Query(context.Background(), sql, gameId)
    if err != nil {
        log.Fatal("Error querying for player lines: ", err)
    }
    defer rows.Close()
    for rows.Next() {
        var player Player
        var teamIndex string
        err = rows.Scan(&player.Index, &player.Name, &teamIndex)
        if err != nil {
            log.Fatal("Error converting rows to playerLines: ", err)
        }
        log.Printf("Player %v: on team %s", player, teamIndex)

        if teamIndex == homeIndex {
            playerMap["home"] = append(playerMap["home"], player)
        } else {
            playerMap["away"] = append(playerMap["away"], player)
        }
    }

    return playerMap, nil
}

func GetPlayerPerByYear(player string, startDate time.Time, endDate time.Time) map[int]PlayerAvg {
    playerStats := make(map[int]PlayerAvg)

    for d := startDate; d.After(endDate) == false; d = d.AddDate(1, 0, 0) {
        useDate := d.AddDate(1, 0, 0)
        if useDate.After(endDate){
            useDate = endDate
        }

        yearlyStats, _ := GetPlayerStats(player, d, useDate)
        playerStats[d.Year()] = yearlyStats.ConvertToPer()
    }

    return playerStats
}

func GetPlayerStatsWithPlayer(player string, defender string, relationship Relationship, startDate time.Time, endDate time.Time) (PlayerAvg, error) {
    db := storage.GetDB()
    sql := `SELECT count(*) as num_games, avg(minutes) as minutes, avg(points) as points, avg(rebounds) as rebounds, 
            avg(assists) as assists, avg(usg) as usg, avg(ortg) as ortg, avg(drtg) as drtg FROM nba_player_games
                left join games gg on gg.id = nba_player_games.game
                where nba_player_games.player_index = ($1) and nba_player_games.minutes > 10 and gg.date between ($3) and ($4)`
    opponent_filter := `
        AND (
            SELECT COUNT(*) FROM games ga
            LEFT JOIN nba_player_games pg ON pg.game=ga.id
            WHERE ga.id=gg.id AND pg.player_index IN (($1),($2))
        ) > 1`
    teammate_filter := `
         AND (
             SELECT COUNT(*) FROM games ga
             LEFT JOIN nba_player_games pg ON pg.game=ga.id
             WHERE ga.id=gg.id AND pg.player_index=($1)
         ) = 1
         AND (
             SELECT COUNT(*) FROM games ga
             LEFT JOIN nba_player_games pg ON pg.game=ga.id
             WHERE ga.id=gg.id AND pg.player_index=($2)
         ) = 0`

    switch relationship {
    case Teammate:
        sql = sql + teammate_filter
    case Opponent:
        sql = sql + opponent_filter
    }

    rows, err := db.Query(context.Background(), sql, player, defender, startDate, endDate)
    if err != nil {
        log.Fatal("Error querying for player stats: ", err)
    }

    stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NBAAvg])
    if err != nil {
        return NBAAvg{}, err
    }

    return stats, nil
}

func GetPlayerPerWithPlayerByYear(player string, defender string, relationship Relationship, startDate time.Time, endDate time.Time) map[int]PlayerAvg {
    playerStats := make(map[int]PlayerAvg)

    for d := startDate; d.After(endDate) == false; d = d.AddDate(1, 0, 0) {
        useDate := d.AddDate(1, 0, 0)
        if useDate.After(endDate){
            useDate = endDate
        }

        yearlyStats, _ := GetPlayerStatsWithPlayer(player, defender, relationship, d, useDate)
        playerStats[d.Year()] = yearlyStats.ConvertToPer()
    }

    return playerStats
}

func CalculatePIPFactor(controlMap map[int]PlayerAvg, relatedMap map[int]PlayerAvg) PlayerAvg {
    var totals PlayerAvg
    for year := range controlMap {
        pChange := relatedMap[year].CompareAvg(controlMap[year])
        if totals == nil {
            totals = pChange
        } else {
            totals = totals.AddAvg(pChange)
        }
    }

    return totals
}
