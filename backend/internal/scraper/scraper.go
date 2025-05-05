package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

func ScrapeNbaTeams() {
    c := colly.NewCollector()
    var nbaTeams []teams.Team

    c.OnHTML("table#confs_standings_E > tbody", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            index := strings.Split(tr.ChildAttr("a", "href"), "/")[2]
            name := tr.ChildText("a")
            nbaTeams = append(nbaTeams, teams.Team{Index: index, Name: name})
        })
    })
    c.OnHTML("table#confs_standings_W > tbody", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            index := strings.Split(tr.ChildAttr("a", "href"), "/")[2]
            name := tr.ChildText("a")
            nbaTeams = append(nbaTeams, teams.Team{Index: index, Name: name})
        })
    })

    c.Visit(utils.SportConfigs[utils.NBA].Domain + "/leagues/NBA_2024_standings.html")

    teams.AddTeams(nbaTeams)
}

func ScrapeMLBTeams() {
    c := colly.NewCollector()
    var mlbTeams []teams.Team

    c.OnHTML("table > tbody", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            index := "MLB_" + strings.Split(tr.ChildAttr("a", "href"), "/")[2]
            name := tr.ChildText("a")
            mlbTeams = append(mlbTeams, teams.Team{Index: index, Name: name})
        })
    })

    c.Visit(utils.SportConfigs[utils.MLB].Domain + "/leagues/majors/2024-standings.shtml")

    teams.AddTeams(mlbTeams)
}

func ScrapeGames(sport utils.Sport, startDate time.Time, endDate time.Time) error {
    config, ok := utils.SportConfigs[sport]
    if !ok {
        return fmt.Errorf("unsupported sport: %s", sport)
    }

    c := colly.NewCollector()
    c.OnHTML("td.gamelink", func(e *colly.HTMLElement) {
        games := e.ChildAttrs("a", "href")
        for _, gameString := range games {
            time.Sleep(4 * time.Second)
            scrapeGame(sport, gameString)
        }
    })

    for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
        log.Printf("Scraping %s games for date: %v", sport, d)
        time.Sleep(4 * time.Second)

        url := fmt.Sprintf("%s%s/index.fcgi?month=%d&day=%d&year=%d",
            config.Domain,
            config.BoxScoreURL,
            d.Month(),
            d.Day(),
            d.Year(),
        )
        c.Visit(url)
    }

    return nil
}

func getDate(gameString string, sport utils.Sport) (time.Time, error) {
    parts := strings.Split(gameString, "/")
    if len(parts) < 3 {
        return time.Time{}, fmt.Errorf("invalid game string format: %s", gameString)
    }

    var dateStr string
    switch sport {
    case utils.NBA:
        // Format: /boxscores/202503010CHO.html
        dateStr = parts[2][:8]
    case utils.MLB:
        // Format: /boxes/PHI/PHI202310240.shtml
        dateStr = parts[3][3:11]
    default:
        return time.Time{}, fmt.Errorf("unsupported sport: %s", sport)
    }

    // Format as YYYY-MM-DD
    dateString := fmt.Sprintf("%s-%s-%s", dateStr[:4], dateStr[4:6], dateStr[6:8])

    date, err := time.Parse("2006-01-02", dateString)
    if err != nil {
        return time.Time{}, err
    }

    return date, nil
}

