package sportsbook

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
)

func mockPropOddsGames(endpoint string, addlArgs []string) (response string, err error) {
	response = `{"league":"nba","date":"2023-10-24","games":[{"id":1022300061,"game_id":"4622c02f9bd1df188631c86e04036049",
    "away_team":"Los Angeles Lakers","home_team":"Denver Nuggets","start_timestamp":"2023-10-24T23:30:00Z","participants":[]},
    {"id":1022300062,"game_id":"9b3130e607f80aa4912aa184e2f4eab3","away_team":"Phoenix Suns","home_team":"Golden State Warriors",
    "start_timestamp":"2023-10-25T02:00:00Z","participants":[]}]}`
	return response, nil
}

func TestGetGamesForDate(t *testing.T) {
	startDate, _ := time.Parse("2006-01-02", "2023-10-24")
	t1, _ := time.Parse("2006-01-02T15:04:05", "2023-10-24T23:30:00")
	t2, _ := time.Parse("2006-01-02T15:04:05", "2023-10-25T02:00:00")
	want := []Game{
		{
			ID:        "4622c02f9bd1df188631c86e04036049",
			Timestamp: t1,
		},
		{
			ID:        "9b3130e607f80aa4912aa184e2f4eab3",
			Timestamp: t2,
		},
	}

	svc := NewPropPicksService(PropPicksServiceDeps{})
	res := svc.GetGamesForDate(startDate, mockPropOddsGames)
	if !reflect.DeepEqual(res, want) {
		t.Fatalf(`getgamesfordate = %q, want match for %q`, res, want)
	}
}

func TestPPUpdateLinesUsesInjectedService(t *testing.T) {
	called := false
	svc := NewPropPicksService(PropPicksServiceDeps{
		Store: fakeSportsbookStore{getLastLineFn: func(oddsType string) (odds.PlayerLine, error) {
			return odds.PlayerLine{Timestamp: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
		}},
		Now: func() time.Time { return time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC) },
		RunGetGames: func(startDate, endDate time.Time) {
			called = true
		},
	})

	if err := svc.UpdateLines(); err != nil {
		t.Fatalf("UpdateLines() error = %v", err)
	}
	if !called {
		t.Fatalf("expected injected RunGetGames to be called")
	}

	svcErr := NewPropPicksService(PropPicksServiceDeps{
		Store: fakeSportsbookStore{getLastLineFn: func(oddsType string) (odds.PlayerLine, error) {
			return odds.PlayerLine{}, errors.New("no lines")
		}},
	})
	if err := svcErr.UpdateLines(); err == nil {
		t.Fatalf("expected error when last line lookup fails")
	}
}

func TestGetGamesUsesInjectedDependencies(t *testing.T) {
	responses := map[string]string{
		"/beta/games/nba":                          `{"league":"nba","date":"2099-1-1","games":[{"id":1,"game_id":"g1","away_team":"A","home_team":"B","start_timestamp":"2099-01-01T02:00:00Z","participants":[]}]}`,
		"/beta/odds/g1/player_points_over_under":   `{"game_id":"g1","sportsbooks":[{"bookie_key":"fanduel","market":{"market_key":"player_points_over_under","outcomes":[{"timestamp":"2099-01-01T00:05:00","handicap":20.5,"odds":-110,"participant":0,"participant_name":"","name":"Aaron Gordon Over 20.5","description":"","deep":null}]}}]}`,
		"/beta/odds/g1/player_rebounds_over_under": `{"game_id":"g1","sportsbooks":[{"bookie_key":"fanduel","market":{"market_key":"player_rebounds_over_under","outcomes":[{"timestamp":"2099-01-01T00:05:00","handicap":7.5,"odds":-105,"participant":0,"participant_name":"","name":"Aaron Gordon Over 7.5","description":"","deep":null}]}}]}`,
		"/beta/odds/g1/player_assists_over_under":  `{"game_id":"g1","sportsbooks":[{"bookie_key":"fanduel","market":{"market_key":"player_assists_over_under","outcomes":[{"timestamp":"2099-01-01T00:05:00","handicap":3.5,"odds":120,"participant":0,"participant_name":"","name":"Aaron Gordon Over 3.5","description":"","deep":null}]}}]}`,
	}

	added := 0
	svc := NewPropPicksService(PropPicksServiceDeps{
		Sources: fakeSportsbookSources{getPropOddsFn: func(endpoint string, addlArgs []string) (string, error) {
			if v, ok := responses[endpoint]; ok {
				return v, nil
			}
			return `{"game_id":"","sportsbooks":[]}`, nil
		}},
		Store: fakeSportsbookStore{playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) { return "p1", nil }, addPlayerLinesFn: func(lines []odds.PlayerLine) { added += len(lines) }},
	})

	svc.GetGames(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	if added != len(markets) {
		t.Fatalf("expected one line per market, got %d", added)
	}
}

func TestGetLinesForMarket_ParsesAndFiltersByTime(t *testing.T) {
	svc := NewPropPicksService(PropPicksServiceDeps{
		Store: fakeSportsbookStore{playerNameToIndexFn: func(nameMap map[string]string, playerName string) (string, error) {
			if playerName == "Bad Name" {
				return "", errors.New("missing")
			}
			return "idx1", nil
		}},
	})

	getter := func(endpoint string, addlArgs []string) (string, error) {
		return `{
			"game_id":"g1",
			"sportsbooks":[{
				"bookie_key":"fanduel",
				"market":{
					"market_key":"player_points_over_under",
					"outcomes":[
						{"timestamp":"2099-01-01T00:05:00","handicap":20.5,"odds":-110,"participant":0,"participant_name":"","name":"Aaron Gordon Over 20.5","description":"","deep":null},
						{"timestamp":"2099-01-01T00:10:00","handicap":20.5,"odds":-105,"participant":0,"participant_name":"","name":"Bad Name Over 20.5","description":"","deep":null},
						{"timestamp":"2099-01-01T05:00:00","handicap":21.5,"odds":120,"participant":0,"participant_name":"","name":"Aaron Gordon Under 21.5","description":"","deep":null}
					]
				}
			}]
		}`, nil
	}

	lines := svc.GetLinesForMarket(Game{ID: "g1", Timestamp: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}, "player_points_over_under", "points", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one valid filtered line, got %d", len(lines))
	}
	if lines[0].PlayerIndex != "idx1" || lines[0].Stat != "points" {
		t.Fatalf("unexpected parsed line: %+v", lines[0])
	}
}
