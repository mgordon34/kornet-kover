package main

import (
	"log"
	// "time"

	// "github.com/mgordon34/kornet-kover/internal/scraper"
	// "github.com/mgordon34/kornet-kover/internal/sportsbook"

	"github.com/mgordon34/kornet-kover/api/games"

	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    // storage.InitDB()
    // storage.InitTables()
    db := storage.GetDB()
    log.Println(db)
    storage.InitTables()

    game, err := games.GetLastGame()
    if err != nil {
        log.Fatal("Error getting game: ", err)
    }
    log.Println(game)

    // startDate, err := time.Parse("2006-01-02", "2024-05-01")
    // if err != nil {
    //     log.Fatal("Error parsing time: ", err)
    // }
    // endDate, err := time.Parse("2006-01-02", "2024-06-17")
    // if err != nil {
    //     log.Fatal("Error parsing time: ", err)
    // }
    // scraper.ScrapeGames(startDate, endDate)

    // game, err := games.GetLastGame()
    // if err != nil {
    //     log.Fatal("Error getting game: ", err)
    // }
    // log.Printf("game: %v", game)
    

    // sportsbook.GetGames(startDate, endDate)
}
