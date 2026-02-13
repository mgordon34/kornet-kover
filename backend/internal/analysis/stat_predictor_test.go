package analysis

import (
	"errors"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
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
