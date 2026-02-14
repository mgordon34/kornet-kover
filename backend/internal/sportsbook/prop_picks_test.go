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

func mockPropOddsOdds(endpoint string, addlArgs []string) (response string, err error) {
	response = `{"game_id":"4622c02f9bd1df188631c86e04036049","sportsbooks":[{"bookie_key":"pinnacle","market":{
        "market_key":"player_rebounds_over_under","outcomes":[{"timestamp":"2023-10-24T22:37:45","handicap":5.5,"odds":-157,
        "participant":null,"participant_name":null,"name":"Over","description":"Aaron Gordon (Rebounds)","deep":null}]}}]}`
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
	res := PPGetGamesForDate(startDate, mockPropOddsGames)
	if !reflect.DeepEqual(res, want) {
		t.Fatalf(`getgamesfordate = %q, want match for %q`, res, want)
	}
}

func TestPPUpdateLinesUsesInjectedSeams(t *testing.T) {
	origLast := oddsGetLastLinePropFn
	origGetGames := ppGetGamesFn
	t.Cleanup(func() {
		oddsGetLastLinePropFn = origLast
		ppGetGamesFn = origGetGames
	})

	called := false
	oddsGetLastLinePropFn = func(oddsType string) (odds.PlayerLine, error) {
		return odds.PlayerLine{Timestamp: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
	}
	ppGetGamesFn = func(startDate, endDate time.Time) { called = true }

	if err := PPUpdateLines(); err != nil {
		t.Fatalf("PPUpdateLines() error = %v", err)
	}
	if !called {
		t.Fatalf("expected GetGames seam to be called")
	}

	oddsGetLastLinePropFn = func(oddsType string) (odds.PlayerLine, error) {
		return odds.PlayerLine{}, errors.New("no lines")
	}
	if err := PPUpdateLines(); err == nil {
		t.Fatalf("expected error when last line lookup fails")
	}
}

func TestGetGamesUsesInjectedMarketFetchers(t *testing.T) {
	origReq := requestPropOddsFn
	origGetGamesForDate := ppGetGamesForDateFn
	origGetLines := getLinesForMarketFn
	origAdd := oddsAddPlayerLinesPropFn
	t.Cleanup(func() {
		requestPropOddsFn = origReq
		ppGetGamesForDateFn = origGetGamesForDate
		getLinesForMarketFn = origGetLines
		oddsAddPlayerLinesPropFn = origAdd
	})

	requestPropOddsFn = mockPropOddsGames
	ppGetGamesForDateFn = func(date time.Time, apiGetter APIGetter) []Game {
		return []Game{{ID: "g1", Timestamp: date.Add(2 * time.Hour)}}
	}
	getLinesForMarketFn = func(game Game, market string, stat string, apiGetter APIGetter) []odds.PlayerLine {
		return []odds.PlayerLine{{PlayerIndex: "p1", Stat: stat, Side: "Over", Line: 20.5, Odds: -110}}
	}

	added := 0
	oddsAddPlayerLinesPropFn = func(lines []odds.PlayerLine) { added += len(lines) }

	GetGames(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	if added != len(markets) {
		t.Fatalf("expected one line per market, got %d", added)
	}
}

func TestGetLinesForMarket_ParsesAndFiltersByTime(t *testing.T) {
	origResolver := playerNameToIndexPropFn
	t.Cleanup(func() { playerNameToIndexPropFn = origResolver })

	playerNameToIndexPropFn = func(nameMap map[string]string, playerName string) (string, error) {
		if playerName == "Bad Name" {
			return "", errors.New("missing")
		}
		return "idx1", nil
	}

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

	lines := GetLinesForMarket(Game{ID: "g1", Timestamp: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}, "player_points_over_under", "points", getter)
	if len(lines) != 1 {
		t.Fatalf("expected one valid filtered line, got %d", len(lines))
	}
	if lines[0].PlayerIndex != "idx1" || lines[0].Stat != "points" {
		t.Fatalf("unexpected parsed line: %+v", lines[0])
	}
}

// func testgetoddsformarket(t *testing.t) {
//     gameid := "4622c02f9bd1df188631c86e04036049"
//     market := "rebounds"
//     timestamp, _ := time.parse("2006-01-02", "2023-10-24t22:37:45")
//     want := odds.playerOdds {
//         PlayerIndex: "Test",
//         Date: timestamp,
//         Stat: "rebounds",
//         Line: 5.5,
//         OverOdds: -157,
//         UnderOdds: -110,
//     }
//     res := GetOddsForMarket(gameId, market, mockPropOddsOdds)
//     if !reflect.DeepEqual(res, want){
//         t.Fatalf(`GetGamesForDate = %v, want match for %v`, res, want)
//     }
// }
