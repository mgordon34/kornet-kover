package picks

import (
	"math"
	"testing"
	"time"
)

func TestFormatPicksByStrat_GroupsAndSorts(t *testing.T) {
	now := time.Now()
	input := []PropPickFormatted{
		{Id: 1, StratId: 2, StratName: "B", Name: "p2", Stat: "points", Threes: float32(math.NaN()), Date: now},
		{Id: 2, StratId: 1, StratName: "A", Name: "p1", Stat: "assists", Date: now},
		{Id: 3, StratId: 2, StratName: "B", Name: "p3", Stat: "rebounds", Date: now},
	}

	out := formatPicksByStrat(input)
	if len(out) != 2 {
		t.Fatalf("group count = %d, want 2", len(out))
	}
	if out[0].StratId != 1 || out[1].StratId != 2 {
		t.Fatalf("unexpected strat sort order: %+v", out)
	}
	if out[1].Picks[0].Threes != 0 {
		t.Fatalf("NaN threes should be normalized to 0")
	}
}

func TestBettorHelpers(t *testing.T) {
	if got := formatBettorLineDisplay("Over", "points", 21.5, -110); got != "Over 21.5 points -110" {
		t.Fatalf("formatBettorLineDisplay() = %q", got)
	}

	if got := getPredictedValue("points", 10, 2, 3, 1); got != 10 {
		t.Fatalf("points predicted = %v, want 10", got)
	}
	if got := getPredictedValue("rebounds", 10, 2, 3, 1); got != 2 {
		t.Fatalf("rebounds predicted = %v, want 2", got)
	}
	if got := getPredictedValue("assists", 10, 2, 3, 1); got != 3 {
		t.Fatalf("assists predicted = %v, want 3", got)
	}
	if got := getPredictedValue("threes", 10, 2, 3, 1); got != 1 {
		t.Fatalf("threes predicted = %v, want 1", got)
	}
	if got := getPredictedValue("unknown", 10, 2, 3, 1); got != 0 {
		t.Fatalf("unknown stat predicted = %v, want 0", got)
	}
}

func TestGroupBettorPicksByStrategy(t *testing.T) {
	team := "DEN"
	rows := []BettorPickRow{
		{ID: 2, StratID: 2, StratName: "B", PlayerName: "p2", TeamName: &team, Side: "Over", Stat: "points", Line: 20.5, Odds: -110, Points: 24},
		{ID: 1, StratID: 1, StratName: "A", PlayerName: "p1", Side: "Under", Stat: "rebounds", Line: 8.5, Odds: -105, Rebounds: 7},
	}

	out := groupBettorPicksByStrategy(rows)
	if len(out) != 2 {
		t.Fatalf("group count = %d, want 2", len(out))
	}
	if out[0].StratID != 1 || out[1].StratID != 2 {
		t.Fatalf("unexpected strat ordering: %+v", out)
	}
	if out[1].Picks[0].Team != "DEN" {
		t.Fatalf("team should be populated")
	}
	if out[0].Picks[0].Predicted != 7 {
		t.Fatalf("predicted rebounds = %v, want 7", out[0].Picks[0].Predicted)
	}
}