func scrapeGame(sport utils.Sport, gameString string) {
    log.Printf("Scraping %s game: %s", sport, gameString)
    config, ok := utils.SportConfigs[sport]
    if !ok {
        log.Printf("Unsupported sport: %s", sport)
        return
    }
    baseUrl := config.Domain
    
    c := colly.NewCollector()

    // Collect team and score information
    var teams [2]string
    var scores [2]int
    c.OnHTML("div.scorebox", func(div *colly.HTMLElement) {
        div.ForEach("strong", func(i int, e *colly.HTMLElement) {
            if i < 2 {
                team := e.ChildAttr("a", "href")
                teams[i] = strings.Split(team, "/")[2]
                if sport == utils.NBA {
                    teams[i] = strings.ReplaceAll(teams[i], "BKN", "BRK")
                } else if sport == utils.MLB {
                    teams[i] = "MLB_" + teams[i]
                }
            }
        })
        div.ForEach("div.score", func(i int, e *colly.HTMLElement) {
            score, err := strconv.Atoi(e.Text)
            if err != nil {
                return
            }
            scores[i] = score
        })
    })

    // Extract comments from raw HTML
    var commentTables []*goquery.Document
    c.OnResponse(func(r *colly.Response) {
        commentTables = parseTablesFromComments(string(r.Body))
    })

    // Aggregate all player tables to get player stats
    var playerTables []*colly.HTMLElement
    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        playerTables = append(playerTables, t)
    })

    c.Visit(fmt.Sprintf("%s%s", baseUrl, gameString))

    date, err := getDate(gameString, sport)
    if err != nil {
        log.Printf("Error getting date string: %v", err)
        return
    }

    game := games.Game{
        Sport:     string(sport),
        HomeIndex: teams[1],
        AwayIndex: teams[0],
        HomeScore: scores[1],
        AwayScore: scores[0],
        Date:      date,
    }
    gameId, err := games.AddGame(game)
    if err != nil {
        log.Printf("Error adding game: %v", err)
    }

    switch sport {
    case utils.NBA:
        pSlice, pGames := scrapeNBAPlayerStats(playerTables, gameId)
        players.AddPlayers(pSlice)
        players.AddPlayerGames(pGames)
    case utils.MLB:
        pSlice, battingGames, pitchingGames, pbpSlice := scrapeMLBPlayerStats(commentTables, gameId, game)
        players.AddPlayers(pSlice)
        players.AddMLBPlayerGamesBatting(battingGames)
        players.AddMLBPlayerGamesPitching(pitchingGames)
        players.AddMLBPlayByPlays(pbpSlice)
    }

    // players.AddPlayers(pSlice)

    // players.AddPlayerGames(fixPlayerStats(gameId, playerGames))
}

func parseTablesFromComments(html string) []*goquery.Document {
    var commentTables []*goquery.Document

    commentStart := "<!--"
    commentEnd := "-->"
    for {
        start := strings.Index(html, commentStart)
        if start == -1 {
            break
        }
        end := strings.Index(html[start:], commentEnd)
        if end == -1 {
            break
        }
        end += start + len(commentEnd)
        comment := html[start:end]
        if strings.Contains(comment, "<table") {
            // Parse the table HTML from the comment
            tableHTML := comment[len(commentStart):len(comment)-len(commentEnd)]
            tableDoc, err := goquery.NewDocumentFromReader(strings.NewReader(tableHTML))
            if err != nil {
                log.Printf("Error parsing table HTML: %v", err)
                continue
            }
            commentTables = append(commentTables, tableDoc)

        }
        html = html[end:]
    }

    return commentTables
}

func scrapeNBAPlayerStats(playerTables []*colly.HTMLElement, gameId int) ([]players.Player, []players.PlayerGame) {
    var pSlice []players.Player
    playerGames := make(map[string]players.PlayerGame)

    for _, t := range playerTables {

        id := t.Attr("id")
        if strings.Contains(id, "game-basic") {
            pSlice = append(pSlice, getPlayers(t)...)

            teamIndex := strings.Split(id, "-")[1]
            collectStats(t, playerGames, teamIndex)
        }
        if strings.Contains(id, "game-advanced") {
            teamIndex := strings.Split(id, "-")[1]
            collectStats(t, playerGames, teamIndex)
        }
    }

    pGames := fixPlayerStats(gameId, playerGames)

    log.Printf("Players: %v", pSlice)
    log.Printf("Player games: %v", pGames)
    return pSlice, pGames
}

