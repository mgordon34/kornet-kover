package sportsbook

import (
	"reflect"
	"testing"
	"time"

    // "github.com/mgordon34/kornet-kover/api/odds"
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
    want := []string {"4622c02f9bd1df188631c86e04036049","9b3130e607f80aa4912aa184e2f4eab3"}
    res := GetGamesForDate(startDate, mockPropOddsGames)
    if !reflect.DeepEqual(res, want){
        t.Fatalf(`GetGamesForDate = %q, want match for %q`, res, want)
    }
}

// func TestGetOddsForMarket(t *testing.T) {
//     gameId := "4622c02f9bd1df188631c86e04036049"
//     market := "rebounds"
//     timestamp, _ := time.Parse("2006-01-02", "2023-10-24T22:37:45")
//     want := odds.PlayerOdds {
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
