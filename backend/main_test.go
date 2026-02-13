package main

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/backtesting"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

func TestConvertPlayerMaptoPlayerRosters(t *testing.T) {
	in := []players.Player{{Index: "a"}, {Index: "b"}}
	out := convertPlayerMaptoPlayerRosters(in)

	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].PlayerIndex != "a" || out[0].Status != "Available" {
		t.Fatalf("unexpected first roster: %+v", out[0])
	}
}

func TestNewRouterRegistersExpectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouter()
	routes := r.Routes()

	var paths []string
	for _, route := range routes {
		paths = append(paths, route.Path)
	}

	expected := []string{
		"/update-games",
		"/update-players",
		"/update-lines",
		"/pick-props",
		"/strategies",
		"/prop-picks",
		"/prop-picks/bettor",
	}

	for _, path := range expected {
		if !slices.Contains(paths, path) {
			t.Fatalf("expected route %s to be registered; routes=%v", path, paths)
		}
	}
}

func TestMainAndUpdateEntrypointsUseInjectedFns(t *testing.T) {
	origInit := initTablesFn
	origStart := startServerFn
	origRunServer := runServerFn
	origUpdateGames := updateGamesFn
	origUpdateLines := updateLinesFn
	t.Cleanup(func() {
		initTablesFn = origInit
		startServerFn = origStart
		runServerFn = origRunServer
		updateGamesFn = origUpdateGames
		updateLinesFn = origUpdateLines
	})

	initCalled := false
	startCalled := false
	updateGamesCalled := false
	updateLinesCalled := false

	initTablesFn = func() { initCalled = true }
	startServerFn = func() { startCalled = true }
	updateGamesFn = func(s sports.Sport) error {
		if s != sports.NBA {
			t.Fatalf("expected NBA sport")
		}
		updateGamesCalled = true
		return nil
	}
	updateLinesFn = func() error {
		updateLinesCalled = true
		return nil
	}

	main()
	runUpdateGames()
	runUpdateLines()

	if !initCalled || !startCalled || !updateGamesCalled || !updateLinesCalled {
		t.Fatalf("expected all entrypoint seams to be called")
	}

	runCalled := false
	runServerFn = func(r *gin.Engine) { runCalled = true }
	startServer()
	if !runCalled {
		t.Fatalf("expected startServer to call runServerFn")
	}
}

func TestOperationalHelpersUseInjectedDeps(t *testing.T) {
	origMissing := getMLBPlayersMissingHandednessFn
	origScrapeHandedness := scrapeMLBPlayerHandednessFn
	origAddHandedness := addMLBPlayerHandednessFn
	origPips := getPIPPredictionsForDateFn
	origOdds := getOddsForDateFn
	origSportsbook := sportsbookGetOddsFn
	origPerByYear := getPlayerPerByYearFn
	origPerWith := getPlayerPerWithPlayerByYearFn
	origCalc := calculatePIPFactorFn
	t.Cleanup(func() {
		getMLBPlayersMissingHandednessFn = origMissing
		scrapeMLBPlayerHandednessFn = origScrapeHandedness
		addMLBPlayerHandednessFn = origAddHandedness
		getPIPPredictionsForDateFn = origPips
		getOddsForDateFn = origOdds
		sportsbookGetOddsFn = origSportsbook
		getPlayerPerByYearFn = origPerByYear
		getPlayerPerWithPlayerByYearFn = origPerWith
		calculatePIPFactorFn = origCalc
	})

	getMLBPlayersMissingHandednessFn = func() ([]players.Player, error) {
		return []players.Player{{Index: "mlbtest01"}}, nil
	}
	scrapeMLBPlayerHandednessFn = func(playerIndex string) (string, string, error) { return "R", "L", nil }
	addMLBPlayerHandednessFn = func(playerIndex string, bats string, throws string) error { return nil }
	getPIPPredictionsForDateFn = func(date time.Time) ([]players.NBAPIPPrediction, error) {
		return []players.NBAPIPPrediction{{PlayerIndex: "p1"}}, nil
	}
	getOddsForDateFn = func(s sports.Sport, date time.Time) (map[string]map[string]odds.PlayerOdds, error) {
		return map[string]map[string]odds.PlayerOdds{"p1": {}}, nil
	}
	sportsbookGetOddsFn = func(startDate, endDate time.Time, oddsType string) {}
	getPlayerPerByYearFn = func(s sports.Sport, player string, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{2024: players.NBAAvg{NumGames: 1, Minutes: 30, Points: 20}}
	}
	getPlayerPerWithPlayerByYearFn = func(player, defender string, relationship players.Relationship, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{2024: players.NBAAvg{NumGames: 1, Minutes: 31, Points: 22}}
	}
	calculatePIPFactorFn = func(controlMap, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
		return players.NBAAvg{NumGames: 1, Minutes: 1, Points: 1}
	}

	runUpdateMLBPlayerHandedness()
	runGetPIPPredictions()
	runSportsbookGetGames()
	runGetPlayerOdds()
	if len(runGetPlayerOddsForToday()) == 0 {
		t.Fatalf("expected non-empty odds map")
	}
	runGetPlayerPip()
}

func TestBacktestHelpersUseInjectedFns(t *testing.T) {
	origGetGames := getGamesForDateFn
	origGetPlayers := getPlayersForGameFn
	origRunAnalysis := runMLBAnalysisOnGameFn
	origRunBacktester := runBacktesterFn
	t.Cleanup(func() {
		getGamesForDateFn = origGetGames
		getPlayersForGameFn = origGetPlayers
		runMLBAnalysisOnGameFn = origRunAnalysis
		runBacktesterFn = origRunBacktester
	})

	getGamesForDateFn = func(s sports.Sport, date time.Time) ([]games.Game, error) {
		return nil, nil
	}
	getPlayersForGameFn = func(gameID int, homeIndex, table, sort string) (map[string][]players.Player, error) {
		return map[string][]players.Player{"home": {}, "away": {}}, nil
	}
	runMLBAnalysisOnGameFn = func(roster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate, storePIP bool) []analysis.Analysis {
		return nil
	}

	capturedStrategies := 0
	runBacktesterFn = func(b backtesting.Backtester) {
		capturedStrategies = len(b.Strategies)
	}

	backtestMLB()
	runBacktest()

	if capturedStrategies < 10 {
		t.Fatalf("expected runBacktest to build many strategies, got %d", capturedStrategies)
	}
	_ = fmt.Sprintf("%d", capturedStrategies)
}