func scrapeMLBPlayerStats(commentTables []*goquery.Document, gameId int, game games.Game) ([]players.Player, []players.MLBPlayerGameBatting, []players.MLBPlayerGamePitching, []players.MLBPlayByPlay) {
    battingIndex := 0
    pitchingIndex := 0
    var battingGames []players.MLBPlayerGameBatting
    var pitchingGames []players.MLBPlayerGamePitching
    var pbpSlice []players.MLBPlayByPlay
    pMap := make(map[string]players.Player)
    for _, tableDoc := range commentTables {

        if strings.Contains(tableDoc.Find("table").AttrOr("class", ""), "stats_table") {
            tableDoc.Find("tr").Each(func(i int, s *goquery.Selection) {
                // log.Printf("Row %d: %s", i, s.Text())
            })
        }

        // Scrape batting stats
        if strings.Contains(tableDoc.Find("table").AttrOr("id", ""), "batting") {
            var teamIndex string
            if battingIndex == 0 {
                teamIndex = game.AwayIndex
            } else {
                teamIndex = game.HomeIndex
            }
            log.Printf("Team index: %s", teamIndex)

            tableDoc.Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
                if s.AttrOr("class", "") != "spacer" {
                    playerIndex := strings.Split(strings.Split(s.Find("th").Find("a").AttrOr("href", ""), "/")[3], ".")[0]
                    playerName, err := utils.NormalizeString(s.Find("th").Find("a").Text())
                    if err != nil {
                        log.Printf("Error normalizing player name: %v", err)
                    }
                    pMap[playerName] = players.Player{Index: playerIndex, Name: playerName, Sport: "mlb"}

                    pGame := players.MLBPlayerGameBatting {
                        PlayerIndex: playerIndex,
                        Game:        gameId,
                        TeamIndex:   teamIndex,
                    }
                    pGame = parseMLBPlayerGameBatting(pGame, s)
                    battingGames = append(battingGames, pGame)
                }
            })

            battingIndex++
        }

        // Scrape pitching stats
        if strings.Contains(tableDoc.Find("table").AttrOr("id", ""), "pitching") {
            var teamIndex string

            tableDoc.Find("tr").Each(func(i int, s *goquery.Selection) {
                if s.Find("th").Text() == "Team Totals" {
                    pitchingIndex++
                } else if s.Find("th").AttrOr("aria-label", "") != "Pitching" {

                    if pitchingIndex == 0 {
                        teamIndex = game.AwayIndex
                    } else {
                        teamIndex = game.HomeIndex
                    }

                    if s.AttrOr("class", "") != "spacer" {
                        playerIndex := strings.Split(strings.Split(s.Find("th").Find("a").AttrOr("href", ""), "/")[3], ".")[0]

                        pGame := players.MLBPlayerGamePitching {
                            PlayerIndex: playerIndex,
                            Game:        gameId,
                            TeamIndex:   teamIndex,
                        }
                        pGame = parseMLBPlayerGamePitching(pGame, s)
                        pitchingGames = append(pitchingGames, pGame)
                    }
                }
            })

            pitchingIndex++
        }

		// Scrape at bat stats
        if strings.Contains(tableDoc.Find("table").AttrOr("id", ""), "play_by_play") {
            var prevBatterName string
            var prevPBP players.MLBPlayByPlay
            batterAppearances := make(map[string]map[int]int) // Track appearances per batter per inning
            
            tableDoc.Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
                if strings.HasPrefix(s.AttrOr("id", ""), "event") {
                    batterName, _ := utils.NormalizeString(s.Find("td[data-stat='batter']").Text())
                    pitcherName, _ := utils.NormalizeString(s.Find("td[data-stat='pitcher']").Text())
                    inning, _ := strconv.Atoi(s.Find("th").Text()[1:])
                    
                    // Initialize batter's inning map if not exists
                    if _, exists := batterAppearances[batterName]; !exists {
                        batterAppearances[batterName] = make(map[int]int)
                    }
                    
                    // Increment appearance count for this batter in this inning
                    batterAppearances[batterName][inning]++
                    
                    if batterName == prevBatterName {
                        // Remove previous entry for this batter
                        pbpSlice = pbpSlice[:len(pbpSlice)-1]
                        batterAppearances[batterName][inning]--
                        
                        // Overwrite previous play by play for same batter
                        prevPBP = parseMLBPPlayByPlay(prevPBP, s)
                        prevPBP.Appearance = batterAppearances[batterName][inning]
                        
                        // Add updated entry
                        pbpSlice = append(pbpSlice, prevPBP)
                    } else {
                        // Create new play by play for new batter
                        pbp := players.MLBPlayByPlay{
                            Game: gameId,
                            BatterIndex: pMap[batterName].Index,
                            PitcherIndex: pMap[pitcherName].Index,
                            Inning: inning,
                            Appearance: batterAppearances[batterName][inning],
                        }
                        pbp = parseMLBPPlayByPlay(pbp, s)
                        prevPBP = pbp
                        prevBatterName = batterName
                        
                        // Add new entry
                        pbpSlice = append(pbpSlice, pbp)
                    }
                }
            })
        }

    }

    var pSlice []players.Player
    for _, p := range pMap {
        pSlice = append(pSlice, p)
    }
    log.Printf("Players: %v", pSlice)
    log.Printf("Batting games: %v", battingGames)
    log.Printf("Pitching games: %v", pitchingGames)
    log.Printf("PBP: %v", pbpSlice)
    return pSlice, battingGames, pitchingGames, pbpSlice
}

