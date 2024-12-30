package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/picks"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/strategies"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/backtesting"
	"github.com/mgordon34/kornet-kover/internal/scraper"
	"github.com/mgordon34/kornet-kover/internal/sportsbook"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func main() {
    fmt.Println("Starting server")
    storage.InitTables()
    log.Println("Initialized DB")

    // runUpdateGames()
    // runUpdateLines()
    // runPickProps()

    // runBacktest()

    r := gin.Default()

    r.GET("/update-games", scraper.GetUpdateGames)
    r.GET("/update-lines", sportsbook.GetUpdateLines)
    r.GET("/pick-props", analysis.GetPickProps)

    r.GET("/strategies", strategies.GetStrategies)
    r.GET("/prop-picks", picks.GetPropPicks)

    r.Run(":8080")
}

func runUpdateGames() {
    log.Println("Updating games...")
    scraper.UpdateGames()
}

func runUpdateLines() {
    log.Println("Updating lines...")
    sportsbook.UpdateLines()
}

func runGetPIPPredictions() {
    log.Println("Updating PIPPredictions...")
    date, _ := time.Parse("2006-01-02", "2023-10-30")
    preds, _ := players.GetPIPPredictionsForDate(date)
    for _, pred := range preds {
        log.Println(pred)
    }
    log.Println(len(preds))
}

func runSportsbookGetGames() {
    loc, _ := time.LoadLocation("America/New_York")
    startDate, _ := time.ParseInLocation("2006-01-02", "2024-01-25", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2024-10-24", loc)
    log.Printf("Finding games from %v to %v", startDate, endDate)

    sportsbook.GetOdds(startDate, endDate)
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
        results := analysis.RunAnalysisOnGame(game[0], game[1], today, true, true)
        results = append(results, analysis.RunAnalysisOnGame(game[1], game[0], today, true, true)...)

        for _, outcome := range results {
            log.Printf("[%v]: Base Stats: %v", outcome.PlayerIndex, outcome.BaseStats)
            log.Printf("[%v]: Predicted Stats: %v", outcome.PlayerIndex, outcome.Prediction)

            for stat, value := range outcome.Outliers {
                log.Printf("[%v]: Outlier %v: %v", outcome.PlayerIndex, stat, value)
            }
        }
    }

}

func runBacktest() {
    loc, _ := time.LoadLocation("America/New_York")
    // startDate, _ := time.ParseInLocation("2006-01-02", "2023-11-01", loc)
    startDate, _ := time.ParseInLocation("2006-01-02", "2024-11-01", loc)
    // endDate, _ := time.ParseInLocation("2006-01-02", "2023-10-31", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2024-11-30", loc)
    pPicker := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 0,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    rPicker := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 0,
            "assists": 100,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    aPicker := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": 0,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    pPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 0,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 0,
        MaxUnder: 1000,
        TotalMax: 2000,
    }
    rPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 0,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 0,
        MaxUnder: 1000,
        TotalMax: 1000,
    }
    aPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": 0,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 0,
        MaxUnder: 1000,
        TotalMax: 1000,
    }
    fpPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -125,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    frPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": .3,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -125,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    faPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": .3,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -125,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    fsPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": .3,
            "assists": .3,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: -125,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    fPickerP := analysis.PropSelector{
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": .3,
            "assists": .3,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: -125,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    b := backtesting.Backtester{
        StartDate: startDate,
        EndDate: endDate,
        Strategies: []backtesting.Strategy{
            {PropSelector: pPicker, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: rPicker, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: aPicker, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: pPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: rPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: aPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fpPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: frPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: faPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fsPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fPickerP, BacktestResult: &backtesting.BacktestResult{}},
        },
    }
    b.RunBacktest()
}
