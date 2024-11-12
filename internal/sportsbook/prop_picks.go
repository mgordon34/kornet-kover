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

        var lines []odds.PlayerLine
        games := GetGamesForDate(d, requestPropOdds)
        for _, game := range games {
            for stat, market := range markets {
                log.Printf("Getting odds for %s for game %s", stat, game.ID)
                lines = append(lines, GetLinesForMarket(game, market, stat, requestPropOdds)...)
            }
        }

        odds.AddPlayerLines(lines)
    }
}

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

type Game struct {
    ID string
    Timestamp time.Time
}

func GetGamesForDate(date time.Time, apiGetter APIGetter) []Game {
    var games []Game

    dateArg := "date=" + fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
    addlArgs := []string {dateArg}

    res, err := apiGetter("/beta/games/nba", addlArgs)
    if err != nil {
        log.Fatalf("Error requesting prop-odds service: %v", err)
    }

    var gamesResponses GamesResponse
    if err := json.Unmarshal([]byte(res), &gamesResponses); err != nil {
        panic(err)
    }
    for _, game := range gamesResponses.Games {
        games = append(games, Game{ID: game.GameID, Timestamp: game.StartTimestamp})
    }

    return games
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

func GetLinesForMarket(game Game, market string, stat string, apiGetter APIGetter) []odds.PlayerLine {
    var lines []odds.PlayerLine

    res, err := apiGetter(fmt.Sprintf("/beta/odds/%s/%s", game.ID, market), nil)
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
            pl := odds.PlayerLine{
                Sport: "nba",
                PlayerIndex: playerIndex,
                Timestamp: timestamp,
                Stat: stat,
                Side: outcome.Name,
                Line: outcome.Handicap,
                Odds: outcome.Odds,
            }
            // Only add lines with timestamps before game start + 20 minutes
            if timestamp.Before(game.Timestamp.Add(time.Minute * 20)){
                log.Println(pl)
                lines = append(lines, pl)
            }
        }
    }

    return lines
}

func parseNameFromDescription(description string) string {
    descriptionSlice := strings.Split(description, " ")
    nameSlice := descriptionSlice[:len(descriptionSlice) - 1]
    return strings.Join(nameSlice, " ")
}
