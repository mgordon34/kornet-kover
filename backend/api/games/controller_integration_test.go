//go:build integration
// +build integration

package games

import (
	"context"
	"testing"
	"time"

	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func ensureTeams(t *testing.T) {
	t.Helper()
	db := storage.GetDB()
	_, err := db.Exec(context.Background(), `INSERT INTO teams (index, name) VALUES ('TSTH', 'Test Home') ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatalf("failed inserting home team: %v", err)
	}
	_, err = db.Exec(context.Background(), `INSERT INTO teams (index, name) VALUES ('TSTA', 'Test Away') ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatalf("failed inserting away team: %v", err)
	}
}

func TestAddAndQueryGames(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	storage.InitTables()
	ensureTeams(t)

	date := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	game := Game{Sport: string(sports.NBA), HomeIndex: "TSTH", AwayIndex: "TSTA", HomeScore: 101, AwayScore: 99, Date: date}

	id, err := AddGame(game)
	if err != nil {
		t.Fatalf("AddGame() error = %v", err)
	}
	if id == 0 {
		t.Fatalf("AddGame() returned id=0")
	}

	last, err := GetLastGame()
	if err != nil {
		t.Fatalf("GetLastGame() error = %v", err)
	}
	if last.Id == 0 || last.Date.IsZero() {
		t.Fatalf("GetLastGame() returned invalid row: %+v", last)
	}

	games, err := GetGamesForDate(sports.NBA, date)
	if err != nil {
		t.Fatalf("GetGamesForDate() error = %v", err)
	}
	if len(games) == 0 {
		t.Fatalf("GetGamesForDate() expected at least one game")
	}
}

func TestGetGamesForDateNoResults(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	storage.InitTables()

	games, err := GetGamesForDate(sports.NBA, time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetGamesForDate() error = %v", err)
	}
	if len(games) != 0 {
		t.Fatalf("GetGamesForDate() len = %d, want 0", len(games))
	}
}
