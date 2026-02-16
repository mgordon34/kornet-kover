package sportsbook

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type SportsbookPullType int

const (
	Live SportsbookPullType = iota
	Historical
)

var odds_markets = map[string]string{
	"player_points":   "points",
	"player_rebounds": "rebounds",
	"player_assists":  "assists",
	"player_threes":   "threes",
}

type OddsServiceDeps struct {
	Sources        SportsbookSources
	Store          SportsbookStore
	Now            func() time.Time
	RunGetOdds     func(startDate time.Time, endDate time.Time, oddsType string)
	RunGetLiveOdds func(date time.Time, oddsType string)
}

type OddsService struct {
	deps OddsServiceDeps
}

func NewOddsService(deps OddsServiceDeps) *OddsService {
	if deps.Sources == nil {
		deps.Sources = defaultSportsbookSources{}
	}
	if deps.Store == nil {
		deps.Store = defaultSportsbookStore{}
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}

	svc := &OddsService{deps: deps}

	if deps.RunGetOdds != nil {
		svc.deps.RunGetOdds = deps.RunGetOdds
	} else {
		svc.deps.RunGetOdds = svc.GetOdds
	}
	if deps.RunGetLiveOdds != nil {
		svc.deps.RunGetLiveOdds = deps.RunGetLiveOdds
	} else {
		svc.deps.RunGetLiveOdds = svc.GetLiveOdds
	}

	return svc
}

func requestOddsAPI(endpoint string, addlArgs []string) (response string, err error) {
	base_url := "https://api.the-odds-api.com/v4/" + endpoint + "?"
	args := []string{
		"apiKey=" + os.Getenv("ODDS_API_KEY"),
	}
	args = append(args, addlArgs...)

	res, err := http.Get(base_url + strings.Join(args[:], "&"))
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		fmt.Printf("error converting response body: %s\n", err)
		os.Exit(1)
	}

	return buf.String(), err
}

func UpdateLinesHandler(service *OddsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			c.JSON(http.StatusInternalServerError, "odds service is not configured")
			return
		}

		err := service.UpdateLines()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, "Done")
	}
}

func (s *OddsService) UpdateLines() error {
	lastLine, err := s.deps.Store.GetLastLine("mainline")
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Last line: %v", lastLine)

	loc, _ := time.LoadLocation("America/New_York")
	d := lastLine.Timestamp.In(loc)
	t := s.deps.Now().In(loc)
	startDate := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	s.deps.RunGetOdds(startDate, today, "mainline")
	s.deps.RunGetOdds(startDate, today, "alternate")
	s.deps.RunGetLiveOdds(today, "mainline")
	s.deps.RunGetLiveOdds(today, "alternate")

	return nil
}

type EventsResponse struct {
	Timestamp         time.Time   `json:"timestamp"`
	PreviousTimestamp time.Time   `json:"previous_timestamp"`
	NextTimestamp     time.Time   `json:"next_timestamp"`
	Data              []EventInfo `json:"data"`
}
type EventInfo struct {
	ID           string    `json:"id"`
	SportKey     string    `json:"sport_key"`
	SportTitle   string    `json:"sport_title"`
	CommenceTime time.Time `json:"commence_time"`
	HomeTeam     string    `json:"home_team"`
	AwayTeam     string    `json:"away_team"`
}

func (s *OddsService) GetGamesForDate(date time.Time, config *sports.SportsbookConfig) []EventInfo {
	var games []EventInfo

	endpont := "historical/sports/%s/events/"
	addlArgs := []string{
		"date=" + date.UTC().Format("2006-01-02T15:04:05Z"),
		"commenceTimeFrom=" + date.UTC().Format("2006-01-02T15:04:05Z"),
		"commenceTimeTo=" + date.AddDate(0, 0, 1).UTC().Format("2006-01-02T15:04:05Z"),
	}
	res, err := s.deps.Sources.GetOddsAPI(fmt.Sprintf(endpont, config.LeagueName), addlArgs)
	if err != nil {
		log.Fatal("Error getting odds api: ", err)
	}

	var eventsResponse EventsResponse
	if err := json.Unmarshal([]byte(res), &eventsResponse); err != nil {
		panic(err)
	}
	for _, game := range eventsResponse.Data {
		games = append(games, game)
	}

	return games
}

