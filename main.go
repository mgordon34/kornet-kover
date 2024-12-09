package main

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/backtesting"
	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    storage.InitTables()
    log.Println("Initialized DB")

    // runUpdateGames()
    // runUpdateLines()
    // runPickProps()

    runBacktest()
}

func runUpdateGames() {
    log.Println("Updating games...")
    scraper.UpdateGames()
}

func runUpdateLines() {
    log.Println("Updating lines...")
    getter := sportsbook.OddsAPI{}
    getter.UpdateLines()
}

func runSportsbookGetGames() {
    getter := sportsbook.OddsAPI{}
    loc, _ := time.LoadLocation("America/New_York")
    startDate, _ := time.ParseInLocation("2006-01-02", "2024-01-25", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2024-10-24", loc)
    log.Printf("Finding games from %v to %v", startDate, endDate)

    getter.GetOdds(startDate, endDate)
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

    loc, _ := time.LoadLocation("America/New_York")
    t := time.Now()
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
    for _, game := range games {
        results := analysis.RunAnalysisOnGame(game[0], game[1], today)
        results = append(results, analysis.RunAnalysisOnGame(game[1], game[0], today)...)

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
    loc, _ := time.LoadLocation("America/New_York")
    t := time.Now()
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)

    // Gather player Odds map for upcoming games
    oddsMap, err := odds.GetPlayerOddsForDate(today, []string{"points, rebounds, assists"})
    if err  != nil {
        log.Fatal("Error getting player odds", err)
    }
    // Gather roster for today's games
    games := scraper.ScrapeTodaysGames()
    // games = games[:1]

    // Run analysis on each game
    var results []analysis.Analysis
    for _, game := range games {
        log.Printf("Running analysis on %v vs %v", game[0], game[1])
        results = append(results, analysis.RunAnalysisOnGame(game[0], game[1], today)...)
        results = append(results, analysis.RunAnalysisOnGame(game[1], game[0], today)...)
    }

    picker := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 2,
            "rebounds": .5,
            "assists": .5,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MaxOver: 10,
        MaxUnder: 10,
        TotalMax: 20,
    }
    picks, err := picker.PickProps(oddsMap, results)
    if err  != nil {
        log.Fatal("Error getting picking props", err)
    }
    for _, pick := range picks {
        log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff)
    }
}

func runBacktest() {
    loc, _ := time.LoadLocation("America/New_York")
    startDate, _ := time.ParseInLocation("2006-01-02", "2024-10-23", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2024-10-24", loc)
    picker := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 2,
            "rebounds": .5,
            "assists": .5,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        BetSize: 100,
        MaxOver: 10,
        MaxUnder: 10,
        TotalMax: 20,
    }
    b := backtesting.Backtester{
        StartDate: startDate,
        EndDate: endDate,
        Strategies: []backtesting.Strategy{
            {PropSelector: picker, BacktestResult: &backtesting.BacktestResult{}},
        },
    }
    b.RunBacktest()
}