func parseMLBPPlayByPlay(pbp players.MLBPlayByPlay, row *goquery.Selection) players.MLBPlayByPlay {
    row.Find("td").Each(func(i int, td *goquery.Selection) {
        dataStat := td.AttrOr("data-stat", "")
        switch dataStat {
        case "outs":
            pbp.Outs, _ = strconv.Atoi(td.Text())
        case "pitches_pbp":
            pitch_text := td.Text()
            pbp.Pitches, _ = strconv.Atoi(strings.Split(pitch_text, ",")[0])
        case "play_desc":
            playDesc := strings.ToLower(td.Text())
            pbp.RawResult = playDesc
            if strings.HasPrefix(playDesc, "walk") || strings.HasPrefix(playDesc, "intentional walk") || strings.HasPrefix(playDesc, "hit by pitch") {
                pbp.Result = "Walk"
            } else if strings.HasPrefix(playDesc, "strikeout") {
                pbp.Result = "SO"
            } else if strings.HasPrefix(playDesc, "single") {
                pbp.Result = "1B"
            } else if strings.HasPrefix(playDesc, "double") {
                pbp.Result = "2B"
            } else if strings.HasPrefix(playDesc, "triple") {
                pbp.Result = "3B"
            } else if strings.HasPrefix(playDesc, "home run") {
                pbp.Result = "HR"
            } else if strings.Contains(playDesc, "out") || strings.Contains(playDesc, "flyball") || strings.Contains(playDesc, "popfly") || strings.Contains(playDesc, "double play") {
                pbp.Result = "Out"
            } else if strings.Contains(playDesc, "reached on") {
                pbp.Result = "Reached on Error"
            } else if strings.Contains(playDesc, "caught stealing") || strings.Contains(playDesc, "picked off") || strings.Contains(playDesc, "wild pitch") {
                pbp.Result = "Not Completed"
            } else {
                pbp.Result = "Parse Error"
                log.Printf("Parse Error: %s", playDesc)
            }
        }
    })

    return pbp
}