func (s *OddsService) GetLiveGamesForDate(date time.Time, apiGetter APIGetter) []EventInfo {
	endpont := "sports/%s/events/"
	addlArgs := []string{
		"commenceTimeFrom=" + date.UTC().Format("2006-01-02T15:04:05Z"),
		"commenceTimeTo=" + date.AddDate(0, 0, 1).UTC().Format("2006-01-02T15:04:05Z"),
	}
	if apiGetter == nil {
		apiGetter = s.deps.Sources.GetOddsAPI
	}
	res, err := apiGetter(fmt.Sprintf(endpont, "basketball_nba"), addlArgs)
	if err != nil {
		log.Fatal("Error getting odds api: ", err)
	}

	var events []EventInfo
	if err := json.Unmarshal([]byte(res), &events); err != nil {
		panic(err)
	}

	return events
}

type OddResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	PreviousTimestamp time.Time `json:"previous_timestamp"`
	NextTimestamp     time.Time `json:"next_timestamp"`
	Data              OddsInfo  `json:"data"`
}

type OddsInfo struct {
	ID           string    `json:"id"`
	SportKey     string    `json:"sport_key"`
	SportTitle   string    `json:"sport_title"`
	CommenceTime time.Time `json:"commence_time"`
	HomeTeam     string    `json:"home_team"`
	AwayTeam     string    `json:"away_team"`
	Bookmakers   []struct {
		Key        string    `json:"key"`
		Title      string    `json:"title"`
		LastUpdate time.Time `json:"last_update"`
		Markets    []struct {
			Key        string    `json:"key"`
			LastUpdate time.Time `json:"last_update"`
			Outcomes   []struct {
				Name        string  `json:"name"`
				Description string  `json:"description"`
				Price       int     `json:"price"`
				Point       float32 `json:"point"`
				Link        string  `json:"link"`
			} `json:"outcomes"`
		} `json:"markets"`
	} `json:"bookmakers"`
}

func (s *OddsService) GetOddsForGame(sport sports.Sport, game EventInfo, config *sports.SportsbookConfig) []odds.PlayerLine {
	log.Printf("Getting odds for %s vs %s", game.HomeTeam, game.AwayTeam)
	var lines []odds.PlayerLine
	nameMap := make(map[string]string)

	for _, marketConfig := range config.Markets {
		endpont := "historical/sports/%s/events/%s/odds"
		addlArgs := []string{
			"date=" + game.CommenceTime.UTC().Format("2006-01-02T15:04:05Z"),
			"regions=us",
			"bookmakers=" + marketConfig.Bookmaker,
			"markets=" + strings.Join(marketConfig.Markets, ","),
			"oddsFormat=" + "american",
			"includeLinks=" + "true",
		}
		res, err := s.deps.Sources.GetOddsAPI(fmt.Sprintf(endpont, config.LeagueName, game.ID), addlArgs)
		if err != nil {
			log.Fatal("Error getting odds api: ", err)
		}

		var oddResponse OddResponse
		if err := json.Unmarshal([]byte(res), &oddResponse); err != nil {
			panic(err)
		}

		if len(oddResponse.Data.Bookmakers) == 0 {
			log.Printf("Could not find odds for %s vs %s", game.HomeTeam, game.AwayTeam)
			return lines
		}
		for _, market := range oddResponse.Data.Bookmakers[0].Markets {
			truncated_string := strings.ReplaceAll(market.Key, "_alternate", "")
			stat := config.StatMapping[truncated_string]
			for _, line := range market.Outcomes {
				playerName := strings.Join(strings.Split(line.Description, " ")[:2], " ")
				playerIndex, err := s.deps.Store.PlayerNameToIndex(nameMap, playerName)
				if err != nil {
					log.Printf("Error finding player name: %s", line.Description)
					continue
				}
				line := odds.PlayerLine{
					Sport:       string(sport),
					PlayerIndex: playerIndex,
					Timestamp:   market.LastUpdate,
					Stat:        stat,
					Side:        line.Name,
					Line:        line.Point,
					Type:        getMarketType(market.Key),
					Odds:        line.Price,
					Link:        line.Link,
				}
				lines = append(lines, line)
			}
		}
	}

	return lines
}

