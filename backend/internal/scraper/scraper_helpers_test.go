package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

func TestParseTablesFromComments(t *testing.T) {
	html := `<html><body><!-- <table id="t1"><tr><td>a</td></tr></table> --><!-- no table --></body></html>`
	tables := parseTablesFromComments(html)
	if len(tables) != 1 {
		t.Fatalf("parseTablesFromComments len = %d, want 1", len(tables))
	}
	if tables[0].Find("table").AttrOr("id", "") != "t1" {
		t.Fatalf("expected parsed table id t1")
	}
}

func TestParseMLBPPlayByPlayResultClassification(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{name: "walk", desc: "Walk", want: "Walk"},
		{name: "strikeout", desc: "Strikeout swinging", want: "SO"},
		{name: "single", desc: "Single to left", want: "1B"},
		{name: "double", desc: "Ground-rule double", want: "2B"},
		{name: "triple", desc: "Triple to center", want: "3B"},
		{name: "home run", desc: "Inside-the-park home run", want: "HR"},
		{name: "out", desc: "Flyball to right", want: "Out"},
		{name: "error", desc: "Reached on throwing error", want: "Reached on Error"},
		{name: "not completed", desc: "Runner steals second", want: "Not Completed"},
		{name: "parse error", desc: "Unknown event text", want: "Parse Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<table><tbody><tr><td data-stat="outs">1</td><td data-stat="pitches_pbp">4,2</td><td data-stat="play_desc">` + tt.desc + `</td></tr></tbody></table>`))
			if err != nil {
				t.Fatalf("goquery parse err: %v", err)
			}
			row := doc.Find("tr").First()
			out := parseMLBPPlayByPlay(players.MLBPlayByPlay{}, row)
			if out.Result != tt.want {
				t.Fatalf("result = %q, want %q", out.Result, tt.want)
			}
			if out.Outs != 1 || out.Pitches != 4 {
				t.Fatalf("expected outs=1 pitches=4, got outs=%d pitches=%d", out.Outs, out.Pitches)
			}
		})
	}
}

func TestParseMLBPlayerGameRows(t *testing.T) {
	battingDoc, err := goquery.NewDocumentFromReader(strings.NewReader(`
	<table><tbody><tr>
		<td data-stat="AB">4</td><td data-stat="R">1</td><td data-stat="H">2</td><td data-stat="RBI">3</td>
		<td data-stat="BB">1</td><td data-stat="SO">0</td><td data-stat="PA">5</td>
		<td data-stat="pitches">21</td><td data-stat="strikes_total">14</td>
		<td data-stat="batting_avg">.300</td><td data-stat="onbase_perc">.360</td>
		<td data-stat="slugging_perc">.500</td><td data-stat="onbase_plus_slugging">.860</td>
		<td data-stat="wpa_bat">0.2</td><td data-stat="details">2·HR</td>
	</tr></tbody></table>`))
	if err != nil {
		t.Fatalf("goquery parse err: %v", err)
	}
	batting := parseMLBPlayerGameBatting(players.MLBPlayerGameBatting{}, battingDoc.Find("tr").First())
	if batting.AtBats != 4 || batting.Hits != 2 || batting.HomeRuns != 2 || batting.OPS < 0.8 {
		t.Fatalf("unexpected batting parse: %+v", batting)
	}

	pitchDoc, err := goquery.NewDocumentFromReader(strings.NewReader(`
	<table><tbody><tr>
		<td data-stat="IP">6.1</td><td data-stat="H">5</td><td data-stat="R">2</td><td data-stat="ER">2</td>
		<td data-stat="BB">1</td><td data-stat="SO">7</td><td data-stat="HR">1</td>
		<td data-stat="earned_run_avg">3.12</td><td data-stat="batters_faced">25</td><td data-stat="wpa_def">0.1</td>
	</tr></tbody></table>`))
	if err != nil {
		t.Fatalf("goquery parse err: %v", err)
	}
	pitching := parseMLBPlayerGamePitching(players.MLBPlayerGamePitching{}, pitchDoc.Find("tr").First())
	if pitching.Innings <= 6 || pitching.Strikeouts != 7 || pitching.BattersFaced != 25 {
		t.Fatalf("unexpected pitching parse: %+v", pitching)
	}
}
