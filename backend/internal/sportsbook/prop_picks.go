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
)

var markets = map[string]string{
	"points":   "player_points_over_under",
	"rebounds": "player_rebounds_over_under",
	"assists":  "player_assists_over_under",
}

type PropPicksServiceDeps struct {
	Sources     SportsbookSources
	Store       SportsbookStore
	Now         func() time.Time
	RunGetGames func(startDate time.Time, endDate time.Time)
}

type PropPicksService struct {
	deps PropPicksServiceDeps
}

func NewPropPicksService(deps PropPicksServiceDeps) *PropPicksService {
	if deps.Sources == nil {
		deps.Sources = defaultSportsbookSources{}
	}
	if deps.Store == nil {
		deps.Store = defaultSportsbookStore{}
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}

	svc := &PropPicksService{deps: deps}

	if deps.RunGetGames != nil {
		svc.deps.RunGetGames = deps.RunGetGames
	} else {
		svc.deps.RunGetGames = svc.GetGames
	}

	return svc
}

func requestPropOdds(endpoint string, addlArgs []string) (response string, err error) {
	baseURL := "https://api.prop-odds.com" + endpoint + "?"
	args := []string{
		"api_key=" + os.Getenv("PROP_PICKS_KEY"),
		"tz=" + "America/New_York",
	}
	args = append(args, addlArgs...)

	res, err := http.Get(baseURL + strings.Join(args[:], "&"))
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

func (s *PropPicksService) UpdateLines() error {
	lastLine, err := s.deps.Store.GetLastLine("mainline")
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Last line: %v", lastLine)

	startDate := lastLine.Timestamp
	endDate := s.deps.Now()
	s.deps.RunGetGames(startDate, endDate)

	return nil
}

func (s *PropPicksService) GetGames(startDate time.Time, endDate time.Time) {
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		log.Printf("Scraping games for date: %v", d)

		var lines []odds.PlayerLine
		games := s.GetGamesForDate(d, s.deps.Sources.GetPropOdds)
		for _, game := range games {
			for stat, market := range markets {
				log.Printf("Getting odds for %s for game %s", stat, game.ID)
				lines = append(lines, s.GetLinesForMarket(game, market, stat, s.deps.Sources.GetPropOdds)...)
			}
		}

		s.deps.Store.AddPlayerLines(lines)
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
	ID        string
	Timestamp time.Time
}

func (s *PropPicksService) GetGamesForDate(date time.Time, apiGetter APIGetter) []Game {
	var games []Game

	dateArg := "date=" + fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
	addlArgs := []string{dateArg}

	if apiGetter == nil {
		apiGetter = s.deps.Sources.GetPropOdds
	}

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

func (s *PropPicksService) GetLinesForMarket(game Game, market string, stat string, apiGetter APIGetter) []odds.PlayerLine {
	var lines []odds.PlayerLine
	nameMap := make(map[string]string)

	if apiGetter == nil {
		apiGetter = s.deps.Sources.GetPropOdds
	}

	res, err := apiGetter(fmt.Sprintf("/beta/odds/%s/%s", game.ID, market), nil)
	if err != nil {
		log.Fatalf("Error requesting prop-odds service: %v", err)
	}

	var oddsResponse OddsResponse
	if err := json.Unmarshal([]byte(res), &oddsResponse); err != nil {
		panic(err)
	}
	for _, bookie := range oddsResponse.Sportsbooks {
		if bookie.BookieKey == "fanduel" {
			for _, outcome := range bookie.Market.Outcomes {
				nameSplit := strings.Split(outcome.Name, " ")
				playerName := strings.Join(nameSplit[:len(nameSplit)-2], " ")
				side := nameSplit[len(nameSplit)-2]
				playerIndex, err := s.deps.Store.PlayerNameToIndex(nameMap, playerName)
				if err != nil {
					log.Printf("Error finding player name: %s", playerName)
					continue
				}
				timestamp, err := time.Parse("2006-01-02T15:04:05", outcome.Timestamp)
				if err != nil {
					log.Fatalf("Error parsing timestamp: %v", err)
				}
				pl := odds.PlayerLine{
					Sport:       "nba",
					PlayerIndex: playerIndex,
					Timestamp:   timestamp,
					Stat:        stat,
					Side:        side,
					Line:        outcome.Handicap,
					Odds:        outcome.Odds,
				}

				if timestamp.Before(game.Timestamp.Add(time.Minute * 20)) {
					log.Println(pl)
					lines = append(lines, pl)
				}
			}
		}
	}

	return lines
}

func parseNameFromDescription(description string) string {
	log.Printf("description name: %v", description)
	descriptionSlice := strings.Split(description, " ")
	nameSlice := descriptionSlice[:len(descriptionSlice)-1]
	return strings.Join(nameSlice, " ")
}
