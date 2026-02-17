//go:build integration
// +build integration

package odds

import (
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestOddsDatabaseFlows(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	storage.InitTables()

	players.AddPlayers([]players.Player{{Index: "oddsit01", Sport: "nba", Name: "Odds IT Player"}})

	date := time.Date(2099, 2, 2, 0, 0, 0, 0, time.UTC)
	ts1 := date.Add(2 * time.Hour)
	ts2 := date.Add(3 * time.Hour)

	AddPlayerLines([]PlayerLine{
		{Sport: "nba", PlayerIndex: "oddsit01", Timestamp: ts1, Stat: "points", Side: "Over", Type: "mainline", Line: 21.5, Odds: -120, Link: "a"},
		{Sport: "nba", PlayerIndex: "oddsit01", Timestamp: ts2, Stat: "points", Side: "Over", Type: "mainline", Line: 21.5, Odds: -110, Link: "b"},
		{Sport: "nba", PlayerIndex: "oddsit01", Timestamp: ts2, Stat: "points", Side: "Under", Type: "mainline", Line: 21.5, Odds: -105, Link: "c"},
		{Sport: "nba", PlayerIndex: "oddsit01", Timestamp: ts2, Stat: "points", Side: "Over", Type: "alternate", Line: 24.5, Odds: 180, Link: "d"},
	})

	last, err := GetLastLine("mainline")
	if err != nil {
		t.Fatalf("GetLastLine() error = %v", err)
	}
	if last.Type != "mainline" || last.Id == 0 {
		t.Fatalf("unexpected last line: %+v", last)
	}

	mainLines, err := GetPlayerLinesForDate(sports.NBA, date, "mainline")
	if err != nil {
		t.Fatalf("GetPlayerLinesForDate() error = %v", err)
	}
	if len(mainLines) < 2 {
		t.Fatalf("expected at least two main lines, got %d", len(mainLines))
	}

	oddsMap, err := GetPlayerOddsForDate(sports.NBA, date)
	if err != nil {
		t.Fatalf("GetPlayerOddsForDate() error = %v", err)
	}
	if oddsMap["oddsit01"]["points"].Over.Odds != -110 {
		t.Fatalf("expected closest over odds -110, got %d", oddsMap["oddsit01"]["points"].Over.Odds)
	}

	altMap, err := GetAlternatePlayerOddsForDate(sports.NBA, date)
	if err != nil {
		t.Fatalf("GetAlternatePlayerOddsForDate() error = %v", err)
	}
	if len(altMap["oddsit01"]["points"]) == 0 {
		t.Fatalf("expected alternate lines for oddsit01")
	}
}
