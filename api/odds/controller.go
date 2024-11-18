package odds

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayerLines(playerLines []PlayerLine) {
    log.Printf("Adding %d player lines", len(playerLines))
    db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
        context.Background(),
        `CREATE TEMP TABLE player_lines_temp
        ON COMMIT DROP
        AS SELECT * FROM player_lines
        WITH NO DATA`,
    )
	if err != nil {
        log.Fatal(err)
	}

    var teamsInterface [][]interface{}
    for _, playerLine := range playerLines {
        teamsInterface = append(teamsInterface, []interface{}{
            playerLine.Sport,
            playerLine.PlayerIndex,
            playerLine.Timestamp,
            playerLine.Stat,
            playerLine.Side,
            playerLine.Line,
            playerLine.Odds,
        })
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"player_lines_temp"},
        []string{
            "sport",
            "player_index",
            "timestamp",
            "stat",
            "side",
            "line",
            "odds",
        },
        pgx.CopyFromRows(teamsInterface),
    )
	if err != nil {
        log.Fatal(err)
	}


	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO player_lines (sport, player_index, timestamp, stat, side, line, odds)
        SELECT sport, player_index, timestamp, stat, side, line, odds FROM player_lines_temp
        ON CONFLICT DO NOTHING`,
    )
	if err != nil {
        log.Fatal(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
        log.Fatal(err)
	}
    log.Println("success adding player_lines")
}

func GetPlayerLinesForDate(date time.Time) ([]PlayerLine, error) {
    startDate := date.Add(time.Hour * 7)
    endDate := startDate.AddDate(0, 0, 1)

    db := storage.GetDB()
    sql := `SELECT pl.sport, pl.player_index, pl.timestamp, pl.stat, pl.side, pl.line, pl.odds FROM player_lines pl INNER JOIN
                (select player_index, stat, side, max(timestamp) as latest from player_lines where timestamp between ($1) and ($2) group by player_index, stat, side) mpl 
                on pl.timestamp = mpl.latest and pl.player_index = mpl.player_index and pl.stat = mpl.stat and pl.side = mpl.side;`

    rows, err := db.Query(context.Background(), sql, startDate, endDate)
    if err != nil {
        log.Fatal("Error querying for player lines: ", err)
    }
    pLines, err := pgx.CollectRows(rows, pgx.RowToStructByName[PlayerLine])
    if err != nil {
        log.Fatal("Error converting rows to playerLines: ", err)
    }

    return pLines, nil
}

type PlayerOdds struct {
    Over    PlayerLine
    Under   PlayerLine
}

func GetPlayerOddsForDate(date time.Time) (map[string]PlayerOdds, error) {
    oddsMap := make(map[string]PlayerOdds)
    lines, err := GetPlayerLinesForDate(date)
    if err != nil {
        return oddsMap, err
    }
    for _, line := range lines {
        log.Println(line)
        addLineToOddsMap(oddsMap, line)
    }

    return oddsMap, nil
}

func addLineToOddsMap(oddsMap map[string]PlayerOdds, line PlayerLine) {
    if _, ok := oddsMap[line.PlayerIndex]; !ok {
        oddsMap[line.PlayerIndex] = PlayerOdds{}
    }

    pOdds := oddsMap[line.PlayerIndex]
    if line.Side == "Over" {
        pOdds.Over = line
    } else {
        pOdds.Under = line
    }
    oddsMap[line.PlayerIndex] = pOdds
}
