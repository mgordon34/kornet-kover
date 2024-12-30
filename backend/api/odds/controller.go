package odds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
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
            playerLine.Link,
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
            "link",
        },
        pgx.CopyFromRows(teamsInterface),
    )
	if err != nil {
        log.Fatal(err)
	}


	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO player_lines (sport, player_index, timestamp, stat, side, line, odds, link)
        SELECT sport, player_index, timestamp, stat, side, line, odds, link FROM player_lines_temp
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
    date = date.UTC()
    startDate := date.AddDate(0, 0, -1)
    endDate := date.AddDate(0, 0, 1)

    db := storage.GetDB()
    sql := `SELECT pl.id, pl.sport, pl.player_index, pl.timestamp, pl.stat, pl.side, pl.line, pl.odds, pl.link FROM player_lines pl INNER JOIN
                (select player_index, stat, side, line, max(timestamp) as latest from player_lines where timestamp between ($1) and ($2) group by player_index, stat, side, line) mpl 
                on pl.timestamp = mpl.latest and pl.player_index = mpl.player_index and pl.stat = mpl.stat and pl.side = mpl.side and pl.line = mpl.line;`

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

func GetLastLine() (PlayerLine, error) {
    db := storage.GetDB()

    sql := `
	SELECT id, sport, player_index, timestamp, stat, side, line, odds, link from player_lines
    ORDER BY timestamp DESC
    LIMIT 1`

    row, _ := db.Query(context.Background(), sql)
    pLine, err := pgx.CollectOneRow(row, pgx.RowToStructByName[PlayerLine])
    if err != nil {
        return PlayerLine{}, errors.New(fmt.Sprintf("Error getting last game: %v", err))
    }

    return pLine, nil
}

func GetPlayerOddsForDate(date time.Time, stats []string) (map[string]map[string]PlayerOdds, error) {
    oddsMap := make(map[string]map[string]PlayerOdds)

    lines, err := GetPlayerLinesForDate(date)
    if err != nil {
        return oddsMap, err
    }
    for _, line := range lines {
        addLineToOddsMap(oddsMap, line)
    }

    return oddsMap, nil
}

func addLineToOddsMap(oddsMap map[string]map[string]PlayerOdds, line PlayerLine) {
    if _, ok := oddsMap[line.PlayerIndex]; !ok {
        oddsMap[line.PlayerIndex] = make(map[string]PlayerOdds)
    }
    if _, ok := oddsMap[line.PlayerIndex][line.Stat]; !ok {
        oddsMap[line.PlayerIndex][line.Stat] = PlayerOdds{}
    }

    pOdds := oddsMap[line.PlayerIndex][line.Stat]
    if line.Side == "Over" && isLineCloser(pOdds.Over, line) {
        pOdds.Over = line
    } else if isLineCloser(pOdds.Under, line) {
        pOdds.Under = line
    }
    oddsMap[line.PlayerIndex][line.Stat] = pOdds
}

func isLineCloser(curLine PlayerLine, newLine PlayerLine) bool {
    if curLine.Odds == 0 {
        return true
    }
    curOdds := math.Abs(math.Abs(float64(curLine.Odds)))
    newOdds := math.Abs(math.Abs(float64(newLine.Odds)))

    return newOdds < curOdds
}
