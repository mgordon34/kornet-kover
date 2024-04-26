package scraper

import (
	"fmt"
	"log"
    "strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)


func ScrapeGames(startDate time.Time, endDate time.Time) {
    baseUrl := "https://www.basketball-reference.com/boxscores/index.fcgi?month=%d&day=%d&year=%d"
    c := colly.NewCollector()
    c.OnHTML("td.gamelink", func(e *colly.HTMLElement) {
        games := e.ChildAttrs("a", "href")
        for _, gameString := range games {
            scrapeGame(gameString)
        }
    })

    for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Scraping games for date: %v", d)
        c.Visit(fmt.Sprintf(baseUrl, d.Month(), d.Day(), d.Year()))
    }
}

func scrapeGame(gameString string) {
    baseUrl := "https://www.basketball-reference.com"
    var teams [2]string
    var scores [2]int
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

        log.Printf("Away %s %d vs Home %s %d", teams[0], scores[0], teams[1], scores[1])
    })

    c.Visit(baseUrl + gameString)
}
