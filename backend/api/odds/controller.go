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
            playerLine.Type,
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
            "type",
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
        `INSERT INTO player_lines (sport, player_index, timestamp, stat, side, type, line, odds, link)
        SELECT sport, player_index, timestamp, stat, side, type, line, odds, link FROM player_lines_temp
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

func GetPlayerLinesForDate(date time.Time, lineType string) ([]PlayerLine, error) {
    date = date.UTC()
    endDate := date.AddDate(0, 0, 1)

    db := storage.GetDB()
    sql := `SELECT pl.id, pl.sport, pl.player_index, pl.timestamp, pl.stat, pl.side, pl.type, pl.line, pl.odds, pl.link FROM player_lines pl INNER JOIN
                (select player_index, stat, side, line, max(timestamp) as latest from player_lines where (timestamp between ($1) and ($2)) and type = ($3) group by player_index, stat, side, line) mpl 
                on pl.timestamp = mpl.latest and pl.player_index = mpl.player_index and pl.stat = mpl.stat and pl.side = mpl.side and pl.line = mpl.line;`

    rows, err := db.Query(context.Background(), sql, date, endDate, lineType)
    if err != nil {
        log.Fatal("Error querying for player lines: ", err)
    }
    defer rows.Close()
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

func GetLastLine(oddsType string) (PlayerLine, error) {
    db := storage.GetDB()

    sql := `
	SELECT id, sport, player_index, timestamp, stat, side, type, line, odds, link from player_lines
	where type = ($1)
    ORDER BY timestamp DESC
    LIMIT 1`

    row, _ := db.Query(context.Background(), sql, oddsType)
    defer row.Close()
    pLine, err := pgx.CollectOneRow(row, pgx.RowToStructByName[PlayerLine])
    if err != nil {
        return PlayerLine{}, errors.New(fmt.Sprintf("Error getting last game: %v", err))
    }

    return pLine, nil
}

func GetPlayerOddsForDate(date time.Time, stats []string) (map[string]map[string]PlayerOdds, error) {
    oddsMap := make(map[string]map[string]PlayerOdds)

    lines, err := GetPlayerLinesForDate(date, "mainline")
    if err != nil {
        return oddsMap, err
    }
    for _, line := range lines {
        addLineToOddsMap(oddsMap, line)
    }

    return oddsMap, nil
}

func GetAlternatePlayerOddsForDate(date time.Time, stats []string) (map[string]map[string][]PlayerLine, error) {
    oddsMap := make(map[string]map[string][]PlayerLine)

    lines, err := GetPlayerLinesForDate(date, "alternate")
    if err != nil {
        return oddsMap, err
    }
    for _, line := range lines {
        addAlternateLineToOddsMap(oddsMap, line)
    }

    return oddsMap, nil
}

func addAlternateLineToOddsMap(oddsMap map[string]map[string][]PlayerLine, line PlayerLine) {
    if _, ok := oddsMap[line.PlayerIndex]; !ok {
        oddsMap[line.PlayerIndex] = make(map[string][]PlayerLine)
    }
    if _, ok := oddsMap[line.PlayerIndex][line.Stat]; !ok {
        oddsMap[line.PlayerIndex][line.Stat] = []PlayerLine{}
    }

    oddsMap[line.PlayerIndex][line.Stat] = append(oddsMap[line.PlayerIndex][line.Stat], line)
}

func addLineToOddsMap(oddsMap map[string]map[string]PlayerOdds, line PlayerLine) {
    if _, ok := oddsMap[line.PlayerIndex]; !ok {
        oddsMap[line.PlayerIndex] = make(map[string]PlayerOdds)
    }
    if _, ok := oddsMap[line.PlayerIndex][line.Stat]; !ok {
        oddsMap[line.PlayerIndex][line.Stat] = PlayerOdds{}
    }

    pOdds := oddsMap[line.PlayerIndex][line.Stat]
    if line.Side == "Over" && isLineCloser(pOdds.Over, line, 0) {
        pOdds.Over = line
    } else if line.Side == "Under" && isLineCloser(pOdds.Under, line, 0) {
        pOdds.Under = line
    }
    oddsMap[line.PlayerIndex][line.Stat] = pOdds
}

func isLineCloser(currLine PlayerLine, newLine PlayerLine, target int) bool {
    if currLine.Odds == 0 {
        return true
    }
	curDiff := getDistanceFromTarget(currLine.Odds, target)
	newDiff := getDistanceFromTarget(newLine.Odds, target)

    return newDiff < curDiff
}

func getDistanceFromTarget(odds int, target int) float64 {
	var normalizedOdds int
	if odds < 0 {
		normalizedOdds = odds + 100
	} else {
		normalizedOdds = odds - 100
	}

	return math.Abs(float64(normalizedOdds - target))
}
