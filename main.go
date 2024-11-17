package main

import (
	"log"
	"time"


	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitTables()
    log.Println("Initialized DB")

    runSportsbookGetGames()
}

func runUpdateGames() {
    scraper.UpdateGames()
}

func runSportsbookGetGames() {
    startDate, err := time.Parse("2006-01-02", "2023-03-26")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2023-03-26")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    log.Printf("Finding games from %v to %v", startDate, endDate)

    sportsbook.GetGames(startDate, endDate)
}
