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

type OddsAPI struct {
}

func requestOddsAPI(endpoint string, addlArgs []string) (response string, err error) {
    base_url := "https://api.the-odds-api.com/v4/" + endpoint + "?"
    args := []string{
        "apiKey=" + os.Getenv("ODDS_API_KEY"),
    }
    args = append(args, addlArgs...)

    url := base_url + strings.Join(args[:], "&")
    log.Println(url)
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
	Data              []struct {
		ID           string    `json:"id"`
		SportKey     string    `json:"sport_key"`
		SportTitle   string    `json:"sport_title"`
		CommenceTime time.Time `json:"commence_time"`
		HomeTeam     string    `json:"home_team"`
		AwayTeam     string    `json:"away_team"`
	} `json:"data"`
}

func (o OddsAPI) GetOddsAPIGamesForDate(date time.Time, apiGetter APIGetter) []string {
    var games []string

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
        games = append(games, game.ID)
    }

    return games
}

func (o OddsAPI) GetOdds(date time.Time) {
    games := o.GetOddsAPIGamesForDate(date, requestOddsAPI)

    for _, game := range games {
        log.Printf("Game: %s", game)
    }
}
