package main

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitDB()
    storage.InitTables()

    startDate, err := time.Parse("2006-01-02", "2023-10-24")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2023-10-24")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    startDate = endDate
    endDate = startDate
    // scraper.ScrapeGames(startDate, endDate)

    sportsbook.GetGames(startDate, endDate)
}
