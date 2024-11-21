package main

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitTables()
    log.Println("Initialized DB")

    runGetPlayerPip()
}

func runUpdateGames() {
    scraper.UpdateGames()
}

func runSportsbookGetGames() {
    startDate, err := time.Parse("2006-01-02", "2024-10-22")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2024-11-17")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    log.Printf("Finding games from %v to %v", startDate, endDate)

    sportsbook.GetGames(startDate, endDate)
}

func runGetPlayerOdds() {
    startDate, err := time.Parse("2006-01-02", "2024-10-25")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }

    oddsMap, err := odds.GetPlayerOddsForDate(startDate)
    if err  != nil {
        log.Fatal("Error getting player odds", err)
    }
    for i, pOdds := range oddsMap {
        log.Printf("Player: %v, Odds: %v", i, pOdds)
    }
}

func runGetPlayerPip() {
    startDate, err := time.Parse("2006-01-02", "2018-10-13")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2024-11-16")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    index := "tatumja01"
    player, err := players.GetPlayer(index)
    pindex := "daniedy01"
    pplayer, err := players.GetPlayer(pindex)

    controlMap := players.GetPlayerPerByYear(player, startDate, endDate)
    affectedMap := players.GetPlayerPerWithPlayerByYear(player, pplayer, players.Opponent, startDate, endDate)
    pipFactor := players.CalculatePIPFactor(controlMap, affectedMap)
    prediction := controlMap[2024].PredictStats(pipFactor)
    log.Println(pipFactor)
    log.Println(controlMap[2024])
    log.Println(prediction)
}
