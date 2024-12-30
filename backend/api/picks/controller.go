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
	INSERT INTO prop_picks (user_id, line_id, valid, date)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING
    RETURNING ID`
    var resId int
    err := db.QueryRow(context.Background(), sqlStmt, pick.UserId, pick.LineId, pick.Valid, pick.Date).Scan(&resId)
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
            pick.UserId,
            pick.LineId,
            pick.Valid,
            pick.Date,
        })
    }

	_, err = txn.CopyFrom(
        context.Background(),
        pgx.Identifier{"prop_picks_temp"},
        []string{
            "user_id",
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
        `INSERT INTO prop_picks (user_id, line_id, valid, date)
        SELECT user_id, line_id, valid, date FROM prop_picks_temp
        ON CONFLICT (user_id, line_id, date) DO UPDATE
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

func getPropPicks(userId int, date time.Time) ([]PropPick, error) {
    db := storage.GetDB()

    sql := `
    SELECT * from prop_picks
    WHERE user_id=($1) and date=($2)`

    row, _ := db.Query(context.Background(), sql, userId, date)
    strats, err := pgx.CollectRows(row, pgx.RowToStructByName[PropPick])
    if err != nil {
        return strats, errors.New(fmt.Sprintf("Error getting prop picks for user %d on %v: %v", userId, date, err))
    }

    return strats, nil
}

func GetPropPicks(c *gin.Context) {
    id, err := strconv.Atoi(c.Query("user_id"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    date, err := time.Parse("2006-01-02",c.Query("date"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }

    strats, err := getPropPicks(id, date)
    if err != nil {
        log.Println("Error in GetPropPicks:", err)
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, strats)
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
