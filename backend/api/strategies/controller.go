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

type StrategyServiceDeps struct {
	GetStrategies func(userID int) ([]Strategy, error)
	GetStrategy   func(strategyID int) (Strategy, error)
}

type StrategyService struct {
	deps StrategyServiceDeps
}

func NewStrategyService(deps StrategyServiceDeps) *StrategyService {
	if deps.GetStrategies == nil {
		deps.GetStrategies = getStrategies
	}
	if deps.GetStrategy == nil {
		deps.GetStrategy = getStrategy
	}
	return &StrategyService{deps: deps}
}

func (s *StrategyService) GetStrategiesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		log.Println(id)

		strats, err := s.deps.GetStrategies(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, strats)
	}
}

func (s *StrategyService) GetStrategyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stratID, err := strconv.Atoi(c.Param("strat"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		log.Println(stratID)

		strat, err := s.deps.GetStrategy(stratID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, strat)
	}
}

func addStrategy(strat Strategy) (int, error) {
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

	row, err := db.Query(context.Background(), sql, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying strategies for user %d: %w", userId, err)
	}
	strats, err := pgx.CollectRows(row, pgx.RowToStructByName[Strategy])
	if err != nil {
		return strats, errors.New(fmt.Sprintf("Error getting strategies for user %d: %v", userId, err))
	}

	return strats, nil
}

func GetStrategies(c *gin.Context) {
	NewStrategyService(StrategyServiceDeps{}).GetStrategiesHandler()(c)
}

func getStrategy(stratId int) (Strategy, error) {
	db := storage.GetDB()

	sql := `
    SELECT * from strategies
    WHERE id = ($1)`

	row, err := db.Query(context.Background(), sql, stratId)
	if err != nil {
		return Strategy{}, fmt.Errorf("error querying strategy %d: %w", stratId, err)
	}
	strat, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[Strategy])
	if err != nil {
		return strat, errors.New(fmt.Sprintf("Error getting strategy %d: %v", stratId, err))
	}

	return strat, nil
}

func GetStrategy(c *gin.Context) {
	NewStrategyService(StrategyServiceDeps{}).GetStrategyHandler()(c)
}
