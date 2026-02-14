package sportsbook

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

func TestGetLiveGamesForDate_UsesInjectedGetter(t *testing.T) {
	called := false
	getter := func(endpoint string, addlArgs []string) (string, error) {
		called = true
		return `[{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-02T00:00:00Z","home_team":"A","away_team":"B"}]`, nil
	}

	games := GetLiveGamesForDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), getter)
	if !called {
		t.Fatalf("expected injected getter to be called")
	}
	if len(games) != 1 || games[0].ID != "g1" {
		t.Fatalf("unexpected games response: %+v", games)
	}
}

func TestGetLiveOddsForGame_ParsesLinesWithInjectedDependencies(t *testing.T) {
	originalResolver := playerNameToIndex
	playerNameToIndex = func(nameMap map[string]string, playerName string) (string, error) {
		if playerName == "Bad Player" {
			return "", errors.New("not found")
		}
		return "idx1", nil
	}
	t.Cleanup(func() { playerNameToIndex = originalResolver })

	getter := func(endpoint string, addlArgs []string) (string, error) {
		return `{
			"id":"g1",
			"bookmakers":[{
				"key":"williamhill_us",
				"markets":[{
					"key":"player_points",
					"last_update":"2026-01-02T00:00:00Z",
					"outcomes":[
						{"name":"Over","description":"Aaron Gordon","price":-110,"point":20.5,"link":"x"},
						{"name":"Under","description":"Bad Player","price":-105,"point":20.5,"link":"x"}
					]
				}]
			}]
		}`, nil
	}

	lines := GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "mainline", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one line after failed player lookup skip, got %d", len(lines))
	}
	if lines[0].Type != "mainline" || lines[0].Stat != "points" || lines[0].PlayerIndex != "idx1" {
		t.Fatalf("unexpected line: %+v", lines[0])
	}
}

func TestGetLiveOddsForGame_AlternateMarketType(t *testing.T) {
	originalResolver := playerNameToIndex
	playerNameToIndex = func(nameMap map[string]string, playerName string) (string, error) {
		return "idx1", nil
	}
	t.Cleanup(func() { playerNameToIndex = originalResolver })

	getter := func(endpoint string, addlArgs []string) (string, error) {
		return `{
			"id":"g1",
			"bookmakers":[{
				"key":"fanduel",
				"markets":[{
					"key":"player_points_alternate",
					"last_update":"2026-01-02T00:00:00Z",
					"outcomes":[
						{"name":"Over","description":"Aaron Gordon","price":130,"point":23.5,"link":"x"}
					]
				}]
			}]
		}`, nil
	}

	lines := GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "alternate", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one alternate line, got %d", len(lines))
	}
	if lines[0].Type != "alternate" || lines[0].Stat != "points" || lines[0].Odds != 130 {
		t.Fatalf("unexpected alternate line: %+v", lines[0])
	}
}

func TestGetLiveOddsForGame_NoBookmakers(t *testing.T) {
	getter := func(endpoint string, addlArgs []string) (string, error) {
		return `{"id":"g1","bookmakers":[]}`, nil
	}

	lines := GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "mainline", getter)
	if len(lines) != 0 {
		t.Fatalf("expected no lines when no bookmakers, got %d", len(lines))
	}
}

