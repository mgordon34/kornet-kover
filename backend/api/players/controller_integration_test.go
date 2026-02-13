//go:build integration
// +build integration

package players

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestPlayerNameToIndex(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration-style test")
	}
	storage.InitTables()

	nameMap := map[string]string{}
	playerName := "Aaron Gordon"
	want := "gordoaa01"
	index, err := PlayerNameToIndex(nameMap, playerName)
	if err != nil {
		t.Fatalf(`PlayerNameToIndex resulted in err: %v`, err)
	}
	if index != want {
		t.Fatalf(`PlayerNameToIndex = %s, want match for %s`, index, want)
	}
}

func TestPlayerNameToIndexWithBadName(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration-style test")
	}
	storage.InitTables()

	nameMap := map[string]string{}
	badName := "Aaron Gordo"
	index, err := PlayerNameToIndex(nameMap, badName)
	if err == nil {
		t.Fatalf(`PlayerNameToIndex incorrectly found result: %v`, index)
	}
}

func TestPlayerNameToIndex_SuffixTolerantLookup(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration-style test")
	}
	storage.InitTables()

	db := storage.GetDB()
	suffix := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	fullName := "Player " + suffix + " Jr."
	shortName := "Player " + suffix
	index := "sfx" + suffix

	_, err := db.Exec(context.Background(), `INSERT INTO players (index, sport, name) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, index, "nba", fullName)
	if err != nil {
		t.Fatalf("failed to insert suffix player: %v", err)
	}

	got, err := PlayerNameToIndex(map[string]string{}, shortName)
	if err != nil {
		t.Fatalf("PlayerNameToIndex() suffix lookup error = %v", err)
	}
	if got != index {
		t.Fatalf("PlayerNameToIndex() = %q, want %q", got, index)
	}
}

func TestPlayerControllerDatabaseFlows(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration-style test")
	}
	storage.InitTables()

	suffix := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	home := "H" + suffix
	away := "A" + suffix
	mlbHome := "MLB_H" + suffix
	mlbAway := "MLB_A" + suffix
	teams.AddTeams([]teams.Team{{Index: home, Name: "Home"}, {Index: away, Name: "Away"}, {Index: mlbHome, Name: "MLB Home"}, {Index: mlbAway, Name: "MLB Away"}})

	nbaDate := time.Date(2099, 4, 1, 0, 0, 0, 0, time.UTC)
	g1, err := games.AddGame(games.Game{Sport: "nba", HomeIndex: home, AwayIndex: away, HomeScore: 100, AwayScore: 90, Date: nbaDate})
	if err != nil {
		t.Fatalf("AddGame nba g1 error = %v", err)
	}
	g2, err := games.AddGame(games.Game{Sport: "nba", HomeIndex: home, AwayIndex: away, HomeScore: 110, AwayScore: 95, Date: nbaDate.AddDate(0, 0, 1)})
	if err != nil {
		t.Fatalf("AddGame nba g2 error = %v", err)
	}

	p1 := "nbaa" + suffix
	p2 := "nbab" + suffix
	p3 := "nbac" + suffix
	AddPlayers([]Player{{Index: p1, Sport: "nba", Name: "NBA A " + suffix}, {Index: p2, Sport: "nba", Name: "NBA B " + suffix}, {Index: p3, Sport: "nba", Name: "NBA C " + suffix}})

	AddPlayerGames([]PlayerGame{
		{PlayerIndex: p1, Game: g1, TeamIndex: home, Minutes: 30, Points: 20, Rebounds: 8, Assists: 6, Threes: 2, Usg: 24, Ortg: 110, Drtg: 106},
		{PlayerIndex: p1, Game: g2, TeamIndex: home, Minutes: 32, Points: 24, Rebounds: 9, Assists: 7, Threes: 3, Usg: 25, Ortg: 112, Drtg: 105},
		{PlayerIndex: p2, Game: g1, TeamIndex: home, Minutes: 28, Points: 15, Rebounds: 5, Assists: 4, Threes: 1, Usg: 20, Ortg: 108, Drtg: 107},
		{PlayerIndex: p3, Game: g1, TeamIndex: away, Minutes: 31, Points: 18, Rebounds: 7, Assists: 5, Threes: 2, Usg: 22, Ortg: 109, Drtg: 108},
	})

	gotPlayer, err := GetPlayer(p1)
	if err != nil || gotPlayer.Index != p1 {
		t.Fatalf("GetPlayer() got=%+v err=%v", gotPlayer, err)
	}

	stats, err := GetPlayerStats(p1, nbaDate, nbaDate.AddDate(0, 0, 2))
	if err != nil || !stats.IsValid() {
		t.Fatalf("GetPlayerStats() stats=%+v err=%v", stats, err)
	}

	playerMap, err := GetPlayersForGame(g1, home, "nba_player_games", "minutes")
	if err != nil || len(playerMap["home"]) == 0 || len(playerMap["away"]) == 0 {
		t.Fatalf("GetPlayersForGame() map=%+v err=%v", playerMap, err)
	}

	statMap, err := GetPlayerStatsForGames([]string{fmt.Sprintf("%d", g1)})
	if err != nil || len(statMap) == 0 {
		t.Fatalf("GetPlayerStatsForGames() len=%d err=%v", len(statMap), err)
	}

	opp, err := GetPlayerStatsWithPlayer(p1, p3, Opponent, nbaDate, nbaDate.AddDate(0, 0, 2))
	if err != nil || !opp.IsValid() {
		t.Fatalf("GetPlayerStatsWithPlayer opponent err=%v stats=%+v", err, opp)
	}

	teamMateLike, err := GetPlayerStatsWithPlayer(p1, "missing"+suffix, Teammate, nbaDate, nbaDate.AddDate(0, 0, 2))
	if err != nil || !teamMateLike.IsValid() {
		t.Fatalf("GetPlayerStatsWithPlayer teammate-filter err=%v stats=%+v", err, teamMateLike)
	}

	if len(GetPlayerPerByYear(sports.NBA, p1, nbaDate, nbaDate.AddDate(0, 0, 2))) == 0 {
		t.Fatalf("GetPlayerPerByYear() should return at least one year")
	}
	if len(GetPlayerPerWithPlayerByYear(p1, p3, Opponent, nbaDate, nbaDate.AddDate(0, 0, 2))) == 0 {
		t.Fatalf("GetPlayerPerWithPlayerByYear() should return at least one year")
	}

	factor := CalculatePIPFactor(
		map[int]PlayerAvg{2099: NBAAvg{NumGames: 1, Minutes: 30, Points: 20}},
		map[int]PlayerAvg{2099: NBAAvg{NumGames: 1, Minutes: 33, Points: 24}},
	)
	if factor == nil {
		t.Fatalf("CalculatePIPFactor() returned nil")
	}

	AddPIPPrediction([]NBAPIPPrediction{{PlayerIndex: p1, Date: nbaDate, Version: CurrNBAPIPPredVersion(), NumGames: 5, Minutes: 31, Points: 23, Rebounds: 8, Assists: 6, Threes: 2, Usg: 22, Ortg: 111, Drtg: 107}})
	preds, err := GetPIPPredictionsForDate(nbaDate)
	if err != nil || len(preds) == 0 {
		t.Fatalf("GetPIPPredictionsForDate() len=%d err=%v", len(preds), err)
	}
	pred, err := GetPlayerPIPPrediction(p1, nbaDate)
	if err != nil || pred.PlayerIndex != p1 {
		t.Fatalf("GetPlayerPIPPrediction() pred=%+v err=%v", pred, err)
	}
	if created := GetOrCreatePrediction(p1, nbaDate); created.GetStats()["points"] == 0 {
		t.Fatalf("GetOrCreatePrediction() should return populated stats")
	}

	UpdatePlayerTables("new" + suffix)
	if _, err := GetPlayer("new" + suffix); err != nil {
		t.Fatalf("UpdatePlayerTables() should insert missing player: %v", err)
	}

	err = UpdateRosters([]PlayerRoster{{Sport: "nba", PlayerIndex: p1, TeamIndex: home, Status: "Available", AvgMins: 30}})
	if err != nil {
		t.Fatalf("UpdateRosters() error = %v", err)
	}
	rosters, err := GetActiveRosters()
	if err != nil || len(rosters[home]) == 0 {
		t.Fatalf("GetActiveRosters() err=%v rosters=%+v", err, rosters)
	}

	mlbDate := time.Date(2099, 5, 1, 0, 0, 0, 0, time.UTC)
	mlbGame, err := games.AddGame(games.Game{Sport: "mlb", HomeIndex: mlbHome, AwayIndex: mlbAway, HomeScore: 6, AwayScore: 3, Date: mlbDate})
	if err != nil {
		t.Fatalf("AddGame mlb error = %v", err)
	}
	batter := "mlbb" + suffix
	pitcher := "mlbp" + suffix
	AddPlayers([]Player{{Index: batter, Sport: "mlb", Name: "MLB Batter " + suffix}, {Index: pitcher, Sport: "mlb", Name: "MLB Pitcher " + suffix}})

	AddMLBPlayerGamesBatting([]MLBPlayerGameBatting{{PlayerIndex: batter, Game: mlbGame, TeamIndex: mlbAway, AtBats: 4, Runs: 1, Hits: 2, RBIs: 2, HomeRuns: 1, Walks: 1, Strikeouts: 1, PAs: 5, Pitches: 20, Strikes: 13, BA: 0.3, OBP: 0.4, SLG: 0.5, OPS: 0.9, WPA: 0.2, Details: "HR"}})
	AddMLBPlayerGamesPitching([]MLBPlayerGamePitching{{PlayerIndex: pitcher, Game: mlbGame, TeamIndex: mlbHome, Innings: 6.0, Hits: 5, Runs: 2, EarnedRuns: 2, Walks: 1, Strikeouts: 7, HomeRuns: 1, ERA: 3.0, BattersFaced: 24, WPA: 0.1}})
	AddMLBPlayByPlays([]MLBPlayByPlay{{BatterIndex: batter, PitcherIndex: pitcher, Game: mlbGame, Inning: 1, Outs: 1, Appearance: 1, Pitches: 4, Result: "HR", RawResult: "home run"}})

	mlbStats, err := GetMLBStats(batter, mlbDate, mlbDate.AddDate(0, 0, 1))
	if err != nil || !mlbStats.IsValid() {
		t.Fatalf("GetMLBStats() stats=%+v err=%v", mlbStats, err)
	}
	vsPitcher, err := GetMLBPlayerStatsWithPlayer(batter, pitcher, mlbDate, mlbDate.AddDate(0, 0, 1))
	if err != nil || vsPitcher.PAs == 0 {
		t.Fatalf("GetMLBPlayerStatsWithPlayer() stats=%+v err=%v", vsPitcher, err)
	}
	bMap, err := GetMLBBattingStatsForGames([]string{fmt.Sprintf("%d", mlbGame)})
	if err != nil || len(bMap) == 0 {
		t.Fatalf("GetMLBBattingStatsForGames() len=%d err=%v", len(bMap), err)
	}
	if len(GetMLBPlayerPerWithPlayerByYear(batter, pitcher, mlbDate, mlbDate.AddDate(0, 0, 1))) == 0 {
		t.Fatalf("GetMLBPlayerPerWithPlayerByYear() should return at least one year")
	}

	missing, err := GetMLBPlayersMissingHandedness()
	if err != nil {
		t.Fatalf("GetMLBPlayersMissingHandedness() err=%v", err)
	}
	if len(missing) == 0 {
		t.Fatalf("expected at least one MLB player missing handedness")
	}

	if err := AddMLBPlayerHandedness(batter, "R", "R"); err != nil {
		t.Fatalf("AddMLBPlayerHandedness() err=%v", err)
	}
}
