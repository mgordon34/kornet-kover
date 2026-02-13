package backtesting

import (
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
)

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
	pick := analysis.PropPick{
		Stat:    "points",
		Side:    "Over",
		BetSize: 100,
		PlayerLine: odds.PlayerLine{
			Line: 20.5,
			Odds: -110,
		},
		Analysis: analysis.Analysis{PlayerIndex: "p1"},
	}

	res.addResult(pick, nil)
	if len(res.Bets) != 0 {
		t.Fatalf("nil result should be skipped")
	}

	res.addResult(pick, players.NBAAvg{NumGames: 1, Points: 25})
	if res.Wins != 1 || res.Losses != 0 {
		t.Fatalf("expected a win, got wins=%d losses=%d", res.Wins, res.Losses)
	}

	res.addResult(analysis.PropPick{
		Stat: "points", Side: "Under", BetSize: 100,
		PlayerLine: odds.PlayerLine{Line: 20.5, Odds: -110},
		Analysis:   analysis.Analysis{PlayerIndex: "p2"},
	}, players.NBAAvg{NumGames: 1, Points: 25})

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

func TestResultReportingHelpersDoNotPanic(t *testing.T) {
	win := &analysis.PropPick{
		Stat:       "points",
		Side:       "Over",
		Diff:       2.5,
		PDiff:      0.2,
		BetSize:    100,
		Result:     "Win",
		PlayerLine: odds.PlayerLine{Line: 20.5, Odds: -110},
		Analysis:   analysis.Analysis{PlayerIndex: "p1", Prediction: players.NBAAvg{NumGames: 1, Points: 25}},
	}
	loss := &analysis.PropPick{
		Stat:       "rebounds",
		Side:       "Under",
		Diff:       -1.0,
		PDiff:      -0.1,
		BetSize:    100,
		Result:     "Loss",
		PlayerLine: odds.PlayerLine{Line: 8.5, Odds: 250},
		Analysis:   analysis.Analysis{PlayerIndex: "p2", Prediction: players.NBAAvg{NumGames: 1, Rebounds: 10}},
	}

	b := BacktestResult{Bets: []*analysis.PropPick{win, loss}, StartDate: time.Now(), EndDate: time.Now()}
	b.printResults("test")
	b.resultBreakdown()
}
