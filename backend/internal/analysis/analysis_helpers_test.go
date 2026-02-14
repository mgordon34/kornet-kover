package analysis

import (
	"testing"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
)

func TestPrunePlayers(t *testing.T) {
	in := []players.PlayerRoster{
		{PlayerIndex: "a", Status: "Available", AvgMins: 20},
		{PlayerIndex: "b", Status: "Out", AvgMins: 30},
		{PlayerIndex: "c", Status: "Available", AvgMins: 8},
	}
	out := prunePlayers(in)
	if len(out) != 1 || out[0] != "a" {
		t.Fatalf("prunePlayers() = %+v, want [a]", out)
	}
}

func TestOutliersAndHasOutlier(t *testing.T) {
	base := players.NBAAvg{NumGames: 1, Minutes: 20, Points: 10, Rebounds: 5, Assists: 3, Threes: 1, Usg: 10, Ortg: 100, Drtg: 100}
	pred := players.NBAAvg{NumGames: 1, Minutes: 20, Points: 13, Rebounds: 4, Assists: 2, Threes: 1.2, Usg: 10, Ortg: 100, Drtg: 100}
	outliers := GetOutliers(base, pred)

	a := Analysis{Outliers: outliers}
	if !a.HasOutlier("points", "Over") {
		t.Fatalf("expected points over outlier")
	}
	if a.HasOutlier("points", "Under") {
		t.Fatalf("did not expect points under outlier")
	}
	if a.HasOutlier("unknown", "Over") {
		t.Fatalf("unknown stat should not be outlier")
	}
}

func TestOddsDiffHelpers(t *testing.T) {
	pOdds := odds.PlayerOdds{
		Over:  odds.PlayerLine{Line: 20, Side: "Over"},
		Under: odds.PlayerLine{Line: 20, Side: "Under"},
	}

	diff, pDiff := GetOddsDiff(pOdds, 25)
	if diff <= 0 || pDiff <= 0 {
		t.Fatalf("expected positive over diff, got %v %v", diff, pDiff)
	}

	diff2, _ := GetNewOddsDiff(odds.PlayerLine{Line: 10, Side: "Under"}, 8)
	if diff2 <= 0 {
		t.Fatalf("expected positive under diff, got %v", diff2)
	}

	diff3, pDiff3 := GetBaseOddsDiff(10, 12)
	if diff3 != 2 || pDiff3 != 0.2 {
		t.Fatalf("GetBaseOddsDiff() = (%v, %v), want (2, 0.2)", diff3, pDiff3)
	}
}

func TestIsPickEligible(t *testing.T) {
	selector := PropSelector{
		Thresholds:   map[string]float32{"points": 0.1},
		TresholdType: Percent,
		MinOdds:      -120,
		MaxOdds:      200,
		MinLine:      1,
		MaxLine:      30,
		MinDiff:      1,
		MinMinutes:   10,
		MinGames:     1,
	}

	pick := PropPick{
		Stat:       "points",
		Side:       "Over",
		Diff:       2,
		PDiff:      0.2,
		PlayerOdds: odds.PlayerOdds{Over: odds.PlayerLine{Line: 20, Odds: -110, Side: "Over"}},
		Analysis:   Analysis{Prediction: players.NBAAvg{NumGames: 2, Minutes: 30, Points: 25}},
	}

	if !selector.isPickElligible(pick) {
		t.Fatalf("pick should be eligible")
	}

	selector.RequireOutlier = true
	if selector.isPickElligible(pick) {
		t.Fatalf("pick without outlier should be ineligible when outlier required")
	}
}
