package scraper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type fakeScraperSources struct {
	scrapeGamesFn          func(sport sports.Sport, startDate time.Time, endDate time.Time) error
	getInjuredPlayersFn    func() map[string]string
	scrapePlayersForTeamFn func(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster
}

func (f fakeScraperSources) ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error {
	if f.scrapeGamesFn == nil {
		return errors.New("ScrapeGames not configured")
	}
	return f.scrapeGamesFn(sport, startDate, endDate)
}

func (f fakeScraperSources) GetInjuredPlayers() map[string]string {
	if f.getInjuredPlayersFn == nil {
		return map[string]string{}
	}
	return f.getInjuredPlayersFn()
}

func (f fakeScraperSources) ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
	if f.scrapePlayersForTeamFn == nil {
		return nil
	}
	return f.scrapePlayersForTeamFn(teamIndex, injuredPlayers)
}

type fakeScraperStore struct {
	getLastGameFn        func() (games.Game, error)
	getTeamsFn           func() ([]teams.Team, error)
	updatePlayerTablesFn func(playerIndex string)
	updateRostersFn      func(rosterSlots []players.PlayerRoster) error
}

func (f fakeScraperStore) GetLastGame() (games.Game, error) {
	if f.getLastGameFn == nil {
		return games.Game{}, errors.New("GetLastGame not configured")
	}
	return f.getLastGameFn()
}

func (f fakeScraperStore) GetTeams() ([]teams.Team, error) {
	if f.getTeamsFn == nil {
		return nil, errors.New("GetTeams not configured")
	}
	return f.getTeamsFn()
}

func (f fakeScraperStore) UpdatePlayerTables(playerIndex string) {
	if f.updatePlayerTablesFn != nil {
		f.updatePlayerTablesFn(playerIndex)
	}
}

func (f fakeScraperStore) UpdateRosters(rosterSlots []players.PlayerRoster) error {
	if f.updateRostersFn == nil {
		return nil
	}
	return f.updateRostersFn(rosterSlots)
}

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

