package analysis

import (
	"errors"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

func TestGetOrCreatePredictionBranches(t *testing.T) {
	endDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	controlMap := map[int]players.PlayerAvg{2026: players.NBAAvg{NumGames: 1, Minutes: 30, Points: 20}}
	want := players.NBAPIPPrediction{PlayerIndex: "p1", NumGames: 2, Minutes: 31, Points: 21}

	origGet := getPlayerPIPPredictionFn
	origCreate := createPIPPredictionFn
	t.Cleanup(func() {
		getPlayerPIPPredictionFn = origGet
		createPIPPredictionFn = origCreate
	})

	getPlayerPIPPredictionFn = func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
		return players.NBAPIPPrediction{}, errors.New("not found")
	}
	createPIPPredictionFn = func(playerIndex string, opponents []string, relationship players.Relationship, control map[int]players.PlayerAvg, startDate, endDate time.Time) players.NBAPIPPrediction {
		return want
	}

	got := GetOrCreatePrediction("p1", []string{"d1"}, players.Opponent, controlMap, endDate.AddDate(-1, 0, 0), endDate, false)
	if got.PlayerIndex != want.PlayerIndex || got.Points != want.Points {
		t.Fatalf("GetOrCreatePrediction() fallback create = %+v, want %+v", got, want)
	}

	getPlayerPIPPredictionFn = func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
		return players.NBAPIPPrediction{PlayerIndex: "p2", Points: 9}, nil
	}
	createPIPPredictionFn = func(playerIndex string, opponents []string, relationship players.Relationship, control map[int]players.PlayerAvg, startDate, endDate time.Time) players.NBAPIPPrediction {
		t.Fatalf("create should not be called when prediction exists")
		return players.NBAPIPPrediction{}
	}

	got2 := GetOrCreatePrediction("p2", []string{"d1"}, players.Opponent, controlMap, endDate.AddDate(-1, 0, 0), endDate, false)
	if got2.PlayerIndex != "p2" || got2.Points != 9 {
		t.Fatalf("GetOrCreatePrediction() existing = %+v", got2)
	}

	createPIPPredictionFn = func(playerIndex string, opponents []string, relationship players.Relationship, control map[int]players.PlayerAvg, startDate, endDate time.Time) players.NBAPIPPrediction {
		return players.NBAPIPPrediction{PlayerIndex: "forced", Points: 7}
	}
	getPlayerPIPPredictionFn = func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
		t.Fatalf("get should not be called during forceUpdate")
		return players.NBAPIPPrediction{}, nil
	}
	got3 := GetOrCreatePrediction("p3", []string{"d1"}, players.Opponent, controlMap, endDate.AddDate(-1, 0, 0), endDate, true)
	if got3.PlayerIndex != "forced" {
		t.Fatalf("force update result = %+v", got3)
	}
}