func parseMLBPlayerGameBatting(pGame players.MLBPlayerGameBatting, s *goquery.Selection) players.MLBPlayerGameBatting {
    s.Find("td").Each(func(i int, td *goquery.Selection) {
        dataStat := td.AttrOr("data-stat", "")
        switch dataStat {
        case "AB":
            pGame.AtBats, _ = strconv.Atoi(td.Text())
        case "R":
            pGame.Runs, _ = strconv.Atoi(td.Text())
        case "H":
            pGame.Hits, _ = strconv.Atoi(td.Text())
        case "RBI":
            pGame.RBIs, _ = strconv.Atoi(td.Text())
        case "BB":
            pGame.Walks, _ = strconv.Atoi(td.Text())
        case "SO":
            pGame.Strikeouts, _ = strconv.Atoi(td.Text())
        case "PA":
            pGame.PAs, _ = strconv.Atoi(td.Text())
        case "pitches":
            pGame.Pitches, _ = strconv.Atoi(td.Text())
        case "strikes_total":
            pGame.Strikes, _ = strconv.Atoi(td.Text())
        case "onbase_perc":
            obp, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.OBP = float32(obp)
        case "slugging_perc":
            slg, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.SLG = float32(slg)
        case "onbase_plus_slugging":
            ops, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.OPS = float32(ops)
        case "wpa_bat":
            wpa, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.WPA = float32(wpa)
        case "details":
            pGame.Details = td.Text()
            pGame.HomeRuns = parseHomeRuns(pGame.Details)
        }
    })

    return pGame
}

func parseMLBPlayerGamePitching(pGame players.MLBPlayerGamePitching, s *goquery.Selection) players.MLBPlayerGamePitching {
    s.Find("td").Each(func(i int, td *goquery.Selection) {
        dataStat := td.AttrOr("data-stat", "")
        switch dataStat {
        case "IP":
            ip, err := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            if err != nil {
                log.Printf("Error parsing innings: %v", err)
            }
            pGame.Innings = float32(ip)
        case "H":
            pGame.Hits, _ = strconv.Atoi(td.Text())
        case "R":
            pGame.Runs, _ = strconv.Atoi(td.Text())
        case "ER":
            pGame.EarnedRuns, _ = strconv.Atoi(td.Text())
        case "BB":
            pGame.Walks, _ = strconv.Atoi(td.Text())
        case "SO":
            pGame.Strikeouts, _ = strconv.Atoi(td.Text())
        case "HR":
            pGame.HomeRuns, _ = strconv.Atoi(td.Text())
        case "earned_run_avg":
            era, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.ERA = float32(era)
        case "batters_faced":
            pGame.BattersFaced, _ = strconv.Atoi(td.Text())
        case "wpa_def":
            wpa, _ := strconv.ParseFloat(strings.TrimSpace(td.Text()), 32)
            pGame.WPA = float32(wpa)
        }
    })

    return pGame
}

func parseHomeRuns(details string) int {
    if details == "" {
        return 0
    }

    stats := strings.Split(details, ",")
    for _, stat := range stats {
        stat = strings.TrimSpace(stat)
        if strings.HasSuffix(stat, "HR") {
            if strings.Contains(stat, "·") {
                num := strings.Split(stat, "·")[0]
                hrs, _ := strconv.Atoi(num)
                return hrs
            }
            return 1
        }
    }
    return 0
}
func collectStats(t *colly.HTMLElement, playerGames map[string]players.PlayerGame, tIndex string) {
    t.ForEach("tbody > tr", func(i int, tr *colly.HTMLElement) {
        if i == 5 {
            return
        }
        index := strings.Split(tr.ChildAttr("a", "href"), "/")[3]
        index = strings.TrimSuffix(index, ".html")

        _, exists := playerGames[index]
        if !exists {
            playerGames[index] = players.PlayerGame{PlayerIndex: index, TeamIndex: tIndex}
        }
        tr.ForEach("td", func(i int, td *colly.HTMLElement) {
            stat := td.Attr("data-stat")
            value := td.Text
            playerGames[index] = addPlayerStat(stat, value, playerGames[index])
        })
    })
}

func getPlayers(t *colly.HTMLElement) []players.Player {
    var pSlice []players.Player

    t.ForEach("tbody > tr", func(i int, tr *colly.HTMLElement) {
        if i == 5 {
            return
        }

        index := strings.Split(tr.ChildAttr("a", "href"), "/")[3]
        index = strings.TrimSuffix(index, ".html")
        name := tr.ChildText("a")

        pSlice = append(pSlice, players.Player{Index: index, Sport: "nba", Name: name})
    })

    return pSlice
}

