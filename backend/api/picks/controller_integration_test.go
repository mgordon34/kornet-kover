//go:build integration
// +build integration

package picks

import (
	"context"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestPicksDatabaseFlows(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	storage.InitTables()
	db := storage.GetDB()

	teams.AddTeams([]teams.Team{{Index: "PKH", Name: "Picks Home"}})
	players.AddPlayers([]players.Player{{Index: "picksit01", Sport: "nba", Name: "Picks IT Player"}})

	var userID int
	err := db.QueryRow(context.Background(), `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`, "picks-user", "picks-user@example.com", "pw").Scan(&userID)
	if err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	var stratID int
	err = db.QueryRow(context.Background(), `INSERT INTO strategies (user_id, name) VALUES ($1, $2) RETURNING id`, userID, "Picks Strategy").Scan(&stratID)
	if err != nil {
		t.Fatalf("insert strategy error = %v", err)
	}

	date := time.Date(2099, 3, 3, 0, 0, 0, 0, time.UTC)
	odds.AddPlayerLines([]odds.PlayerLine{{Sport: "nba", PlayerIndex: "picksit01", Timestamp: date.Add(time.Hour), Stat: "points", Side: "Over", Type: "mainline", Line: 20.5, Odds: -110, Link: "x"}})

	lines, err := odds.GetPlayerLinesForDate(sports.NBA, date, "mainline")
	if err != nil || len(lines) == 0 {
		t.Fatalf("GetPlayerLinesForDate() = %d, err=%v", len(lines), err)
	}
	lineID := lines[0].Id

	players.AddPIPPrediction([]players.NBAPIPPrediction{{PlayerIndex: "picksit01", Date: date, Version: players.CurrNBAPIPPredVersion(), NumGames: 5, Minutes: 30, Points: 22, Rebounds: 7, Assists: 5, Threes: 2, Usg: 20, Ortg: 110, Drtg: 107}})

	err = players.UpdateRosters([]players.PlayerRoster{{Sport: "nba", PlayerIndex: "picksit01", TeamIndex: "PKH", Status: "Available", AvgMins: 30}})
	if err != nil {
		t.Fatalf("UpdateRosters() error = %v", err)
	}

	pickID, err := addPropPick(PropPick{StratId: stratID, LineId: lineID, Valid: true, Date: date})
	if err != nil || pickID == 0 {
		t.Fatalf("addPropPick() id=%d err=%v", pickID, err)
	}

	one, err := getPropPick(pickID)
	if err != nil || one.Id != pickID {
		t.Fatalf("getPropPick() got=%+v err=%v", one, err)
	}

	rows, err := getPropPicks(userID, date)
	if err != nil || len(rows) == 0 {
		t.Fatalf("getPropPicks() len=%d err=%v", len(rows), err)
	}

	bRows, err := getBettorPicks(userID, date)
	if err != nil || len(bRows) == 0 {
		t.Fatalf("getBettorPicks() len=%d err=%v", len(bRows), err)
	}

	err = AddPropPicks([]PropPick{{StratId: stratID, LineId: lineID, Valid: true, Date: date}})
	if err != nil {
		t.Fatalf("AddPropPicks() error = %v", err)
	}

	MarkOldPicksInvalid(stratID, date)
	rowsAfter, err := getPropPicks(userID, date)
	if err != nil {
		t.Fatalf("getPropPicks() after invalidate err = %v", err)
	}
	if len(rowsAfter) != 0 {
		t.Fatalf("expected no valid picks after invalidation, got %d", len(rowsAfter))
	}
}
