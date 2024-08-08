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

var markets = map[string]string{
    "points": "player_points_over_under",
    "rebounds": "player_rebounds_over_under",
    "assists": "player_assists_over_under",
}

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

type OddsResponse struct {
	GameID      string `json:"game_id"`
	Sportsbooks []struct {
		BookieKey string `json:"bookie_key"`
		Market    struct {
			MarketKey string `json:"market_key"`
			Outcomes  []struct {
				Timestamp       string  `json:"timestamp"`
				Handicap        float32 `json:"handicap"`
				Odds            int     `json:"odds"`
				Participant     int     `json:"participant"`
				ParticipantName string  `json:"participant_name"`
				Name            string  `json:"name"`
				Description     string  `json:"description"`
				Deep            any     `json:"deep"`
			} `json:"outcomes"`
		} `json:"market"`
	} `json:"sportsbooks"`
}

type PlayerLine struct {
    Side string
    Line float32
    Odds int
    Timestamp time.Time
}

func requestPropOdds(endpoint string, addlArgs []string) (response string, err error) {
    base_url := "https://api.prop-odds.com" + endpoint + "?"
    args := []string{
        "api_key=" + os.Getenv("PROP_PICKS_KEY"),
        "tz=" + "America/New_York",
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

func GetGames(startDate time.Time, endDate time.Time) {
    for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Scraping games for date: %v", d)

        var odds []odds.PlayerOdds
        gameIds := GetGamesForDate(d, requestPropOdds)
        for _, gameId := range gameIds {
            for stat, market := range markets {
                log.Printf("Getting odds for %s for game %s", stat, gameId)
                odds = append(odds, GetOddsForMarket(gameId, market, stat, requestPropOdds)...)
            }
        }
    }
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

func GetOddsForMarket(gameId string, market string, stat string, apiGetter APIGetter) []odds.PlayerOdds {
    // var oddsMap map[string]odds.PlayerOdds

    res, err := apiGetter(fmt.Sprintf("/beta/odds/%s/%s", gameId, market), nil)
    if err != nil {
        log.Fatalf("Error requesting prop-odds service: %v", err)
    }

    var oddsResponse OddsResponse
    if err := json.Unmarshal([]byte(res), &oddsResponse); err != nil {
        panic(err)
    }
    for _, bookie := range oddsResponse.Sportsbooks {
        if bookie.BookieKey != "pinnacle" {
            continue
        }
        for _, outcome := range bookie.Market.Outcomes {
            playerName := parseNameFromDescription(outcome.Description)
            playerIndex, err := players.PlayerNameToIndex(playerName)
            if err != nil {
                log.Printf("Error getting player index from name %s: %v", playerName, err)
            }
            timestamp, err := time.Parse("2006-01-02T15:04:05", outcome.Timestamp)
            if err != nil {
                log.Fatalf("Error parsing timestamp: %v", err)
            }
            side := outcome.Name
            line := outcome.Handicap
            odds := outcome.Odds
            po := odds.PlayerOdds{
                PlayerIndex: playerIndex,
                Date: timestamp,
                Stat: stat,
                Line: line,
            }
            log.Printf("PlayerOdds: %v", po)
            log.Printf("%s[%v]: %s %f at %d", playerIndex, timestamp, side, line, odds)
        }
    }

    return nil
}

func parseNameFromDescription(description string) string {
    descriptionSlice := strings.Split(description, " ")
    nameSlice := descriptionSlice[:len(descriptionSlice) - 1]
    return strings.Join(nameSlice, " ")
}