func addPlayerStat(stat string, value string, playerGame players.PlayerGame) players.PlayerGame {
    switch stat {
    case "mp":
        s := strings.Split(value, ":")
        minStr, secStr := s[0], s[1]
        minutes, _ := strconv.Atoi(minStr)
        seconds, _ := strconv.Atoi(secStr)
        minsPlayed := float32(minutes) + float32(seconds)/60
        playerGame.Minutes = minsPlayed
    case "pts":
        points, _ := strconv.Atoi(value)
        playerGame.Points = points
    case "trb":
        reb, _ := strconv.Atoi(value)
        playerGame.Rebounds = reb
    case "ast":
        ast, _ := strconv.Atoi(value)
        playerGame.Assists = ast
    case "fg3":
        fg3, _ := strconv.Atoi(value)
        playerGame.Threes = fg3
    case "usg_pct":
        u, _ := strconv.ParseFloat(value, 32)
        usg := float32(u)
        playerGame.Usg = usg
    case "off_rtg":
        ortg, _ := strconv.Atoi(value)
        playerGame.Ortg = ortg
    case "def_rtg":
        drtg, _ := strconv.Atoi(value)
        playerGame.Drtg = drtg
    }
    return playerGame
}

func fixPlayerStats(gameId int, pMap map[string]players.PlayerGame) []players.PlayerGame {
    var pSlice []players.PlayerGame

    for _, v := range pMap {
        if v.Minutes == 0 {
            continue
        }
        v.Game = gameId
        pSlice = append(pSlice, v)
    }

    return pSlice
}

func GetUpdateGames(c *gin.Context) {
    err := UpdateGames(utils.NBA)
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, "Done")
}

func GetUpdateActiveRosters(c *gin.Context) {
    err := UpdateActiveRosters()
    if err != nil {
        msg := fmt.Sprint("Error updating active rosters: ", err)
        log.Println(msg)
        c.JSON(http.StatusInternalServerError, msg)
    }
    c.JSON(http.StatusOK, "Done")
}

// UpdateGames will add any new game and corresponding stats to the database
// This is done by utilizing GetLastGame to determine the date window to perform game scraping
// Returns the number of new games added or error
// TODO: Optimizations for offseason could be made here
func UpdateGames(sport utils.Sport) error {
    lastGame, err := games.GetLastGame()
    if err != nil {
        return err
    }

    lastGameDate := lastGame.Date
    startDate := lastGameDate.AddDate(0, 0, 1)
    endDate := time.Now()
    return ScrapeGames(sport, startDate, endDate)
}

func UpdateActiveRosters() error {
    var activeRoster []players.PlayerRoster
    injuredPlayers := GetInjuredPlayers()
    tList, err := teams.GetTeams()
    if err != nil {
        return err
    }

    for _, team := range tList {
        activeRoster = append(activeRoster, scrapePlayersForTeam(team.Index, injuredPlayers)...)
    }

    activeRoster = pruneActiveRoster(activeRoster)

    for _, player := range activeRoster {
        players.UpdatePlayerTables(player.PlayerIndex)
    }

    err = players.UpdateRosters(activeRoster)
    if err != nil {
        return err
    }

    return nil
}

func scrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
    var roster []players.PlayerRoster

    url := fmt.Sprintf("%s/teams/%v/2025.html", utils.SportConfigs[utils.NBA].Domain, teamIndex)
    c := colly.NewCollector()
    log.Println("Visiting team page for ", teamIndex)
    time.Sleep(4 * time.Second)

    var rosterPlayers []string
    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        id := t.Attr("id")
        if id == "roster" {
            rosterPlayers = getPlayersOnRoster(t)
        } else if id == "per_game_stats" {
            roster = getPlayersByTime(teamIndex, rosterPlayers, injuredPlayers, t)
        }

    })
    c.Visit(url)

    return roster
}

