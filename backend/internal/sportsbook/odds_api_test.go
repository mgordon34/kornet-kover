package sportsbook

import (
	"errors"
	"testing"
	"time"
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