func (s *OddsService) GetLiveOddsForGame(game EventInfo, oddsType string, apiGetter APIGetter) []odds.PlayerLine {
	log.Printf("Getting odds for %s vs %s", game.HomeTeam, game.AwayTeam)
	var lines []odds.PlayerLine
	nameMap := make(map[string]string)

	var markets, bookmakers string
	if oddsType == "alternate" {
		bookmakers = "fanduel"
		markets = "player_points_alternate,player_rebounds_alternate,player_assists_alternate,player_threes_alternate"
	} else {
		bookmakers = "williamhill_us"
		markets = "player_points,player_rebounds,player_assists,player_threes"
	}

	if apiGetter == nil {
		apiGetter = s.deps.Sources.GetOddsAPI
	}

	endpont := "sports/%s/events/%s/odds"
	addlArgs := []string{
		"bookmakers=" + bookmakers,
		"markets=" + markets,
		"oddsFormat=" + "american",
		"includeLinks=" + "true",
	}
	res, err := apiGetter(fmt.Sprintf(endpont, "basketball_nba", game.ID), addlArgs)
	if err != nil {
		log.Fatal("Error getting odds api: ", err)
	}

	var OddsInfo OddsInfo
	if err := json.Unmarshal([]byte(res), &OddsInfo); err != nil {
		panic(err)
	}

	if len(OddsInfo.Bookmakers) == 0 {
		log.Printf("Could not find odds for %s vs %s", game.HomeTeam, game.AwayTeam)
		return lines
	}
	for _, market := range OddsInfo.Bookmakers[0].Markets {
		truncated_string := strings.ReplaceAll(market.Key, "_alternate", "")
		stat := odds_markets[truncated_string]
		for _, line := range market.Outcomes {
			playerName := strings.Join(strings.Split(line.Description, " ")[:2], " ")
			playerIndex, err := s.deps.Store.PlayerNameToIndex(nameMap, playerName)
			if err != nil {
				log.Printf("Error finding player name: %s", line.Description)
				continue
			}
			line := odds.PlayerLine{
				Sport:       "nba",
				PlayerIndex: playerIndex,
				Timestamp:   market.LastUpdate,
				Stat:        stat,
				Side:        line.Name,
				Line:        line.Point,
				Type:        getMarketType(market.Key),
				Odds:        line.Price,
				Link:        line.Link,
			}
			lines = append(lines, line)
		}
	}

	return lines
}

func getMarketType(market string) string {
	if strings.Contains(market, "alternate") {
		return "alternate"
	}
	return "mainline"
}

func (s *OddsService) GetOdds(startDate time.Time, endDate time.Time, oddsType string) {
	sportConfig, ok := sports.Configs[sports.NBA]
	if !ok {
		log.Printf("failed to get sportsbook config for %s: unsupported sport", sports.NBA)
		return
	}

	sportsbookConfig := &sportConfig.Sportsbook

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		log.Printf("Getting historical %s sportsbook odds for %v...", oddsType, d)
		var lines []odds.PlayerLine

		games := s.GetGamesForDate(d, sportsbookConfig)
		for _, game := range games {
			lines = append(lines, s.GetOddsForGame(sports.NBA, game, sportsbookConfig)...)
		}

		s.deps.Store.AddPlayerLines(lines)
	}
}

func (s *OddsService) GetHistoricalOddsForSport(sport sports.Sport, startDate time.Time, endDate time.Time) {
	log.Printf("Getting historical %s sportsbook odds...", sport)
	sportConfig, ok := sports.Configs[sport]
	if !ok {
		log.Printf("failed to get sportsbook config for %s: unsupported sport", sport)
		return
	}
	sportsbookConfig := &sportConfig.Sportsbook

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		log.Printf("Getting historical %s sportsbook odds for %v...", sport, d)

		var lines []odds.PlayerLine
		games := s.GetGamesForDate(d, sportsbookConfig)
		for _, game := range games {
			lines = append(lines, s.GetOddsForGame(sport, game, sportsbookConfig)...)
		}

		s.deps.Store.AddPlayerLines(lines)
	}
}

func (s *OddsService) GetLiveOdds(date time.Time, oddsType string) {
	log.Printf("Getting live %s sportsbook odds for %v...", oddsType, date)
	var lines []odds.PlayerLine

	games := s.GetLiveGamesForDate(date, nil)
	for _, game := range games {
		lines = append(lines, s.GetLiveOddsForGame(game, oddsType, nil)...)
	}

	s.deps.Store.AddPlayerLines(lines)
}
