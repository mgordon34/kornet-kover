package picks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

type PicksServiceDeps struct {
	GetPropPicks   func(userID int, date time.Time) ([]PropPickFormatted, error)
	GetPropPick    func(pickID int) (PropPick, error)
	GetBettorPicks func(userID int, date time.Time) ([]BettorPickRow, error)
	Now            func() time.Time
	LoadLocation   func(name string) (*time.Location, error)
}

type PicksService struct {
	deps PicksServiceDeps
}

func NewPicksService(deps PicksServiceDeps) *PicksService {
	if deps.GetPropPicks == nil {
		deps.GetPropPicks = getPropPicks
	}
	if deps.GetPropPick == nil {
		deps.GetPropPick = getPropPick
	}
	if deps.GetBettorPicks == nil {
		deps.GetBettorPicks = getBettorPicks
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.LoadLocation == nil {
		deps.LoadLocation = time.LoadLocation
	}
	return &PicksService{deps: deps}
}

func (s *PicksService) GetPropPicksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		date, err := time.Parse("2006-01-02", c.Query("date"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		picks, err := s.deps.GetPropPicks(id, date)
		if err != nil {
			log.Println("Error in GetPropPicks:", err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, formatPicksByStrat(picks))
	}
}

func (s *PicksService) GetPropPickHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pickID, err := strconv.Atoi(c.Param("strat"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		log.Println(pickID)

		pick, err := s.deps.GetPropPick(pickID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, pick)
	}
}

func (s *PicksService) GetBettorPropPicksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}

		loc, err := s.deps.LoadLocation("America/New_York")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}
		t := s.deps.Now().In(loc)
		today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)

		rows, err := s.deps.GetBettorPicks(userID, today)
		if err != nil {
			log.Println("Error in GetBettorPropPicks:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, groupBettorPicksByStrategy(rows))
	}
}

