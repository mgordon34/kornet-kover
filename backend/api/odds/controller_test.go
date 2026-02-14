package odds

import "testing"

func TestAddLineToOddsMap_SelectsCloserOdds(t *testing.T) {
	oddsMap := map[string]map[string]PlayerOdds{}

	addLineToOddsMap(oddsMap, PlayerLine{PlayerIndex: "p1", Stat: "points", Side: "Over", Odds: -130, Line: 20.5})
	addLineToOddsMap(oddsMap, PlayerLine{PlayerIndex: "p1", Stat: "points", Side: "Over", Odds: -110, Line: 21.5})
	addLineToOddsMap(oddsMap, PlayerLine{PlayerIndex: "p1", Stat: "points", Side: "Under", Odds: -125, Line: 20.5})

	got := oddsMap["p1"]["points"]
	if got.Over.Odds != -110 {
		t.Fatalf("Over odds = %d, want -110", got.Over.Odds)
	}
	if got.Under.Odds != -125 {
		t.Fatalf("Under odds = %d, want -125", got.Under.Odds)
	}
}

func TestAddAlternateLineToOddsMap_Appends(t *testing.T) {
	oddsMap := map[string]map[string][]PlayerLine{}

	addAlternateLineToOddsMap(oddsMap, PlayerLine{PlayerIndex: "p1", Stat: "rebounds", Side: "Over", Line: 8.5})
	addAlternateLineToOddsMap(oddsMap, PlayerLine{PlayerIndex: "p1", Stat: "rebounds", Side: "Over", Line: 9.5})

	if len(oddsMap["p1"]["rebounds"]) != 2 {
		t.Fatalf("alternate lines len = %d, want 2", len(oddsMap["p1"]["rebounds"]))
	}
}

func TestGetDistanceAndLineCloser(t *testing.T) {
	if !isLineCloser(PlayerLine{}, PlayerLine{Odds: 120}, 0) {
		t.Fatalf("expected empty current line to be replaceable")
	}

	current := PlayerLine{Odds: -130}
	candidate := PlayerLine{Odds: -115}
	if !isLineCloser(current, candidate, 0) {
		t.Fatalf("expected -115 to be closer to target 0 than -130")
	}

	if getDistanceFromTarget(-130, 0) != 30 {
		t.Fatalf("distance for -130 should be 30")
	}
	if getDistanceFromTarget(145, 0) != 45 {
		t.Fatalf("distance for 145 should be 45")
	}
}
