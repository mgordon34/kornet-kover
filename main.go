package main

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitTables()
    log.Println("Initialized DB")

    // runUpdateGames()
    // runUpdateLines()
    // runAnalysis()

    runPickProps()
}

func runUpdateGames() {
    log.Println("Updating games...")
    scraper.UpdateGames()
}

func runUpdateLines() {
    log.Println("Updating lines...")
    sportsbook.UpdateLines()
}

func runSportsbookGetGames() {
    startDate, err := time.Parse("2006-01-02", "2024-11-26")
    if err != nil {
        log.Fatal("Error parsing time: ", err)
    }
    endDate, err := time.Parse("2006-01-02", "2024-11-27")
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

    oddsMap, err := odds.GetPlayerOddsForDate(startDate, []string{"points", "rebounds", "assists"})
    if err  != nil {
        log.Fatal("Error getting player odds", err)
    }
    for i, pOdds := range oddsMap {
        log.Printf("Player: %v, Odds: %v", i, pOdds)
    }
}

func runGetPlayerOddsForToday() map[string]map[string]odds.PlayerOdds {
    t := time.Now()
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

    pOdds, err := odds.GetPlayerOddsForDate(today, []string{"points", "rebounds", "assists"})
    if err  != nil {
        log.Fatal("Error getting player odds", err)
    }

    return pOdds
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
    pindex := "daniedy01"

    controlMap := players.GetPlayerPerByYear(index, startDate, endDate)
    affectedMap := players.GetPlayerPerWithPlayerByYear(index, pindex, players.Opponent, startDate, endDate)
    pipFactor := players.CalculatePIPFactor(controlMap, affectedMap)
    prediction := controlMap[2024].PredictStats(pipFactor)
    log.Println(pipFactor)
    log.Println(controlMap[2024])
    log.Println(prediction)
}

func runAnalysis() {
    log.Println("Running analysis...")
    games := scraper.ScrapeTodaysGames()
    // var games [][]players.Roster
    // homeRoster := players.Roster{Starters: []string{"wiggian01", "greendr01", "podzibr01", "hieldbu01", "jackstr02"}}
    // awayRoster := players.Roster{Starters: []string{"gilgesh01", "willija06", "harteis01", "dortlu01", "wallaca01"}}
    // game := []players.Roster{homeRoster, awayRoster}
    // games = append(games, game)

    for _, game := range games {
        results := analysis.RunAnalysisOnGame(game[0], game[1])
        results = append(results, analysis.RunAnalysisOnGame(game[1], game[0])...)

        for _, outcome := range results {
            log.Printf("[%v]: Base Stats: %v", outcome.PlayerIndex, outcome.BaseStats)
            log.Printf("[%v]: Predicted Stats: %v", outcome.PlayerIndex, outcome.Prediction)

            for stat, value := range outcome.Outliers {
                log.Printf("[%v]: Outlier %v: %v", outcome.PlayerIndex, stat, value)
            }
        }
    }

}

func runPickProps() {
    t := time.Now()
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

    // Gather player Odds map for upcoming games
    oddsMap, err := odds.GetPlayerOddsForDate(today, []string{"points, rebounds, assists"})
    if err  != nil {
        log.Fatal("Error getting player odds", err)
    }
    // Gather roster for today's games
    games := scraper.ScrapeTodaysGames()
    games = games[:1]

    // Run analysis on each game
    var results []analysis.Analysis
    for _, game := range games {
        log.Printf("Running analysis on %v vs %v", game[0], game[1])
        results = append(results, analysis.RunAnalysisOnGame(game[0], game[1])...)
        results = append(results, analysis.RunAnalysisOnGame(game[1], game[0])...)
    }

    picker := analysis.PropSelector{}
    picks, err := picker.PickProps(oddsMap, results)
    if err  != nil {
        log.Fatal("Error getting picking props", err)
    }
    log.Println(picks)
}
