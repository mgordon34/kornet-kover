package picks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func addPropPick(pick PropPick) (int, error) {
    db := storage.GetDB()

    sqlStmt := `
	INSERT INTO prop_picks (strat_id, line_id, valid, date)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING
    RETURNING ID`
    var resId int
    err := db.QueryRow(context.Background(), sqlStmt, pick.StratId, pick.LineId, pick.Valid, pick.Date).Scan(&resId)
	if err != nil {
        return 0, err
	}
    log.Printf("Added prop pick: %v", pick)
    return resId, nil
}

func AddPropPicks(picks []PropPick) error {
    log.Printf("Adding %d prop picks", len(picks))
    db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
        context.Background(),
        `CREATE TEMP TABLE prop_picks_temp
        ON COMMIT DROP
        AS SELECT * FROM prop_picks
        WITH NO DATA`,
    )
	if err != nil {
       return err
	}

    var picksInterface [][]interface{}
    for _, pick := range picks {
        picksInterface = append(picksInterface, []interface{}{
            pick.StratId,
            pick.LineId,
            pick.Valid,
            pick.Date,
        })
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"prop_picks_temp"},
        []string{
            "strat_id",
            "line_id",
            "valid",
            "date",
        },
        pgx.CopyFromRows(picksInterface),
    )
	if err != nil {
        return err
	}


	_, err = txn.Exec(
        context.Background(),
        `INSERT INTO prop_picks (strat_id, line_id, valid, date)
        SELECT strat_id, line_id, valid, date FROM prop_picks_temp
        ON CONFLICT (strat_id, line_id, date) DO UPDATE
        SET valid=excluded.valid`,
    )
	if err != nil {
        return err
	}

	if err := txn.Commit(context.Background()); err != nil {
        return err
	}
    log.Println("success adding prop_picks")

    return nil
}

func getPropPicks(stratId int, date time.Time) ([]PropPick, error) {
    db := storage.GetDB()

    sql := `
    SELECT * from prop_picks
    WHERE strat_id=($1) and date=($2)`

    row, _ := db.Query(context.Background(), sql, stratId, date)
    strats, err := pgx.CollectRows(row, pgx.RowToStructByName[PropPick])
    if err != nil {
        return strats, errors.New(fmt.Sprintf("Error getting prop picks for strat %d on %v: %v", stratId, date, err))
    }

    return strats, nil
}

func GetPropPicks(c *gin.Context) {
    id, err := strconv.Atoi(c.Query("strat_id"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    date, err := time.Parse("2006-01-02",c.Query("date"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }

    picks, err := getPropPicks(id, date)
    if err != nil {
        log.Println("Error in GetPropPicks:", err)
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, picks)
}

func getPropPick(stratId int) (PropPick, error) {
    db := storage.GetDB()

    sql := `
    SELECT * from prop_picks
    WHERE id=($1)`

    row, _ := db.Query(context.Background(), sql, stratId)
    strat, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[PropPick])
    if err != nil {
        return strat, errors.New(fmt.Sprintf("Error getting prop pick %d: %v", stratId, err))
    }

    return strat, nil
}

func GetPropPick(c *gin.Context) {
    pickId, err := strconv.Atoi(c.Param("strat"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    log.Println(pickId)

    strat, err := getPropPick(pickId)
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, strat)
}

func MarkOldPicksInvalid(stratId int, date time.Time) {
    db := storage.GetDB()

    sql := `
    UPDATE prop_picks
    SET valid=false
    WHERE strat_id=($1) AND date=($2)`

    db.QueryRow(context.Background(), sql, stratId, date)
}
