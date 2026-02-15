package analysis

import (
	"errors"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

type fakeAnalysisStore struct {
	getPlayerPIPPredictionFn       func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error)
	addPIPPredictionFn             func(predictions []players.NBAPIPPrediction)
	getPlayerPerByYearFn           func(sport sports.Sport, player string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	getPlayerPerWithPlayerByYearFn func(player string, defender string, relationship players.Relationship, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	getMLBPerWithPlayerByYearFn    func(player string, defender string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	calculatePIPFactorFn           func(controlMap map[int]players.PlayerAvg, relatedMap map[int]players.PlayerAvg) players.PlayerAvg
}

func (f fakeAnalysisStore) GetPlayerPIPPrediction(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
	if f.getPlayerPIPPredictionFn == nil {
		return players.NBAPIPPrediction{}, errors.New("not configured")
	}
	return f.getPlayerPIPPredictionFn(playerIndex, date)
}

func (f fakeAnalysisStore) AddPIPPrediction(predictions []players.NBAPIPPrediction) {
	if f.addPIPPredictionFn != nil {
		f.addPIPPredictionFn(predictions)
	}
}

func (f fakeAnalysisStore) GetPlayerPerByYear(sport sports.Sport, player string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	if f.getPlayerPerByYearFn == nil {
		return nil
	}
	return f.getPlayerPerByYearFn(sport, player, startDate, endDate)
}

func (f fakeAnalysisStore) GetPlayerPerWithPlayerByYear(player string, defender string, relationship players.Relationship, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	if f.getPlayerPerWithPlayerByYearFn == nil {
		return nil
	}
	return f.getPlayerPerWithPlayerByYearFn(player, defender, relationship, startDate, endDate)
}

func (f fakeAnalysisStore) GetMLBPlayerPerWithPlayerByYear(player string, defender string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	if f.getMLBPerWithPlayerByYearFn == nil {
		return nil
	}
	return f.getMLBPerWithPlayerByYearFn(player, defender, startDate, endDate)
}

func (f fakeAnalysisStore) CalculatePIPFactor(controlMap map[int]players.PlayerAvg, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
	if f.calculatePIPFactorFn == nil {
		return nil
	}
	return f.calculatePIPFactorFn(controlMap, relatedMap)
}

func TestGetOrCreatePredictionBranches(t *testing.T) {
	endDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	controlMap := map[int]players.PlayerAvg{2026: players.NBAAvg{NumGames: 1, Minutes: 30, Points: 20}}
	svc := NewAnalysisService(AnalysisServiceDeps{Store: fakeAnalysisStore{
		getPlayerPIPPredictionFn: func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
			return players.NBAPIPPrediction{}, errors.New("not found")
		},
		getPlayerPerWithPlayerByYearFn: func(player, defender string, relationship players.Relationship, startDate, endDate time.Time) map[int]players.PlayerAvg {
			return map[int]players.PlayerAvg{utils.DateToNBAYear(endDate): players.NBAAvg{NumGames: 1, Minutes: 1, Points: 1}}
		},
		calculatePIPFactorFn: func(controlMap, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
			return players.NBAAvg{NumGames: 1, Minutes: 0.1, Points: 0.2, Rebounds: 0.1, Assists: 0.1, Threes: 0.1}
		},
	}})

	got := svc.GetOrCreatePrediction("p1", []string{"d1"}, players.Opponent, controlMap, endDate.AddDate(-1, 0, 0), endDate, false)
	if got.PlayerIndex != "p1" || got.Points == 0 {
		t.Fatalf("GetOrCreatePrediction() fallback create = %+v", got)
	}

	svc2 := NewAnalysisService(AnalysisServiceDeps{Store: fakeAnalysisStore{
		getPlayerPIPPredictionFn: func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
			return players.NBAPIPPrediction{PlayerIndex: "p2", Points: 9}, nil
		},
	}})

	got2 := svc2.GetOrCreatePrediction("p2", []string{"d1"}, players.Opponent, controlMap, endDate.AddDate(-1, 0, 0), endDate, false)
	if got2.PlayerIndex != "p2" || got2.Points != 9 {
		t.Fatalf("GetOrCreatePrediction() existing = %+v", got2)
	}
}

func TestCreateAndStorePIPPrediction_UsesInjectedStore(t *testing.T) {
	called := false
	svc := NewAnalysisService(AnalysisServiceDeps{Store: fakeAnalysisStore{
		addPIPPredictionFn: func(preds []players.NBAPIPPrediction) {
			called = true
			if len(preds) != 1 || preds[0].PlayerIndex != "p1" {
				t.Fatalf("unexpected preds: %+v", preds)
			}
		},
	}})

	svc.CreateAndStorePIPPrediction([]Analysis{{
		PlayerIndex: "p1",
		Prediction:  players.NBAAvg{NumGames: 3, Minutes: 32, Points: 22, Rebounds: 8, Assists: 6, Threes: 2, Usg: 24, Ortg: 112, Drtg: 107},
	}}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	if !called {
		t.Fatalf("expected AddPIPPrediction to be called")
	}
}

func TestRunAnalysisOnGame_UsesServiceDeps(t *testing.T) {
	stored := false
	svc := NewAnalysisService(AnalysisServiceDeps{Store: fakeAnalysisStore{
		getPlayerPIPPredictionFn: func(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
			return players.NBAPIPPrediction{PlayerIndex: playerIndex, NumGames: 3, Minutes: 31, Points: 22, Rebounds: 9, Assists: 7, Threes: 3, Usg: 23, Ortg: 111, Drtg: 106}, nil
		},
		getPlayerPerByYearFn: func(sport sports.Sport, player string, startDate, endDate time.Time) map[int]players.PlayerAvg {
			return map[int]players.PlayerAvg{utils.DateToNBAYear(endDate): players.NBAAvg{NumGames: 2, Minutes: 30, Points: 20, Rebounds: 8, Assists: 6, Threes: 2, Usg: 22, Ortg: 110, Drtg: 107}}
		},
		addPIPPredictionFn: func(predictions []players.NBAPIPPrediction) {
			stored = true
		},
	}})

	out := svc.RunAnalysisOnGame(
		[]players.PlayerRoster{{PlayerIndex: "p1", Status: "Available", AvgMins: 30}, {PlayerIndex: "p2", Status: "Out", AvgMins: 30}},
		[]players.PlayerRoster{{PlayerIndex: "d1", Status: "Available", AvgMins: 25}},
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		false,
		true,
	)

	if len(out) != 1 || out[0].PlayerIndex != "p1" {
		t.Fatalf("RunAnalysisOnGame output = %+v", out)
	}
	if !stored {
		t.Fatalf("expected predictions to be stored")
	}
}

func TestCreatePredictions_WithInjectedStore(t *testing.T) {
	svc := NewAnalysisService(AnalysisServiceDeps{Store: fakeAnalysisStore{
		getPlayerPerWithPlayerByYearFn: func(player, defender string, relationship players.Relationship, startDate, endDate time.Time) map[int]players.PlayerAvg {
			return map[int]players.PlayerAvg{utils.DateToNBAYear(endDate): players.NBAAvg{NumGames: 1, Minutes: 1, Points: 1}}
		},
		getMLBPerWithPlayerByYearFn: func(player, defender string, startDate, endDate time.Time) map[int]players.PlayerAvg {
			return map[int]players.PlayerAvg{endDate.Year(): players.MLBBattingAvg{NumGames: 1, PAs: 1, Hits: 1}}
		},
		calculatePIPFactorFn: func(controlMap, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
			for _, v := range controlMap {
				switch v.(type) {
				case players.NBAAvg:
					return players.NBAAvg{NumGames: 1, Minutes: 0.1, Points: 0.2, Rebounds: 0.1, Assists: 0.1, Threes: 0.1}
				case players.MLBBattingAvg:
					return players.MLBBattingAvg{NumGames: 1, PAs: 0.1, Hits: 0.1}
				}
			}
			return nil
		},
	}})

	nbaDate := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	nbaControl := map[int]players.PlayerAvg{utils.DateToNBAYear(nbaDate): players.NBAAvg{NumGames: 10, Minutes: 30, Points: 20, Rebounds: 8, Assists: 6, Threes: 2, Usg: 20, Ortg: 110, Drtg: 107}}
	nbaPred := svc.CreatePIPPrediction("p1", []string{"d1", "d2"}, players.Opponent, nbaControl, nbaDate.AddDate(-1, 0, 0), nbaDate)
	if nbaPred.PlayerIndex != "p1" || nbaPred.Points <= 20 {
		t.Fatalf("CreatePIPPrediction() = %+v", nbaPred)
	}

	mlbDate := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	mlbControl := map[int]players.PlayerAvg{mlbDate.Year(): players.MLBBattingAvg{NumGames: 10, PAs: 5, Hits: 2, Runs: 1, AtBats: 4, RBIs: 1, HomeRuns: 1, Walks: 1, Strikeouts: 1, Pitches: 20, Strikes: 12, BA: 0.3, OBP: 0.4, SLG: 0.5, OPS: 0.9, WPA: 0.1}}
	mlbPred := svc.CreateMLBPrediction("b1", []string{"p1"}, players.Opponent, mlbControl, mlbDate.AddDate(-1, 0, 0), mlbDate)
	if mlbPred.Hits <= 2 {
		t.Fatalf("CreateMLBPrediction() = %+v", mlbPred)
	}
}