func TestCollyTableHelpers(t *testing.T) {
	html := `<html><body>
	<table class="stats_table" id="box-HOM-game-basic"><tbody>
	<tr><th><a href="/players/j/jamesle01.html">LeBron James</a></th>
	<td data-stat="mp">30:00</td><td data-stat="pts">25</td><td data-stat="trb">8</td><td data-stat="ast">6</td><td data-stat="fg3">2</td><td data-stat="usg_pct">24.5</td><td data-stat="off_rtg">112</td><td data-stat="def_rtg">108</td>
	</tr>
	</tbody></table>
	<table class="stats_table" id="box-HOM-game-advanced"><tbody>
	<tr><th><a href="/players/j/jamesle01.html">LeBron James</a></th>
	<td data-stat="usg_pct">24.5</td><td data-stat="off_rtg">112</td><td data-stat="def_rtg">108</td></tr>
	</tbody></table>
	<table class="stats_table" id="box-WNBA_HOM-game-basic"><tbody>
	<tr><th><a href="/wnba/players/j/jones01w.html">WNBA Player</a></th>
	<td data-stat="mp">28:30</td><td data-stat="pts">18</td><td data-stat="trb">7</td><td data-stat="ast">5</td><td data-stat="fg3">1</td><td data-stat="usg_pct">20.0</td><td data-stat="off_rtg">105</td><td data-stat="def_rtg">102</td>
	</tr>
	</tbody></table>
	<table class="stats_table" id="box-WNBA_HOM-game-advanced"><tbody>
	<tr><th><a href="/wnba/players/j/jones01w.html">WNBA Player</a></th>
	<td data-stat="usg_pct">20.0</td><td data-stat="off_rtg">105</td><td data-stat="def_rtg">102</td></tr>
	</tbody></table>
	<table id="roster" class="stats_table"><tbody>
	<tr><td data-stat="player"><a href="/players/j/jamesle01.html">LeBron James</a></td></tr>
	</tbody></table>
	<table id="per_game_stats" class="stats_table"><tbody>
	<tr><td data-stat="name_display" data-append-csv="jamesle01">LeBron James</td><td data-stat="mp_per_g">34.2</td></tr>
	</tbody></table>
	</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	c := colly.NewCollector()
	var tables []*colly.HTMLElement
	c.OnHTML("table.stats_table", func(e *colly.HTMLElement) {
		tables = append(tables, e)
	})
	if err := c.Visit(server.URL); err != nil {
		t.Fatalf("colly visit error: %v", err)
	}

	nbaPlayers := getPlayers(tables[0])
	if len(nbaPlayers) != 1 || nbaPlayers[0].Index != "jamesle01" {
		t.Fatalf("unexpected nba players: %+v", nbaPlayers)
	}

	wnbaPlayers := getWNBAPlayers(tables[2])
	if len(wnbaPlayers) != 1 || wnbaPlayers[0].Index != "jones01w" {
		t.Fatalf("unexpected wnba players: %+v", wnbaPlayers)
	}

	statsMap := map[string]players.PlayerGame{}
	collectStats(tables[0], statsMap, "HOM")
	collectWNBAStats(tables[2], statsMap, "WNBA_HOM")
	if statsMap["jamesle01"].Points != 25 || statsMap["jones01w"].Rebounds != 7 {
		t.Fatalf("unexpected collected stats: %+v", statsMap)
	}

	pSlice, pGames := scrapeNBAPlayerStats([]*colly.HTMLElement{tables[0], tables[1]}, 100)
	if len(pSlice) != 1 || len(pGames) != 1 || pGames[0].Game != 100 {
		t.Fatalf("unexpected scrapeNBAPlayerStats output: players=%+v games=%+v", pSlice, pGames)
	}

	wP, wGames := scrapeWNBAPlayerStats([]*colly.HTMLElement{tables[2], tables[3]}, 101)
	if len(wP) != 1 || len(wGames) != 1 || wGames[0].Game != 101 {
		t.Fatalf("unexpected scrapeWNBAPlayerStats output: players=%+v games=%+v", wP, wGames)
	}

	rosterPlayers := getPlayersOnRoster(tables[4])
	if len(rosterPlayers) != 1 || rosterPlayers[0] != "jamesle01" {
		t.Fatalf("unexpected roster players: %+v", rosterPlayers)
	}

	roster := getPlayersByTime("HOM", rosterPlayers, map[string]string{"jamesle01": "Out"}, tables[5])
	if len(roster) != 1 || roster[0].Status != "Out" {
		t.Fatalf("unexpected roster-by-time output: %+v", roster)
	}

	_, _, _, _ = scrapeMLBPlayerStats([]*goquery.Document{}, 1, games.Game{})
	_ = sports.NBA
}

func TestUpdateGamesAndHandlersUseService(t *testing.T) {
	svc := NewScraperService(ScraperServiceDeps{
		Store: fakeScraperStore{
			getLastGameFn: func() (games.Game, error) {
				return games.Game{Date: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
			},
			getTeamsFn:           func() ([]teams.Team, error) { return []teams.Team{}, nil },
			updatePlayerTablesFn: func(playerIndex string) {},
			updateRostersFn:      func(rosterSlots []players.PlayerRoster) error { return nil },
		},
		Sources: fakeScraperSources{
			scrapeGamesFn: func(sport sports.Sport, startDate time.Time, endDate time.Time) error {
				if !startDate.After(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)) {
					t.Fatalf("expected start date after last game date")
				}
				return nil
			},
			getInjuredPlayersFn: func() map[string]string { return map[string]string{} },
		},
		Now: func() time.Time { return time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC) },
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/update-games", UpdateGamesHandler(svc))
	r.GET("/update-players", UpdateActiveRostersHandler(svc))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/update-games", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("update-games status = %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/update-players", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("update-players status = %d", rec2.Code)
	}

	if err := svc.UpdateGames(sports.NBA); err != nil {
		t.Fatalf("UpdateGames() error = %v", err)
	}
}

func TestUpdateActiveRostersUsesService(t *testing.T) {
	updates := 0
	updatedRosters := 0

	svc := NewScraperService(ScraperServiceDeps{
		Store: fakeScraperStore{
			getLastGameFn:        func() (games.Game, error) { return games.Game{}, nil },
			getTeamsFn:           func() ([]teams.Team, error) { return []teams.Team{{Index: "A"}, {Index: "B"}}, nil },
			updatePlayerTablesFn: func(playerIndex string) { updates++ },
			updateRostersFn: func(rosterSlots []players.PlayerRoster) error {
				updatedRosters = len(rosterSlots)
				return nil
			},
		},
		Sources: fakeScraperSources{
			getInjuredPlayersFn: func() map[string]string { return map[string]string{"p2": "Out"} },
			scrapePlayersForTeamFn: func(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
				return []players.PlayerRoster{{Sport: "nba", PlayerIndex: "p1", TeamIndex: teamIndex, Status: "Available", AvgMins: 20}}
			},
		},
		Now: func() time.Time { return time.Now() },
	})

	if err := svc.UpdateActiveRosters(); err != nil {
		t.Fatalf("UpdateActiveRosters() error = %v", err)
	}
	if updates == 0 || updatedRosters == 0 {
		t.Fatalf("expected player table and roster updates to run")
	}
}

func TestGetPlayersByTime_AvailableAndRosterFilter(t *testing.T) {
	html := `<html><body><table id="per_game_stats" class="stats_table"><tbody>
	<tr><td data-stat="name_display" data-append-csv="keep01">Keep Player</td><td data-stat="mp_per_g">30.0</td></tr>
	<tr><td data-stat="name_display" data-append-csv="drop01">Drop Player</td><td data-stat="mp_per_g">10.0</td></tr>
	</tbody></table></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	c := colly.NewCollector()
	var table *colly.HTMLElement
	c.OnHTML("table.stats_table", func(e *colly.HTMLElement) { table = e })
	if err := c.Visit(server.URL); err != nil {
		t.Fatalf("visit err: %v", err)
	}

	roster := getPlayersByTime("TST", []string{"keep01"}, map[string]string{}, table)
	if len(roster) != 1 {
		t.Fatalf("expected one roster player after filter, got %d", len(roster))
	}
	if roster[0].PlayerIndex != "keep01" || roster[0].Status != "Available" {
		t.Fatalf("unexpected roster output: %+v", roster[0])
	}
}
