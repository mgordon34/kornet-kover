package teams

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddTeams(teams []Team) {
    db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
        context.Background(),
        `CREATE TEMP TABLE teams_temp
        ON COMMIT DROP
        AS SELECT * FROM teams
        WITH NO DATA`,
    )
	if err != nil {
		panic(err)
	}

    var teamsInterface [][]interface{}
    for _, team := range teams {
        teamsInterface = append(teamsInterface, []interface{}{team.Index, team.Name})
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"teams_temp"},
        []string{"index", "name"},
        pgx.CopyFromRows(teamsInterface),
    )
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO teams (index, name)
        SELECT * FROM teams_temp
        ON CONFLICT DO NOTHING`,
    )
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}
