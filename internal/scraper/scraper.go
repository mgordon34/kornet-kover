package scraper

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/api/players"
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
