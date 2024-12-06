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

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
)

type OddsAPI struct {
}

var odds_markets = map[string]string{
    "player_points": "points",
    "player_rebounds": "rebounds",
    "player_assists": "assists",
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

func (o OddsAPI) GetOddsAPIGamesForDate(date time.Time, apiGetter APIGetter) []EventInfo {
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

type OddResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	PreviousTimestamp time.Time `json:"previous_timestamp"`
	NextTimestamp     time.Time `json:"next_timestamp"`
	Data              struct {
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
				} `json:"outcomes"`
			} `json:"markets"`
		} `json:"bookmakers"`
	} `json:"data"`
}

func (o OddsAPI) GetOddsAPIOddsForGame(game EventInfo, apiGetter APIGetter) []odds.PlayerLine {
    log.Printf("Getting odds for %s vs %s", game.HomeTeam, game.AwayTeam)
    var lines []odds.PlayerLine
    nameMap := make(map[string]string)

    endpont := "historical/sports/%s/events/%s/odds"
    addlArgs := []string {
        "date=" + game.CommenceTime.UTC().Format("2006-01-02T15:04:05Z"),
        "bookmakers=" + "williamhill_us",
        // "regions=" + "us",
        "markets=" + "player_points,player_rebounds,player_assists",
        "oddsFormat=" + "american",
    }
    res, err := requestOddsAPI(fmt.Sprintf(endpont, "basketball_nba", game.ID), addlArgs)
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
                Odds: line.Price,
            }
            lines = append(lines, line)
        }
    }

    return lines
}

func (o OddsAPI) GetOdds(startDate time.Time, endDate time.Time) {
    for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Getting sportsbook odds for %v...", d)
        var lines []odds.PlayerLine

        games := o.GetOddsAPIGamesForDate(d, requestOddsAPI)
        for _, game := range games {
            lines = append(lines, o.GetOddsAPIOddsForGame(game, requestOddsAPI)...)
        }

        odds.AddPlayerLines(lines)
    }
}
