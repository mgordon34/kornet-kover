package analysis

import (
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
)

func TestPropPickGetLine(t *testing.T) {
	p := PropPick{PlayerLine: odds.PlayerLine{Line: 22.5, Side: "Over"}}
	if got := p.GetLine(); got.Line != 22.5 {
		t.Fatalf("GetLine() with PlayerLine = %+v", got)
	}

	p = PropPick{Side: "Over", PlayerOdds: odds.PlayerOdds{Over: odds.PlayerLine{Line: 10}}}
	if got := p.GetLine(); got.Line != 10 {
		t.Fatalf("GetLine() for over = %+v", got)
	}

	p = PropPick{Side: "Under", PlayerOdds: odds.PlayerOdds{Under: odds.PlayerLine{Line: 12}}}
	if got := p.GetLine(); got.Line != 12 {
		t.Fatalf("GetLine() for under = %+v", got)
	}
}

func TestConvertToPicksModel(t *testing.T) {
	now := time.Now()
	selector := PropSelector{StratId: 7}
	models := selector.convertToPicksModel([]PropPick{{LineId: 10}, {LineId: 11}}, now)
	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}
	if models[0].StratId != 7 || models[0].LineId != 10 || !models[0].Valid {
		t.Fatalf("unexpected first model: %+v", models[0])
	}
}

func TestSortPicksRanksByStatThenPDiff(t *testing.T) {
	selector := PropSelector{}
	items := []PropPick{
		{Stat: "assists", PDiff: 0.9},
		{Stat: "points", PDiff: 0.1},
		{Stat: "points", PDiff: 0.8},
		{Stat: "rebounds", PDiff: 0.7},
	}
	selector.sortPicks(items)

	if items[0].Stat != "points" || items[0].PDiff != 0.8 {
		t.Fatalf("unexpected first pick after sort: %+v", items[0])
	}
	if items[1].Stat != "points" {
		t.Fatalf("expected points picks first, got %+v", items)
	}
}

func TestPickPropsAndPickAlternateProps(t *testing.T) {
	analysis := Analysis{
		PlayerIndex: "p1",
		Prediction:  players.NBAAvg{NumGames: 20, Minutes: 35, Points: 26, Rebounds: 9, Assists: 6, Threes: 3},
		Outliers:    map[string]float32{"points": 0.2},
	}

	selector := PropSelector{
		Thresholds:   map[string]float32{"points": 0.1},
		TresholdType: Percent,
		MinOdds:      -200,
		MaxOdds:      200,
		MinGames:     1,
		MinMinutes:   1,
		MaxOver:      10,
		MaxUnder:     10,
		BetSize:      50,
	}

	props := map[string]map[string]odds.PlayerOdds{
		"p1": {
			"points": {
				Over:  odds.PlayerLine{Id: 1, Side: "Over", Line: 22.5, Odds: -110},
				Under: odds.PlayerLine{Id: 2, Side: "Under", Line: 22.5, Odds: -110},
			},
		},
	}

	picks, err := selector.PickProps(props, []Analysis{analysis}, time.Now(), false)
	if err != nil {
		t.Fatalf("PickProps() error = %v", err)
	}
	if len(picks) != 1 || picks[0].LineId != 1 || picks[0].Side != "Over" {
		t.Fatalf("unexpected PickProps output: %+v", picks)
	}

	altProps := map[string]map[string][]odds.PlayerLine{
		"p1": {
			"points": {
				{Id: 10, Side: "Over", Line: 23.5, Odds: 150},
				{Id: 11, Side: "Under", Line: 27.5, Odds: 180},
			},
		},
	}

	altPicks, err := selector.PickAlternateProps(altProps, []Analysis{analysis}, time.Now(), false)
	if err != nil {
		t.Fatalf("PickAlternateProps() error = %v", err)
	}
	if len(altPicks) == 0 {
		t.Fatalf("expected alternate picks")
	}
}
