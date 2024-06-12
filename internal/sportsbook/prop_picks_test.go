package sportsbook

import (
	"reflect"
	"testing"
	"time"
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
    want := []string {"4622c02f9bd1df188631c86e04036049","9b3130e607f80aa4912aa184e2f4eab3"}
    res := GetGamesForDate(startDate, mockPropOddsGames)
    if !reflect.DeepEqual(res, want){
        t.Fatalf(`GetGamesForDate = %q, want match for %#q`, res, want)
    }
}
