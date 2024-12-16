package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
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

    c.Visit( "https://www.basketball-reference.com/leagues/NBA_2024_standings.html")

    teams.AddTeams(nbaTeams)
}

func ScrapeGames(startDate time.Time, endDate time.Time) {
    baseUrl := "https://www.basketball-reference.com/boxscores/index.fcgi?month=%d&day=%d&year=%d"
    c := colly.NewCollector()
    c.OnHTML("td.gamelink", func(e *colly.HTMLElement) {
        games := e.ChildAttrs("a", "href")
        for _, gameString := range games {
            time.Sleep(4 * time.Second)
            scrapeGame(gameString)
        }
    })

    for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Scraping games for date: %v", d)
        time.Sleep(4 * time.Second)

        c.Visit(fmt.Sprintf(baseUrl, d.Month(), d.Day(), d.Year()))
    }
}

func scrapeGame(gameString string) {
    baseUrl := "https://www.basketball-reference.com"
    var teams [2]string
    var scores [2]int
    var pSlice []players.Player
    playerGames := make(map[string]players.PlayerGame)
    c := colly.NewCollector()

    c.OnHTML("div.scorebox", func(div *colly.HTMLElement) {
        div.ForEach("strong", func(i int, e *colly.HTMLElement) {
            team := e.ChildAttr("a", "href")
            teams[i] = strings.Split(team, "/")[2]
        })
        div.ForEach("div.score", func(i int, e *colly.HTMLElement) {
            score, err := strconv.Atoi(e.Text)
            if err != nil {
                return
            }
            scores[i] = score
        })
    })

    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
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
    })

    c.Visit(baseUrl + gameString)

    dateString := strings.Split(gameString, "/")[2][:8]
    dateString = fmt.Sprintf("%s-%s-%s", dateString[:4], dateString[4:6], dateString[6:8])
    date, err := time.Parse("2006-01-02", dateString)
    if err != nil {
        return
    }

    game := games.Game {
        Sport: "nba",
        HomeIndex: teams[1],
        AwayIndex: teams[0],
        HomeScore: scores[1],
        AwayScore: scores[0],
        Date: date,
    }
    gameId, err := games.AddGame(game)
    if err != nil {
        return
    }

    players.AddPlayers(pSlice)

    var pgSlice []players.PlayerGame
    pgSlice = fixPlayerStats(gameId, playerGames)
    players.AddPlayerGames(pgSlice)
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

func getPlayers(t *colly.HTMLElement) []players.Player{
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
    err := UpdateGames()
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, "Done")
}

// UpdateGames will add any new game and corresponding stats to the database
// This is done by utilizing GetLastGame to determine the date window to perform game scraping
// Returns the number of new games added or error
// TODO: Optimizations for offseason could be made here
func UpdateGames() error{
    lastGame, err := games.GetLastGame()
    if err != nil {
        return err
    }

    startDate := lastGame.Date
    endDate := time.Now()
    ScrapeGames(startDate, endDate)

    return nil
}

func ScrapeTodaysGames() [][]players.Roster {
    baseUrl := "https://www.basketball-reference.com/leagues/NBA_2025_games-%v.html"
    c := colly.NewCollector()
    var games [][]players.Roster
    now := time.Now()
    month := strings.ToLower(now.Month().String())
    dateStr := now.Format("20060102")

    missingPlayers := GetInjuredPlayers()

    c.OnHTML("table.stats_table", func(t *colly.HTMLElement) {
        t.ForEach("tr", func(i int, tr *colly.HTMLElement) {
            dataStat := tr.ChildAttr("th", "csk")
            if dataStat != "" && dataStat[:8] == dateStr{
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

    str := fmt.Sprintf(baseUrl, month)
    c.Visit(str)

    return games
}

func getRosterForTeam(teamIndex string, missingPlayers map[string]string) players.Roster {
    var roster = players.Roster{}
    url := fmt.Sprintf("https://www.basketball-reference.com/teams/%v/2025.html", teamIndex)
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
            if dataStat == "name_display" && td.Attr("data-append-csv") != ""{
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
                    if strings.Contains(reason, "out") || strings.Contains(reason, "doubtful") || strings.Contains(reason, "questionable"){
                        players[pIndex] = reason
                    }
                }
            })
        })
    })

    c.Visit("https://www.basketball-reference.com/friv/injuries.fcgi")
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
