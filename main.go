package main

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitDB()
    storage.InitTables()

    startDate, err := time.Parse("2006-01-02", "2018-10-16")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2018-10-17")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    scraper.ScrapeGames(startDate, endDate)
}
