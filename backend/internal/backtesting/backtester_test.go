package backtesting

import (
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type fakeBacktesterDataSource struct {
	getGamesForDateFn         func(sport sports.Sport, date time.Time) ([]games.Game, error)
	getPlayerStatsForGamesFn  func(gameIDs []string) (map[string]players.PlayerAvg, error)
	getAlternateOddsForDateFn func(sport sports.Sport, date time.Time) (map[string]map[string][]odds.PlayerLine, error)
	getPlayersForGameFn       func(gameID int, homeIndex string, playerGameTable string, sortString string) (map[string][]players.Player, error)
	runAnalysisOnGameFn       func(roster []players.PlayerRoster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate bool, storePIP bool) []analysis.Analysis
}

func (f fakeBacktesterDataSource) GetGamesForDate(sport sports.Sport, date time.Time) ([]games.Game, error) {
	return f.getGamesForDateFn(sport, date)
}

func (f fakeBacktesterDataSource) GetPlayerStatsForGames(gameIDs []string) (map[string]players.PlayerAvg, error) {
	return f.getPlayerStatsForGamesFn(gameIDs)
}

func (f fakeBacktesterDataSource) GetAlternatePlayerOddsForDate(sport sports.Sport, date time.Time) (map[string]map[string][]odds.PlayerLine, error) {
	return f.getAlternateOddsForDateFn(sport, date)
}

func (f fakeBacktesterDataSource) GetPlayersForGame(gameID int, homeIndex string, playerGameTable string, sortString string) (map[string][]players.Player, error) {
	return f.getPlayersForGameFn(gameID, homeIndex, playerGameTable, sortString)
}

func (f fakeBacktesterDataSource) RunAnalysisOnGame(roster []players.PlayerRoster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate bool, storePIP bool) []analysis.Analysis {
	return f.runAnalysisOnGameFn(roster, opponents, endDate, forceUpdate, storePIP)
}

func TestCalculateProfit(t *testing.T) {
	if got := calculateProfit(100, -110); got <= 0 {
		t.Fatalf("negative odds should return positive profit, got %v", got)
	}
	if got := calculateProfit(100, 150); got != 150 {
		t.Fatalf("positive odds profit = %v, want 150", got)
	}
}

func TestAddResultWinAndLossAndNil(t *testing.T) {
	res := &BacktestResult{}
	pick := analysis.PropPick{Stat: "points", Side: "Over", BetSize: 100, PlayerLine: odds.PlayerLine{Line: 20.5, Odds: -110}, Analysis: analysis.Analysis{PlayerIndex: "p1"}}

	res.addResult(pick, nil)
	if len(res.Bets) != 0 {
		t.Fatalf("nil result should be skipped")
	}

	res.addResult(pick, players.NBAAvg{NumGames: 1, Points: 25})
	if res.Wins != 1 || res.Losses != 0 {
		t.Fatalf("expected a win, got wins=%d losses=%d", res.Wins, res.Losses)
	}

	res.addResult(analysis.PropPick{Stat: "points", Side: "Under", BetSize: 100, PlayerLine: odds.PlayerLine{Line: 20.5, Odds: -110}, Analysis: analysis.Analysis{PlayerIndex: "p2"}}, players.NBAAvg{NumGames: 1, Points: 25})
	if res.Losses != 1 {
		t.Fatalf("expected one loss, got %d", res.Losses)
	}
}

func TestConvertHelpers(t *testing.T) {
	playersIn := []players.Player{{Index: "a"}, {Index: "b"}}
	rosters := convertPlayerMaptoPlayerRosters(playersIn)
	if len(rosters) != 2 || rosters[0].Status != "Available" {
		t.Fatalf("unexpected rosters conversion: %+v", rosters)
	}

	indexes := convertPlayerstoIndex(playersIn)
	if len(indexes) != 2 || indexes[0] != "a" {
		t.Fatalf("unexpected indexes conversion: %+v", indexes)
	}
}

func TestBacktestDateBranchesWithInjectedDeps(t *testing.T) {
	b := NewBacktester(
		time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
		nil,
		BacktesterDeps{DataSource: fakeBacktesterDataSource{
			getGamesForDateFn: func(sport sports.Sport, date time.Time) ([]games.Game, error) {
				return []games.Game{{Id: 1, HomeIndex: "H", AwayIndex: "A"}}, nil
			},
			getPlayerStatsForGamesFn: func(gameIDs []string) (map[string]players.PlayerAvg, error) {
				return map[string]players.PlayerAvg{"p1": players.NBAAvg{NumGames: 1, Points: 22}}, nil
			},
			getAlternateOddsForDateFn: func(sport sports.Sport, date time.Time) (map[string]map[string][]odds.PlayerLine, error) {
				return map[string]map[string][]odds.PlayerLine{"p1": {"points": {{Id: 1, Side: "Over", Line: 20.5, Odds: 200}}}}, nil
			},
			getPlayersForGameFn: func(gameID int, homeIndex, table, sort string) (map[string][]players.Player, error) {
				arr := []players.Player{{Index: "p1"}, {Index: "p2"}, {Index: "p3"}, {Index: "p4"}, {Index: "p5"}, {Index: "p6"}, {Index: "p7"}, {Index: "p8"}}
				return map[string][]players.Player{"home": arr, "away": arr}, nil
			},
			runAnalysisOnGameFn: func(roster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate, storePIP bool) []analysis.Analysis {
				return []analysis.Analysis{{PlayerIndex: "p1", Prediction: players.NBAAvg{NumGames: 2, Minutes: 30, Points: 25}, Outliers: map[string]float32{"points": 0.2}}}
			},
		}},
	)

	strat := Strategy{PropSelector: analysis.PropSelector{Thresholds: map[string]float32{"points": 0.1}, TresholdType: analysis.Percent, MinOdds: -200, MaxOdds: 500, MaxOver: 10, MaxUnder: 10, BetSize: 100, MinGames: 1, MinMinutes: 1}, BacktestResult: &BacktestResult{}}
	b.Strategies = []Strategy{strat}
	b.backtestDate(b.StartDate)

	if b.Strategies[0].Wins+b.Strategies[0].Losses == 0 {
		t.Fatalf("expected at least one evaluated bet in strategy result")
	}
}