func addPropPick(pick PropPick) (int, error) {
	db := storage.GetDB()

	sqlStmt := `
    INSERT INTO prop_picks (strat_id, line_id, valid, date)
    VALUES ($1, $2, $3, $4)
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

type PropPickFormatted struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	StratId   int       `json:"strat_id"`
	StratName string    `json:"strat_name"`
	Name      string    `json:"player_name"`
	Side      string    `json:"side"`
	Line      float32   `json:"line"`
	Stat      string    `json:"stat"`
	Odds      int       `json:"odds"`
	NumGames  int       `json:"num_games"`
	Points    float32   `json:"points"`
	Rebounds  float32   `json:"rebounds"`
	Assists   float32   `json:"assists"`
	Threes    float32   `json:"threes"`
	Minutes   float32   `json:"minutes"`
	Date      time.Time `json:"date"`
}

func getPropPicks(userId int, date time.Time) ([]PropPickFormatted, error) {
	db := storage.GetDB()

	sql := `
    SELECT pp.id, u.id as user_id, pp.strat_id, s.name as strat_name, p.name, pl.side, pl.line, pl.stat, pl.odds, 
    npp.num_games, npp.points, npp.rebounds, npp.assists, npp.threes, npp.minutes, pp.date from prop_picks pp
    LEFT JOIN player_lines pl on pl.id = pp.line_id
    LEFT JOIN players p on p.index = pl.player_index
    LEFT JOIN nba_pip_predictions npp on npp.player_index = pl.player_index and npp.date = pp.date
    LEFT JOIN strategies s on s.id = pp.strat_id
    LEFT JOIN users u on u.id = s.user_id
    WHERE pp.valid=true and u.id=($1) and pp.date=($2)`

	row, _ := db.Query(context.Background(), sql, userId, date)
	picks, err := pgx.CollectRows(row, pgx.RowToStructByName[PropPickFormatted])
	if err != nil {
		return picks, errors.New(fmt.Sprintf("Error getting prop picks for strat %d on %v: %v", userId, date, err))
	}
	defer row.Close()

	for i, pick := range picks {
		if math.IsNaN(float64(pick.Threes)) {
			picks[i].Threes = 0.0
		}
		if math.IsNaN(float64(pick.Assists)) {
			picks[i].Assists = 0.0
		}
		if math.IsNaN(float64(pick.Rebounds)) {
			picks[i].Rebounds = 0.0
		}
	}

	return picks, nil
}

type PropPicksResponse struct {
	StratId   int        `json:"strat_id"`
	StratName string     `json:"strat_name"`
	Picks     []PickInfo `json:"picks"`
}
type PickInfo struct {
	Id       int       `json:"id"`
	Name     string    `json:"player_name"`
	Side     string    `json:"side"`
	Line     float32   `json:"line"`
	Stat     string    `json:"stat"`
	Odds     int       `json:"odds"`
	NumGames int       `json:"num_games"`
	Points   float32   `json:"points"`
	Rebounds float32   `json:"rebounds"`
	Assists  float32   `json:"assists"`
	Threes   float32   `json:"threes"`
	Minutes  float32   `json:"minutes"`
	Date     time.Time `json:"date"`
}

func formatPicksByStrat(picks []PropPickFormatted) []PropPicksResponse {
	var stratKeys []int
	pickMap := make(map[int]PropPicksResponse)
	for _, pick := range picks {
		_, ok := pickMap[pick.StratId]
		if !ok {
			stratKeys = append(stratKeys, pick.StratId)
			pickMap[pick.StratId] = PropPicksResponse{
				StratId:   pick.StratId,
				StratName: pick.StratName,
			}
		}
		if math.IsNaN(float64(pick.Threes)) {
			pick.Threes = 0
		}

		pPick := pickMap[pick.StratId]
		pPick.Picks = append(pPick.Picks, PickInfo{
			Id:       pick.Id,
			Name:     pick.Name,
			Side:     pick.Side,
			Line:     pick.Line,
			Stat:     pick.Stat,
			Odds:     pick.Odds,
			NumGames: pick.NumGames,
			Points:   pick.Points,
			Rebounds: pick.Rebounds,
			Assists:  pick.Assists,
			Minutes:  pick.Minutes,
			Threes:   pick.Threes,
			Date:     pick.Date,
		})
		pickMap[pick.StratId] = pPick
	}

	var sortedPicks []PropPicksResponse
	sort.Ints(stratKeys)
	for _, stratId := range stratKeys {
		sortedPicks = append(sortedPicks, pickMap[stratId])
	}

	return sortedPicks
}

func GetPropPicks(c *gin.Context) {
	NewPicksService(PicksServiceDeps{}).GetPropPicksHandler()(c)
}

func getPropPick(stratId int) (PropPick, error) {
	db := storage.GetDB()

	sql := `
    SELECT * from prop_picks
    WHERE id=($1)`

	row, _ := db.Query(context.Background(), sql, stratId)
	defer row.Close()
	strat, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[PropPick])
	if err != nil {
		return strat, errors.New(fmt.Sprintf("Error getting prop pick %d: %v", stratId, err))
	}

	return strat, nil
}

func GetPropPick(c *gin.Context) {
	NewPicksService(PicksServiceDeps{}).GetPropPickHandler()(c)
}

func MarkOldPicksInvalid(stratId int, date time.Time) {
	db := storage.GetDB()

	sql := `
    UPDATE prop_picks
    SET valid=false
    WHERE strat_id=($1) AND date=($2)`

	_, err := db.Exec(context.Background(), sql, stratId, date)
	if err != nil {
		log.Fatal("Error marking old picks invalid: ", err)
	}
}

// BettorPickRow represents a single row from the database query for bettor picks
type BettorPickRow struct {
	ID         int     `db:"id"`
	StratID    int     `db:"strat_id"`
	StratName  string  `db:"strat_name"`
	PlayerName string  `db:"player_name"`
	TeamName   *string `db:"team_name"` // nullable
	Side       string  `db:"side"`
	Line       float32 `db:"line"`
	Stat       string  `db:"stat"`
	Odds       int     `db:"odds"`
	Points     float32 `db:"points"`
	Rebounds   float32 `db:"rebounds"`
	Assists    float32 `db:"assists"`
	Threes     float32 `db:"threes"`
}

// BettorPick represents a single pick formatted for bettors
type BettorPick struct {
	ID          int     `json:"id"`
	PlayerName  string  `json:"player_name"`
	Team        string  `json:"team"`
	LineDisplay string  `json:"line_display"`
	Predicted   float32 `json:"predicted"`
}

// BettorStrategyPicks groups picks by strategy for bettor view
type BettorStrategyPicks struct {
	StratID   int          `json:"strat_id"`
	StratName string       `json:"strat_name"`
	Picks     []BettorPick `json:"picks"`
}

// formatBettorLineDisplay formats line info as "over 12.5 points -110"
func formatBettorLineDisplay(side, stat string, line float32, odds int) string {
	return fmt.Sprintf("%s %.1f %s %d", side, line, stat, odds)
}

// getPredictedValue returns the predicted value for the given stat
func getPredictedValue(stat string, points, rebounds, assists, threes float32) float32 {
	switch stat {
	case "points":
		return points
	case "rebounds":
		return rebounds
	case "assists":
		return assists
	case "threes":
		return threes
	default:
		return 0
	}
}

// getBettorPicks retrieves prop picks formatted for bettors
func getBettorPicks(userId int, date time.Time) ([]BettorPickRow, error) {
	db := storage.GetDB()

	sql := `
    SELECT 
        pp.id,
        s.id as strat_id,
        s.name as strat_name,
        p.name as player_name,
        t.name as team_name,
        pl.side,
        pl.line,
        pl.stat,
        pl.odds,
        COALESCE(npp.points, 0) as points,
        COALESCE(npp.rebounds, 0) as rebounds,
        COALESCE(npp.assists, 0) as assists,
        COALESCE(npp.threes, 0) as threes
    FROM prop_picks pp
    INNER JOIN strategies s ON s.id = pp.strat_id
    INNER JOIN player_lines pl ON pl.id = pp.line_id
    INNER JOIN players p ON p.index = pl.player_index
    LEFT JOIN active_rosters ar ON ar.player_index = pl.player_index AND ar.sport = pl.sport
    LEFT JOIN teams t ON t.index = ar.team_index
    LEFT JOIN LATERAL (
        SELECT points, rebounds, assists, threes
        FROM nba_pip_predictions npp
        WHERE npp.player_index = pl.player_index AND npp.date = pp.date
        ORDER BY npp.version DESC
        LIMIT 1
    ) npp ON true
    WHERE pp.valid = true
        AND s.user_id = ($1)
        AND pp.date = ($2)
    ORDER BY s.id, pp.id`

	rows, err := db.Query(context.Background(), sql, userId, date)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error querying bettor picks for user %d on %v: %v", userId, date, err))
	}
	defer rows.Close()

	pickRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[BettorPickRow])
	if err != nil {
		return pickRows, errors.New(fmt.Sprintf("Error getting bettor picks for user %d on %v: %v", userId, date, err))
	}

	// Handle NaN values
	for i, row := range pickRows {
		if math.IsNaN(float64(row.Threes)) {
			pickRows[i].Threes = 0.0
		}
		if math.IsNaN(float64(row.Assists)) {
			pickRows[i].Assists = 0.0
		}
		if math.IsNaN(float64(row.Rebounds)) {
			pickRows[i].Rebounds = 0.0
		}
		if math.IsNaN(float64(row.Points)) {
			pickRows[i].Points = 0.0
		}
	}

	return pickRows, nil
}

// groupBettorPicksByStrategy groups bettor picks by strategy
func groupBettorPicksByStrategy(rows []BettorPickRow) []BettorStrategyPicks {
	var stratKeys []int
	pickMap := make(map[int]BettorStrategyPicks)

	for _, row := range rows {
		_, ok := pickMap[row.StratID]
		if !ok {
			stratKeys = append(stratKeys, row.StratID)
			pickMap[row.StratID] = BettorStrategyPicks{
				StratID:   row.StratID,
				StratName: row.StratName,
			}
		}

		team := ""
		if row.TeamName != nil {
			team = *row.TeamName
		}

		predicted := getPredictedValue(row.Stat, row.Points, row.Rebounds, row.Assists, row.Threes)

		pPick := pickMap[row.StratID]
		pPick.Picks = append(pPick.Picks, BettorPick{
			ID:          row.ID,
			PlayerName:  row.PlayerName,
			Team:        team,
			LineDisplay: formatBettorLineDisplay(row.Side, row.Stat, row.Line, row.Odds),
			Predicted:   predicted,
		})
		pickMap[row.StratID] = pPick
	}

	var sortedPicks []BettorStrategyPicks
	sort.Ints(stratKeys)
	for _, stratId := range stratKeys {
		sortedPicks = append(sortedPicks, pickMap[stratId])
	}

	return sortedPicks
}

// GetBettorPropPicks handles GET /prop-picks/bettor endpoint
func GetBettorPropPicks(c *gin.Context) {
	NewPicksService(PicksServiceDeps{}).GetBettorPropPicksHandler()(c)
}