func TestUpdateLinesAndHandlerUseInjectedSeams(t *testing.T) {
	origLast := oddsGetLastLineFn
	origRange := getOddsRangeFn
	origLive := getLiveOddsFn
	origRunner := updateLinesRunnerFn
	t.Cleanup(func() {
		oddsGetLastLineFn = origLast
		getOddsRangeFn = origRange
		getLiveOddsFn = origLive
		updateLinesRunnerFn = origRunner
	})

	oddsGetLastLineFn = func(oddsType string) (odds.PlayerLine, error) {
		return odds.PlayerLine{Timestamp: time.Date(2099, 1, 1, 12, 0, 0, 0, time.UTC)}, nil
	}
	count := 0
	getOddsRangeFn = func(startDate, endDate time.Time, oddsType string) { count++ }
	getLiveOddsFn = func(date time.Time, oddsType string) { count++ }

	if err := UpdateLines(); err != nil {
		t.Fatalf("UpdateLines() error = %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 odds update calls, got %d", count)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/update-lines", GetUpdateLines)
	updateLinesRunnerFn = func() error { return nil }
	req := httptest.NewRequest(http.MethodGet, "/update-lines", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	updateLinesRunnerFn = func() error { return errors.New("boom") }
	req2 := httptest.NewRequest(http.MethodGet, "/update-lines", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}

func TestOddsBatchFunctionsUseInjectedFetchers(t *testing.T) {
	origCfg := getSportsbookConfigFn
	origGetGames := getGamesForDateFn
	origGetOdds := getOddsForGameFn
	origGetLiveGames := getLiveGamesForDateFn
	origGetLiveOdds := getLiveOddsForGameFn
	origAddLines := oddsAddPlayerLinesFn
	t.Cleanup(func() {
		getSportsbookConfigFn = origCfg
		getGamesForDateFn = origGetGames
		getOddsForGameFn = origGetOdds
		getLiveGamesForDateFn = origGetLiveGames
		getLiveOddsForGameFn = origGetLiveOdds
		oddsAddPlayerLinesFn = origAddLines
	})

	getSportsbookConfigFn = func(sport sports.Sport) *sports.SportsbookConfig {
		return &sports.SportsbookConfig{LeagueName: "basketball_nba"}
	}
	getGamesForDateFn = func(date time.Time, config *sports.SportsbookConfig) []EventInfo {
		return []EventInfo{{ID: "g1", HomeTeam: "A", AwayTeam: "B"}}
	}
	getOddsForGameFn = func(sport sports.Sport, game EventInfo, config *sports.SportsbookConfig) []odds.PlayerLine {
		return []odds.PlayerLine{{PlayerIndex: "p1", Type: "mainline", Stat: "points", Side: "Over", Line: 20.5, Odds: -110}}
	}
	getLiveGamesForDateFn = func(date time.Time, apiGetter APIGetter) []EventInfo {
		return []EventInfo{{ID: "g1", HomeTeam: "A", AwayTeam: "B"}}
	}
	getLiveOddsForGameFn = func(game EventInfo, oddsType string, apiGetter APIGetter) []odds.PlayerLine {
		return []odds.PlayerLine{{PlayerIndex: "p1", Type: oddsType, Stat: "points", Side: "Over", Line: 20.5, Odds: -110}}
	}

	added := 0
	oddsAddPlayerLinesFn = func(lines []odds.PlayerLine) { added += len(lines) }

	GetOdds(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC), "mainline")
	GetHistoricalOddsForSport(sports.NBA, time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC))
	GetLiveOdds(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), "mainline")

	if added == 0 {
		t.Fatalf("expected odds lines to be aggregated and added")
	}
}

func TestGetGamesForDateAndGetOddsForGame_WithInjectedRequester(t *testing.T) {
	origReq := requestOddsAPIFn
	origResolver := playerNameToIndex
	t.Cleanup(func() {
		requestOddsAPIFn = origReq
		playerNameToIndex = origResolver
	})

	responses := map[string]string{
		"historical/sports/basketball_nba/events/":        `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":[{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B"}]}`,
		"historical/sports/basketball_nba/events/g1/odds": `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B","bookmakers":[{"key":"fanduel","title":"FanDuel","last_update":"2026-01-01T22:00:00Z","markets":[{"key":"player_points","last_update":"2026-01-01T22:00:00Z","outcomes":[{"name":"Over","description":"Aaron Gordon","price":-110,"point":20.5,"link":"x"}]}]}]}}`,
	}

	requestOddsAPIFn = func(endpoint string, addlArgs []string) (string, error) {
		if v, ok := responses[endpoint]; ok {
			return v, nil
		}
		return `{"data":{"bookmakers":[]}}`, nil
	}
	playerNameToIndex = func(nameMap map[string]string, playerName string) (string, error) {
		return "idx1", nil
	}

	config := &sports.SportsbookConfig{
		LeagueName: "basketball_nba",
		StatMapping: map[string]string{
			"player_points": "points",
		},
		Markets: map[string]sports.MarketConfig{"mainline": {Bookmaker: "fanduel", Markets: []string{"player_points"}}},
	}

	games := GetGamesForDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), config)
	if len(games) != 1 || games[0].ID != "g1" {
		t.Fatalf("GetGamesForDate() = %+v", games)
	}

	lines := GetOddsForGame(sports.NBA, games[0], config)
	if len(lines) != 1 || lines[0].PlayerIndex != "idx1" || lines[0].Stat != "points" {
		t.Fatalf("GetOddsForGame() = %+v", lines)
	}
}