func TestCreateAndStorePIPPrediction_UsesInjectedStore(t *testing.T) {
	origStore := addPIPPredictionFn
	t.Cleanup(func() { addPIPPredictionFn = origStore })

	called := false
	addPIPPredictionFn = func(preds []players.NBAPIPPrediction) {
		called = true
		if len(preds) != 1 {
			t.Fatalf("stored preds len = %d, want 1", len(preds))
		}
		if preds[0].PlayerIndex != "p1" || preds[0].Points != 22 {
			t.Fatalf("unexpected stored prediction: %+v", preds[0])
		}
	}

	CreateAndStorePIPPrediction([]Analysis{{
		PlayerIndex: "p1",
		Prediction:  players.NBAAvg{NumGames: 3, Minutes: 32, Points: 22, Rebounds: 8, Assists: 6, Threes: 2, Usg: 24, Ortg: 112, Drtg: 107},
	}}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	if !called {
		t.Fatalf("expected AddPIPPrediction seam to be called")
	}
}

func TestRunAnalysisOnGame_WithSeams(t *testing.T) {
	origPer := getPlayerPerByYearFn
	origGet := getOrCreatePredictionFn
	origStore := createAndStorePIPPredictionFn
	t.Cleanup(func() {
		getPlayerPerByYearFn = origPer
		getOrCreatePredictionFn = origGet
		createAndStorePIPPredictionFn = origStore
	})

	storeCalled := false
	createAndStorePIPPredictionFn = func(analyses []Analysis, date time.Time) { storeCalled = true }
	getPlayerPerByYearFn = func(sport sports.Sport, player string, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{utils.DateToNBAYear(endDate): players.NBAAvg{NumGames: 2, Minutes: 30, Points: 20, Rebounds: 8, Assists: 6, Threes: 2, Usg: 22, Ortg: 110, Drtg: 107}}
	}
	getOrCreatePredictionFn = func(playerIndex string, opponents []string, relationship players.Relationship, controlMap map[int]players.PlayerAvg, startDate, endDate time.Time, forceUpdate bool) players.NBAPIPPrediction {
		return players.NBAPIPPrediction{PlayerIndex: playerIndex, NumGames: 3, Minutes: 31, Points: 22, Rebounds: 9, Assists: 7, Threes: 3, Usg: 23, Ortg: 111, Drtg: 106}
	}

	out := RunAnalysisOnGame(
		[]players.PlayerRoster{{PlayerIndex: "p1", Status: "Available", AvgMins: 30}, {PlayerIndex: "p2", Status: "Out", AvgMins: 30}},
		[]players.PlayerRoster{{PlayerIndex: "d1", Status: "Available", AvgMins: 25}},
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		false,
		true,
	)

	if len(out) != 1 || out[0].PlayerIndex != "p1" {
		t.Fatalf("RunAnalysisOnGame output = %+v", out)
	}
	if !storeCalled {
		t.Fatalf("expected store hook to be called")
	}
}

func TestRunMLBAnalysisOnGame_WithSeams(t *testing.T) {
	origPer := getPlayerPerByYearFn
	origCreateMLB := createMLBPredictionFn
	origStore := createAndStorePIPPredictionFn
	t.Cleanup(func() {
		getPlayerPerByYearFn = origPer
		createMLBPredictionFn = origCreateMLB
		createAndStorePIPPredictionFn = origStore
	})

	stored := false
	createAndStorePIPPredictionFn = func(analyses []Analysis, date time.Time) { stored = true }
	getPlayerPerByYearFn = func(sport sports.Sport, player string, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{endDate.Year(): players.MLBBattingAvg{NumGames: 2, PAs: 5, Hits: 2, HomeRuns: 1, Strikeouts: 1, Runs: 1, AtBats: 4, RBIs: 2, Walks: 1, Pitches: 20, Strikes: 12, BA: 0.3, OBP: 0.4, SLG: 0.5, OPS: 0.9, WPA: 0.1}}
	}
	createMLBPredictionFn = func(playerIndex string, opponents []string, relationship players.Relationship, controlMap map[int]players.PlayerAvg, startDate, endDate time.Time) players.MLBBattingAvg {
		return players.MLBBattingAvg{NumGames: 1, PAs: 6, Hits: 3}
	}

	out := RunMLBAnalysisOnGame(
		[]players.PlayerRoster{{PlayerIndex: "b1", Status: "Available", AvgMins: 20}},
		[]players.PlayerRoster{{PlayerIndex: "p1", Status: "Available", AvgMins: 20}},
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		false,
		true,
	)

	if len(out) != 0 {
		t.Fatalf("RunMLBAnalysisOnGame currently expected empty predictions, got %+v", out)
	}
	if !stored {
		t.Fatalf("expected store hook to be called for MLB path")
	}
}

func TestCreatePIPPredictionAndCreateMLBPrediction_WithInjectedSeams(t *testing.T) {
	origPerWith := getPlayerPerWithPlayerByYearFn
	origMLBPerWith := getMLBPlayerPerWithPlayerByYearFn
	origCalc := calculatePIPFactorPredFn
	t.Cleanup(func() {
		getPlayerPerWithPlayerByYearFn = origPerWith
		getMLBPlayerPerWithPlayerByYearFn = origMLBPerWith
		calculatePIPFactorPredFn = origCalc
	})

	getPlayerPerWithPlayerByYearFn = func(player, defender string, relationship players.Relationship, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{utils.DateToNBAYear(endDate): players.NBAAvg{NumGames: 1, Minutes: 1, Points: 1}}
	}
	getMLBPlayerPerWithPlayerByYearFn = func(player, defender string, startDate, endDate time.Time) map[int]players.PlayerAvg {
		return map[int]players.PlayerAvg{endDate.Year(): players.MLBBattingAvg{NumGames: 1, PAs: 1, Hits: 1}}
	}
	calculatePIPFactorPredFn = func(controlMap, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
		for _, v := range controlMap {
			switch v.(type) {
			case players.NBAAvg:
				return players.NBAAvg{NumGames: 1, Minutes: 0.1, Points: 0.2, Rebounds: 0.1, Assists: 0.1, Threes: 0.1}
			case players.MLBBattingAvg:
				return players.MLBBattingAvg{NumGames: 1, PAs: 0.1, Hits: 0.1}
			}
		}
		return players.NBAAvg{}
	}

	nbaDate := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	nbaControl := map[int]players.PlayerAvg{utils.DateToNBAYear(nbaDate): players.NBAAvg{NumGames: 10, Minutes: 30, Points: 20, Rebounds: 8, Assists: 6, Threes: 2, Usg: 20, Ortg: 110, Drtg: 107}}
	nbaPred := CreatePIPPrediction("p1", []string{"d1", "d2"}, players.Opponent, nbaControl, nbaDate.AddDate(-1, 0, 0), nbaDate)
	if nbaPred.PlayerIndex != "p1" || nbaPred.Points <= 20 {
		t.Fatalf("CreatePIPPrediction() = %+v", nbaPred)
	}

	mlbDate := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	mlbControl := map[int]players.PlayerAvg{mlbDate.Year(): players.MLBBattingAvg{NumGames: 10, PAs: 5, Hits: 2, Runs: 1, AtBats: 4, RBIs: 1, HomeRuns: 1, Walks: 1, Strikeouts: 1, Pitches: 20, Strikes: 12, BA: 0.3, OBP: 0.4, SLG: 0.5, OPS: 0.9, WPA: 0.1}}
	mlbPred := CreateMLBPrediction("b1", []string{"p1"}, players.Opponent, mlbControl, mlbDate.AddDate(-1, 0, 0), mlbDate)
	if mlbPred.Hits <= 2 {
		t.Fatalf("CreateMLBPrediction() = %+v", mlbPred)
	}
}
