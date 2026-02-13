package scraper

import (
	"testing"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

func TestGetDateBySport(t *testing.T) {
	nbaDate, err := getDate("/boxscores/202603010CHO.html", sports.NBA)
	if err != nil || nbaDate.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("NBA getDate failed: %v %v", nbaDate, err)
	}

	wnbaDate, err := getDate("/wnba/boxscores/202603010CHO.html", sports.WNBA)
	if err != nil || wnbaDate.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("WNBA getDate failed: %v %v", wnbaDate, err)
	}

	mlbDate, err := getDate("/boxes/PHI/PHI202310240.shtml", sports.MLB)
	if err != nil || mlbDate.Format("2006-01-02") != "2023-10-24" {
		t.Fatalf("MLB getDate failed: %v %v", mlbDate, err)
	}

	if _, err := getDate("bad", sports.NBA); err == nil {
		t.Fatalf("expected error for invalid game string")
	}
}

func TestParseHomeRuns(t *testing.T) {
	if got := parseHomeRuns(""); got != 0 {
		t.Fatalf("empty details got %d, want 0", got)
	}
	if got := parseHomeRuns("2\u00b7HR, BB"); got != 2 {
		t.Fatalf("expected 2 HR, got %d", got)
	}
	if got := parseHomeRuns("HR, SO"); got != 1 {
		t.Fatalf("expected 1 HR, got %d", got)
	}
}

func TestAddPlayerStatAndFixPlayerStats(t *testing.T) {
	pg := players.PlayerGame{}
	pg = addPlayerStat("mp", "10:30", pg)
	pg = addPlayerStat("pts", "12", pg)
	pg = addPlayerStat("trb", "8", pg)
	pg = addPlayerStat("ast", "5", pg)
	pg = addPlayerStat("fg3", "3", pg)
	pg = addPlayerStat("usg_pct", "23.4", pg)
	pg = addPlayerStat("off_rtg", "111", pg)
	pg = addPlayerStat("def_rtg", "104", pg)

	if pg.Minutes <= 10 || pg.Points != 12 || pg.Rebounds != 8 || pg.Assists != 5 || pg.Threes != 3 || pg.Ortg != 111 || pg.Drtg != 104 {
		t.Fatalf("unexpected parsed player game: %+v", pg)
	}

	withMinutes := players.PlayerGame{PlayerIndex: "a", Minutes: 5}
	withoutMinutes := players.PlayerGame{PlayerIndex: "b", Minutes: 0}
	out := fixPlayerStats(123, map[string]players.PlayerGame{"a": withMinutes, "b": withoutMinutes})
	if len(out) != 1 || out[0].Game != 123 || out[0].PlayerIndex != "a" {
		t.Fatalf("fixPlayerStats unexpected output: %+v", out)
	}
}

func TestPruneActiveRoster(t *testing.T) {
	in := []players.PlayerRoster{
		{PlayerIndex: "a"},
		{PlayerIndex: "b"},
		{PlayerIndex: "a"},
	}
	out := pruneActiveRoster(in)
	if len(out) != 2 {
		t.Fatalf("pruneActiveRoster len = %d, want 2", len(out))
	}
}