func getPlayersOnRoster(t *colly.HTMLElement) []string {
    var rosterPlayers []string

    t.ForEach("tbody", func(i int, tb *colly.HTMLElement) {
        tb.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            tr.ForEach("td", func(i int, td *colly.HTMLElement) {
                dataStat := td.Attr("data-stat")

                if dataStat == "player" {
                    firstSplit := strings.Split(td.ChildAttr("a", "href"), "/")[3]
                    playerIndex := strings.Split(firstSplit, ".")[0]
                    rosterPlayers = append(rosterPlayers, playerIndex)
                }
            })

        })
    })

    return rosterPlayers
}

func getPlayersByTime(teamIndex string, rosterPlayers []string, injuredPlayers map[string]string, t *colly.HTMLElement) []players.PlayerRoster {
    var roster []players.PlayerRoster

    t.ForEach("tbody", func(i int, tb *colly.HTMLElement) {
        tb.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            var playerIndex string
            var avgMins float32

            tr.ForEach("td", func(i int, td *colly.HTMLElement) {
                dataStat := td.Attr("data-stat")

                if dataStat == "name_display" && td.Attr("data-append-csv") != "" {
                    playerIndex = td.Attr("data-append-csv")
                } else if dataStat == "mp_per_g" {
                    mins, _ := strconv.ParseFloat(td.Text, 64)
                    avgMins = float32(mins)
                }

            })

            // Remove players that are no longer listed on active roster
            if !slices.Contains(rosterPlayers, playerIndex) {
                log.Printf("%v is no longer on the roster", playerIndex)
                return
            }

            var status string
            if _, ok := injuredPlayers[playerIndex]; ok {
                log.Printf("%v is out for today", playerIndex)
                status = "Out"
            } else {
                status = "Available"
            }

            roster = append(roster, players.PlayerRoster{
                Sport:       "nba",
                PlayerIndex: playerIndex,
                TeamIndex:   teamIndex,
                Status:      status,
                AvgMins:     avgMins,
            })
        })
    })

    return roster
}

func pruneActiveRoster(activeRoster []players.PlayerRoster) []players.PlayerRoster {
    var prunedRoster []players.PlayerRoster
    var foundPlayers []string

    for _, player := range activeRoster {
        if slices.Contains(foundPlayers, player.PlayerIndex) {
            log.Printf("Found duplicate for %s, skipping...", player.PlayerIndex)
        } else {
            foundPlayers = append(foundPlayers, player.PlayerIndex)
            prunedRoster = append(prunedRoster, player)
        }
    }

    return prunedRoster
}

func ScrapeTodaysRosters() [][]players.Roster {
    baseUrl := "%s/leagues/NBA_2025_games-%v.html"
    c := colly.NewCollector()
    var games [][]players.Roster
    now := time.Now()
    month := strings.ToLower(now.Month().String())
    dateStr := now.Format("20060102")

    missingPlayers := GetInjuredPlayers()

    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            dataStat := tr.ChildAttr("th", "csk")
            if dataStat != "" && dataStat[:8] == dateStr {
                var homeRoster, awayRoster players.Roster
                tr.ForEach("td", func(i int, td *colly.HTMLElement) {
                    dataStat := td.Attr("data-stat")
                    if dataStat == "home_team_name" {
                        homeIndex := strings.Split(td.ChildAttr("a", "href"), "/")[2]
                        homeRoster = getRosterForTeam(homeIndex, missingPlayers)
                    } else if dataStat == "visitor_team_name" {
                        awayIndex := strings.Split(td.ChildAttr("a", "href"), "/")[2]
                        awayRoster = getRosterForTeam(awayIndex, missingPlayers)
                    }
                })
                games = append(games, []players.Roster{homeRoster, awayRoster})
            }
        })
    })

    str := fmt.Sprintf(baseUrl, utils.SportConfigs[utils.NBA].Domain, month)
    c.Visit(str)

    return games
}

