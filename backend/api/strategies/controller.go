package strategies

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func AddStrategy(strat Strategy) (int, error) {
    db := storage.GetDB()

    sqlStmt := `
	INSERT INTO strategies (user_id, name)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING
    RETURNING ID`
    var resId int
    err := db.QueryRow(context.Background(), sqlStmt, strat.UserId, strat.Name).Scan(&resId)
	if err != nil {
        return 0, err
	}
    log.Printf("Added strategy: %v", strat)
    return resId, nil
}

func getStrategies(userId int) ([]Strategy, error) {
    db := storage.GetDB()

    sql := `
    SELECT * from strategies
    WHERE user_id = ($1)`

    row, _ := db.Query(context.Background(), sql, userId)
    strats, err := pgx.CollectRows(row, pgx.RowToStructByName[Strategy])
    if err != nil {
        return strats, errors.New(fmt.Sprintf("Error getting last game: %v", err))
    }

    return strats, nil
}

func GetStrategies(c *gin.Context) {
    id, err := strconv.Atoi(c.Query("user_id"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    log.Println(id)

    strats, err := getStrategies(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, strats)
}
