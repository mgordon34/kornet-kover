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
	"github.com/mgordon34/kornet-kover/api/players"
)

type SportsbookPullType int

const (
    Live SportsbookPullType = iota
    Historical
)

var odds_markets = map[string]string{
    "player_points": "points",
    "player_rebounds": "rebounds",
    "player_assists": "assists",
    "player_threes": "threes",
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

func GetUpdateLines(c *gin.Context) {
    err := UpdateLines()
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, "Done")
}

func UpdateLines() error {
    lastLine, err := odds.GetLastLine()
    if err != nil {
        log.Println(err)
        return err
    }
    log.Printf("Last line: %v", lastLine)

    loc, _ := time.LoadLocation("America/New_York")
    d := lastLine.Timestamp.In(loc)
    t := time.Now().In(loc)
    startDate := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
    GetOdds(startDate, today)
    GetLiveOdds(today)

    return nil
}

type EventsResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	PreviousTimestamp time.Time `json:"previous_timestamp"`
	NextTimestamp     time.Time `json:"next_timestamp"`
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

func GetGamesForDate(date time.Time, apiGetter APIGetter) []EventInfo {
    var games []EventInfo

    endpont := "historical/sports/%s/events/"
    addlArgs := []string {
        "date=" + date.UTC().Format("2006-01-02T15:04:05Z"),
        "commenceTimeFrom=" + date.UTC().Format("2006-01-02T15:04:05Z"),
        "commenceTimeTo=" + date.AddDate(0,0,1).UTC().Format("2006-01-02T15:04:05Z"),
    }
    res, err := requestOddsAPI(fmt.Sprintf(endpont, "basketball_nba"), addlArgs)
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

func GetLiveGamesForDate(date time.Time, apiGetter APIGetter) []EventInfo {
    endpont := "sports/%s/events/"
    addlArgs := []string {
        "commenceTimeFrom=" + date.UTC().Format("2006-01-02T15:04:05Z"),
        "commenceTimeTo=" + date.AddDate(0,0,1).UTC().Format("2006-01-02T15:04:05Z"),
    }
    res, err := requestOddsAPI(fmt.Sprintf(endpont, "basketball_nba"), addlArgs)
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
                Price       int `json:"price"`
                Point       float32 `json:"point"`
                Link        string `json:"link"`
            } `json:"outcomes"`
        } `json:"markets"`
    } `json:"bookmakers"`
}

func GetOddsForGame(game EventInfo, apiGetter APIGetter) []odds.PlayerLine {
    log.Printf("Getting odds for %s vs %s", game.HomeTeam, game.AwayTeam)
    var lines []odds.PlayerLine
    nameMap := make(map[string]string)

    endpont := "historical/sports/%s/events/%s/odds"
    addlArgs := []string {
        "date=" + game.CommenceTime.UTC().Format("2006-01-02T15:04:05Z"),
        "bookmakers=" + "williamhill_us",
        "markets=" + "player_points,player_rebounds,player_assists,player_threes",
        "oddsFormat=" + "american",
        "includeLinks=" + "true",
    }
    res, err := requestOddsAPI(fmt.Sprintf(endpont, "basketball_nba", game.ID), addlArgs)
    log.Println(res)
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
        stat := odds_markets[market.Key]
        for _, line := range market.Outcomes {
            playerName := strings.Join(strings.Split(line.Description, " ")[:2], " ")
            playerIndex, err := players.PlayerNameToIndex(nameMap, playerName)
            if err != nil {
                log.Printf("Error finding player name: %s", line.Description)
                continue
            }
            line := odds.PlayerLine{
                Sport: "nba",
                PlayerIndex: playerIndex,
                Timestamp: market.LastUpdate,
                Stat: stat,
                Side: line.Name,
                Line: line.Point,
				Type: "mainline",
                Odds: line.Price,
                Link: line.Link,
            }
            lines = append(lines, line)
        }
    }

    return lines
}

func GetLiveOddsForGame(game EventInfo, apiGetter APIGetter) []odds.PlayerLine {
    log.Printf("Getting odds for %s vs %s", game.HomeTeam, game.AwayTeam)
    var lines []odds.PlayerLine
    nameMap := make(map[string]string)

    endpont := "sports/%s/events/%s/odds"
    addlArgs := []string {
        "bookmakers=" + "williamhill_us",
        "markets=" + "player_points,player_rebounds,player_assists,player_threes",
        "oddsFormat=" + "american",
        "includeLinks=" + "true",
    }
    res, err := requestOddsAPI(fmt.Sprintf(endpont, "basketball_nba", game.ID), addlArgs)
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
        stat := odds_markets[market.Key]
        for _, line := range market.Outcomes {
            playerName := strings.Join(strings.Split(line.Description, " ")[:2], " ")
            playerIndex, err := players.PlayerNameToIndex(nameMap, playerName)
            if err != nil {
                log.Printf("Error finding player name: %s", line.Description)
                continue
            }
            line := odds.PlayerLine{
                Sport: "nba",
                PlayerIndex: playerIndex,
                Timestamp: market.LastUpdate,
                Stat: stat,
                Side: line.Name,
                Line: line.Point,
				Type: "mainline",
                Odds: line.Price,
                Link: line.Link,
            }
            lines = append(lines, line)
        }
    }

    return lines
}

func GetOdds(startDate time.Time, endDate time.Time) {
    for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
        log.Printf("Getting historical sportsbook odds for %v...", d)
        var lines []odds.PlayerLine

        games := GetGamesForDate(d, requestOddsAPI)
        for _, game := range games {
            lines = append(lines, GetOddsForGame(game, requestOddsAPI)...)
        }

        odds.AddPlayerLines(lines)
    }
}

func GetLiveOdds(date time.Time) {
    log.Printf("Getting live sportsbook odds for %v...", date)
    var lines []odds.PlayerLine

    games := GetLiveGamesForDate(date, requestOddsAPI)
    for _, game := range games {
        lines = append(lines, GetLiveOddsForGame(game, requestOddsAPI)...)
    }

    odds.AddPlayerLines(lines)
}