func ScrapeTodaysGames() [][]string {
    baseUrl := "%s/leagues/NBA_2025_games-%v.html"
    c := colly.NewCollector()
    var games [][]string

    now := time.Now()
    month := strings.ToLower(now.Month().String())
    dateStr := now.Format("20060102")

    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            dataStat := tr.ChildAttr("th", "csk")
            if dataStat != "" && dataStat[:8] == dateStr {
                var matchup []string
                tr.ForEach("td", func(i int, td *colly.HTMLElement) {
                    dataStat := td.Attr("data-stat")
                    if dataStat == "home_team_name" {
                        matchup = append(matchup, strings.Split(td.ChildAttr("a", "href"), "/")[2])
                    } else if dataStat == "visitor_team_name" {
                        matchup = append(matchup, strings.Split(td.ChildAttr("a", "href"), "/")[2])
                    }
                })
                games = append(games, matchup)
            }
        })
    })

    str := fmt.Sprintf(baseUrl, utils.SportConfigs[utils.NBA].Domain, month)
    c.Visit(str)

    return games
}

func getRosterForTeam(teamIndex string, missingPlayers map[string]string) players.Roster {
    var roster = players.Roster{}
    url := fmt.Sprintf("%s/teams/%v/2025.html", utils.SportConfigs[utils.NBA].Domain, teamIndex)
    c := colly.NewCollector()
    log.Println("Visiting team page for ", teamIndex)
    time.Sleep(4 * time.Second)

    index := 0
    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        id := t.Attr("id")
        if id != "per_game_stats" {
            return
        }

        t.ForEach("td", func(i int, td *colly.HTMLElement) {
            dataStat := td.Attr("data-stat")
            if dataStat == "name_display" && td.Attr("data-append-csv") != "" {
                playerIndex := td.Attr("data-append-csv")

                if _, ok := missingPlayers[playerIndex]; ok {
                    log.Printf("%v is out for today", playerIndex)
                    roster.Out = append(roster.Out, playerIndex)
                } else if index < 8 {
                    roster.Starters = append(roster.Starters, playerIndex)
                    index++
                } else {
                    roster.Bench = append(roster.Bench, playerIndex)
                    index++
                }
            }
        })
    })
    c.Visit(url)

    return roster
}

func GetMissingPlayers() map[string]string {
    players := make(map[string]string)
    c := colly.NewCollector()

    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            var pIndex string
            dataStat := tr.ChildAttr("th", "data-stat")
            if dataStat == "player" {
                pIndex = tr.ChildAttr("th", "data-append-csv")
            }
            tr.ForEach("td", func(i int, td *colly.HTMLElement) {
                dataStat := td.Attr("data-stat")
                if dataStat == "note" {
                    reason := strings.ToLower(td.Text)
                    if strings.Contains(reason, "out") || strings.Contains(reason, "doubtful") || strings.Contains(reason, "questionable") {
                        players[pIndex] = reason
                    }
                }
            })
        })
    })

    c.Visit(utils.SportConfigs[utils.NBA].Domain + "/friv/injuries.fcgi")
    return players
}

type PlayerResp struct {
    player string
}

func GetInjuredPlayers() map[string]string {
    injuredPlayers := make(map[string]string)
    var jsonResp []map[string]string
    r, err := http.Get("https://www.rotowire.com/basketball/tables/injury-report.php?team=ALL&pos=ALL")
    if err != nil {
        return injuredPlayers
    }
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        log.Fatal(err)
    }
    json.Unmarshal(bodyBytes, &jsonResp)

    for _, player := range jsonResp {
        if strings.Split(player["status"], " ")[0] == "Out" {
            index, err := players.PlayerNameToIndex(make(map[string]string), player["player"])
            if err != nil {
                log.Printf("Error finding index for player %v", player["player"])
                continue
            }
            injuredPlayers[index] = player["status"]
        }
    }

    return injuredPlayers
}
