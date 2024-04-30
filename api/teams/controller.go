package teams

import (
	"github.com/lib/pq"
    "github.com/mgordon34/kornet-kover/internal/storage"
)

func AddTeams(teams []Team) {
    db := storage.GetDB()
	txn, _ := db.Begin()
	_, err := txn.Exec(`
	CREATE TEMP TABLE teams_temp
	ON COMMIT DROP
	AS SELECT * FROM teams
	WITH NO DATA`)
	if err != nil {
		panic(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("teams_temp", "index", "name"))
	if err != nil {
		panic(err)
	}

	for _, t := range teams {
		if _, err := stmt.Exec(t.Index, t.Name); err != nil {
			panic(err)
		}
	}
	if _, err := stmt.Exec(); err != nil {
		panic(err)
	}
	if err := stmt.Close(); err != nil {
		panic(err)
	}

	_, err = txn.Exec(`
	INSERT INTO teams (index, name)
	SELECT * FROM teams_temp
	ON CONFLICT DO NOTHING`)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(); err != nil {
		panic(err)
	}
}
