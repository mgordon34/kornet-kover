package players

import "testing"

func TestGetStatPchange(t *testing.T) {
	if got := getStatPchange(10, 12); got != 0.2 {
		t.Fatalf("getStatPchange() = %v, want 0.2", got)
	}
}

func TestNBAAvgOperations(t *testing.T) {
	base := NBAAvg{NumGames: 2, Minutes: 30, Points: 20, Rebounds: 10, Assists: 5, Threes: 2, Usg: 25, Ortg: 110, Drtg: 105}
	other := NBAAvg{NumGames: 2, Minutes: 20, Points: 10, Rebounds: 8, Assists: 4, Threes: 1, Usg: 20, Ortg: 108, Drtg: 106}

	if !base.IsValid() {
		t.Fatalf("base should be valid")
	}
	if (NBAAvg{}).IsValid() {
		t.Fatalf("zero avg should be invalid")
	}

	stats := base.GetStats()
	if stats["points"] != 20 {
		t.Fatalf("GetStats points = %v", stats["points"])
	}

	added := base.AddAvg(other).(NBAAvg)
	if added.NumGames != 4 {
		t.Fatalf("AddAvg NumGames = %d, want 4", added.NumGames)
	}

	unchanged := base.AddAvg(NBAAvg{}).(NBAAvg)
	if unchanged.NumGames != base.NumGames {
		t.Fatalf("AddAvg invalid should keep base")
	}

	cmp := other.CompareAvg(base).(NBAAvg)
	if cmp.Points >= 0 {
		t.Fatalf("expected points comparison to be negative")
	}

	per := base.ConvertToPer().(NBAAvg)
	if per.Points != base.Points/base.Minutes {
		t.Fatalf("ConvertToPer points mismatch")
	}

	statsAgain := per.ConvertToStats().(NBAAvg)
	if statsAgain.Points != per.Points*per.Minutes {
		t.Fatalf("ConvertToStats points mismatch")
	}

	pred := base.PredictStats(NBAAvg{NumGames: 1, Minutes: 0.1, Points: 0.2}).(NBAAvg)
	if pred.Minutes <= base.Minutes {
		t.Fatalf("PredictStats should increase minutes with positive factor")
	}
}

func TestMLBAvgOperations(t *testing.T) {
	base := MLBBattingAvg{NumGames: 2, PAs: 8, AtBats: 6, Runs: 2, Hits: 3, RBIs: 2, HomeRuns: 1, Walks: 1, Strikeouts: 2, Pitches: 30, Strikes: 20, OBP: 0.4, SLG: 0.5, OPS: 0.9, WPA: 0.2}
	other := MLBBattingAvg{NumGames: 2, PAs: 10, AtBats: 7, Runs: 1, Hits: 2, RBIs: 1, HomeRuns: 0, Walks: 2, Strikeouts: 3, Pitches: 40, Strikes: 25, OBP: 0.3, SLG: 0.4, OPS: 0.7, WPA: 0.1}

	if !base.IsValid() {
		t.Fatalf("base should be valid")
	}
	if (MLBBattingAvg{}).IsValid() {
		t.Fatalf("zero avg should be invalid")
	}

	stats := base.GetStats()
	if stats["hits"] != 3 {
		t.Fatalf("GetStats hits = %v", stats["hits"])
	}

	added := base.AddAvg(other).(MLBBattingAvg)
	if added.NumGames != 4 {
		t.Fatalf("AddAvg NumGames = %d, want 4", added.NumGames)
	}

	cmp := other.CompareAvg(base).(MLBBattingAvg)
	if cmp.Hits >= 0 {
		t.Fatalf("expected hit comparison to be negative")
	}

	per := base.ConvertToPer().(MLBBattingAvg)
	if per.Hits != base.Hits/base.PAs {
		t.Fatalf("ConvertToPer hits mismatch")
	}

	raw := per.ConvertToStats().(MLBBattingAvg)
	if raw.Hits != per.Hits*per.PAs {
		t.Fatalf("ConvertToStats hits mismatch")
	}

	pred := base.PredictStats(MLBBattingAvg{NumGames: 1, PAs: 0.1, Hits: 0.2}).(MLBBattingAvg)
	if pred.PAs <= base.PAs {
		t.Fatalf("PredictStats should increase PAs with positive factor")
	}
}
