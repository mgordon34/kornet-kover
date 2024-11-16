package odds

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddPlayerLines(playerLines []PlayerLine) {
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
		panic(err)
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
		panic(err)
	}


	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO player_lines (sport, player_index, timestamp, stat, side, line, odds)
        SELECT sport, player_index, timestamp, stat, side, line, odds FROM player_lines_temp
        ON CONFLICT DO NOTHING`,
    )
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func GetPlayerLinesForPlayer(player players.Player) ([]PlayerLine, error) {
    playerLines := []PlayerLine{}
    return playerLines, nil
}
