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
)

type APIGetter func(url string, addlArgs []string) (response string, err error)

type GamesResponse struct {
	League string `json:"league"`
	Date   string `json:"date"`
	Games  []struct {
		ID             int       `json:"id"`
		GameID         string    `json:"game_id"`
		AwayTeam       string    `json:"away_team"`
		HomeTeam       string    `json:"home_team"`
		StartTimestamp time.Time `json:"start_timestamp"`
		Participants   []any     `json:"participants"`
	} `json:"games"`
}

func requestPropOdds(endpoint string, addlArgs []string) (response string, err error) {
    base_url := "https://api.prop-odds.com" + endpoint + "?"
    args := []string{
        "api_key=" + os.Getenv("PROP_PICKS_KEY"),
        "tz=" + "America/New_York",
    }
    args = append(args, addlArgs...)
    log.Printf("url: %s", base_url + strings.Join(args[:], "&"))

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

    log.Printf("Body: %v", buf.String())
    log.Printf("Request: %v", res.Request.URL)

    return buf.String(), err
}

func GetGames(startDate time.Time, endDate time.Time) int {
    var gameIds []string
    for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Scraping games for date: %v", d)

        gameIds = append(gameIds, GetGamesForDate(d, requestPropOdds)...)
        for _, gameId := range gameIds {
            log.Printf("gameId: %s", gameId)
        }
    }

    return len(gameIds)
}

func GetGamesForDate(date time.Time, apiGetter APIGetter) []string {
    var gameIds []string

    dateArg := "date=" + fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
    addlArgs := []string {dateArg}

    res, err := apiGetter("/beta/games/nba", addlArgs)
    if err != nil {
        log.Fatalf("Error requesting prop-odds service: %v", err)
    }

    var games GamesResponse
    if err := json.Unmarshal([]byte(res), &games); err != nil {
        panic(err)
    }
    for _, game := range games.Games {
        gameIds = append(gameIds, game.GameID)
    }

    return gameIds
}
