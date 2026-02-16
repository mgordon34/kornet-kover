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

	svc := NewOddsService(OddsServiceDeps{})
	games := svc.GetLiveGamesForDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), getter)
	if !called {
		t.Fatalf("expected injected getter to be called")
	}
	if len(games) != 1 || games[0].ID != "g1" {
		t.Fatalf("unexpected games response: %+v", games)
	}
}

func TestGetLiveOddsForGame_ParsesLinesWithInjectedDependencies(t *testing.T) {
	svc := NewOddsService(OddsServiceDeps{
		Store: fakeSportsbookStore{playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) {
			if playerName == "Bad Player" {
				return "", errors.New("not found")
			}
			return "idx1", nil
		}},
	})

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

	lines := svc.GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "mainline", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one line after failed player lookup skip, got %d", len(lines))
	}
	if lines[0].Type != "mainline" || lines[0].Stat != "points" || lines[0].PlayerIndex != "idx1" {
		t.Fatalf("unexpected line: %+v", lines[0])
	}
}

func TestGetLiveOddsForGame_AlternateMarketType(t *testing.T) {
	svc := NewOddsService(OddsServiceDeps{
		Store: fakeSportsbookStore{playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) {
			return "idx1", nil
		}},
	})

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

	lines := svc.GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "alternate", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one alternate line, got %d", len(lines))
	}
	if lines[0].Type != "alternate" || lines[0].Stat != "points" || lines[0].Odds != 130 {
		t.Fatalf("unexpected alternate line: %+v", lines[0])
	}
}

func TestGetLiveOddsForGame_NoBookmakers(t *testing.T) {
	svc := NewOddsService(OddsServiceDeps{})
	getter := func(endpoint string, addlArgs []string) (string, error) {
		return `{"id":"g1","bookmakers":[]}`, nil
	}

	lines := svc.GetLiveOddsForGame(EventInfo{ID: "g1", HomeTeam: "A", AwayTeam: "B"}, "mainline", getter)
	if len(lines) != 0 {
		t.Fatalf("expected no lines when no bookmakers, got %d", len(lines))
	}
}

func TestUpdateLinesAndHandlersUseInjectedService(t *testing.T) {
	calls := 0
	svc := NewOddsService(OddsServiceDeps{
		Store: fakeSportsbookStore{getLastLineFn: func(oddsType string) (odds.PlayerLine, error) {
			return odds.PlayerLine{Timestamp: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}, nil
		}},
		Now: func() time.Time { return time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC) },
		RunGetOdds: func(startDate time.Time, endDate time.Time, oddsType string) {
			calls++
		},
		RunGetLiveOdds: func(date time.Time, oddsType string) {
			calls++
		},
	})

	if err := svc.UpdateLines(); err != nil {
		t.Fatalf("UpdateLines() error = %v", err)
	}
	if calls != 4 {
		t.Fatalf("expected 4 odds update calls, got %d", calls)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/update-lines", UpdateLinesHandler(svc))
	req := httptest.NewRequest(http.MethodGet, "/update-lines", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	r2 := gin.New()
	r2.GET("/update-lines", UpdateLinesHandler(NewOddsService(OddsServiceDeps{
		Store: fakeSportsbookStore{getLastLineFn: func(oddsType string) (odds.PlayerLine, error) {
			return odds.PlayerLine{}, errors.New("boom")
		}},
	})))
	req2 := httptest.NewRequest(http.MethodGet, "/update-lines", nil)
	rec2 := httptest.NewRecorder()
	r2.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}

func TestOddsBatchFunctionsUseInjectedDependencies(t *testing.T) {
	responses := map[string]string{
		"historical/sports/basketball_nba/events/":        `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":[{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B"}]}`,
		"historical/sports/basketball_nba/events/g1/odds": `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B","bookmakers":[{"key":"williamhill_us","title":"William Hill","last_update":"2026-01-01T22:00:00Z","markets":[{"key":"player_points","last_update":"2026-01-01T22:00:00Z","outcomes":[{"name":"Over","description":"Aaron Gordon","price":-110,"point":20.5,"link":"x"}]}]}]}}`,
		"sports/basketball_nba/events/":                   `[{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-02T00:00:00Z","home_team":"A","away_team":"B"}]`,
		"sports/basketball_nba/events/g1/odds":            `{"id":"g1","bookmakers":[{"key":"williamhill_us","markets":[{"key":"player_points","last_update":"2026-01-02T00:00:00Z","outcomes":[{"name":"Over","description":"Aaron Gordon","price":-110,"point":20.5,"link":"x"}]}]}]}`,
	}

	added := 0
	svc := NewOddsService(OddsServiceDeps{
		Sources: fakeSportsbookSources{getOddsAPIFn: func(endpoint string, addlArgs []string) (string, error) {
			if v, ok := responses[endpoint]; ok {
				return v, nil
			}
			return `{"data":{"bookmakers":[]}}`, nil
		}},
		Store: fakeSportsbookStore{
			playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) { return "idx1", nil },
			addPlayerLinesFn:    func(lines []odds.PlayerLine) { added += len(lines) },
		},
	})

	start := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC)
	svc.GetOdds(start, end, "mainline")
	svc.GetHistoricalOddsForSport(sports.NBA, start, end)
	svc.GetLiveOdds(start, "mainline")

	if added == 0 {
		t.Fatalf("expected odds lines to be aggregated and added")
	}
}

func TestGetGamesForDateAndGetOddsForGame_WithInjectedRequester(t *testing.T) {
	responses := map[string]string{
		"historical/sports/basketball_nba/events/":        `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":[{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B"}]}`,
		"historical/sports/basketball_nba/events/g1/odds": `{"timestamp":"2026-01-01T00:00:00Z","previous_timestamp":"2025-12-31T00:00:00Z","next_timestamp":"2026-01-02T00:00:00Z","data":{"id":"g1","sport_key":"basketball_nba","sport_title":"NBA","commence_time":"2026-01-01T23:00:00Z","home_team":"A","away_team":"B","bookmakers":[{"key":"fanduel","title":"FanDuel","last_update":"2026-01-01T22:00:00Z","markets":[{"key":"player_points","last_update":"2026-01-01T22:00:00Z","outcomes":[{"name":"Over","description":"Aaron Gordon","price":-110,"point":20.5,"link":"x"}]}]}]}}`,
	}

	svc := NewOddsService(OddsServiceDeps{
		Sources: fakeSportsbookSources{getOddsAPIFn: func(endpoint string, addlArgs []string) (string, error) {
			if v, ok := responses[endpoint]; ok {
				return v, nil
			}
			return `{"data":{"bookmakers":[]}}`, nil
		}},
		Store: fakeSportsbookStore{playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) { return "idx1", nil }},
	})

	config := &sports.SportsbookConfig{
		LeagueName: "basketball_nba",
		StatMapping: map[string]string{
			"player_points": "points",
		},
		Markets: map[string]sports.MarketConfig{
			"mainline": {
				Bookmaker: "fanduel",
				Markets:   []string{"player_points"},
			},
		},
	}

	games := svc.GetGamesForDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), config)
	if len(games) != 1 || games[0].ID != "g1" {
		t.Fatalf("GetGamesForDate() = %+v", games)
	}

	lines := svc.GetOddsForGame(sports.NBA, games[0], config)
	if len(lines) != 1 || lines[0].PlayerIndex != "idx1" || lines[0].Stat != "points" {
		t.Fatalf("GetOddsForGame() = %+v", lines)
	}
}
