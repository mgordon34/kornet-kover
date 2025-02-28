package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
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
	"github.com/mgordon34/kornet-kover/internal/utils"
)

func main() {
    fmt.Println("Starting server")
    storage.InitTables()
    log.Println("Initialized DB")

    // runUpdateGames()
    // runUpdateLines()

    // runBacktest()

    startServer()
}

func startServer() {
    r := gin.Default()

    config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Replace with your frontend domain
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // If using cookies or credentials
		MaxAge:           12 * time.Hour, // Cache preflight response for 12 hours
	}

    r.Use(cors.New(config))

    r.GET("/update-games", scraper.GetUpdateGames)
    r.GET("/update-players", scraper.GetUpdateActiveRosters)
    r.GET("/update-lines", sportsbook.GetUpdateLines)
    r.GET("/pick-props", analysis.GetPickProps)

    r.GET("/strategies", strategies.GetStrategies)
    r.GET("/prop-picks", picks.GetPropPicks)

    r.Run(":8080")
}

func runUpdateGames() {
    log.Println("Updating games...")
    scraper.UpdateGames(utils.NBA)
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
    startDate, _ := time.ParseInLocation("2006-01-02", "2023-10-24", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2025-01-21", loc)
    log.Printf("Finding games from %v to %v", startDate, endDate)

    sportsbook.GetOdds(startDate, endDate, "mainline")
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

func runBacktest() {
    loc, _ := time.LoadLocation("America/New_York")
    startDate, _ := time.ParseInLocation("2006-01-02", "2023-12-01", loc)
    // startDate, _ := time.ParseInLocation("2006-01-02", "2024-12-01", loc)
    // endDate, _ := time.ParseInLocation("2006-01-02", "2025-01-20", loc)
    endDate, _ := time.ParseInLocation("2006-01-02", "2024-04-15", loc)
    pPicker := analysis.PropSelector{
        StratName: "Points Raw",
        Thresholds: map[string]float32{
            "points": -10,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: false,
        MinOdds: 300,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    rPicker := analysis.PropSelector{
        StratName: "Rebounds Raw",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": -10,
            "assists": 100,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: false,
        MinOdds: 300,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    aPicker := analysis.PropSelector{
        StratName: "Assists Raw",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": -10,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: false,
        MinOdds: 300,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    tPicker := analysis.PropSelector{
        StratName: "Threes Raw",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": 100,
            "threes": -10,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: false,
        MinOdds: 300,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 200,
    }
    pPickerP := analysis.PropSelector{
        StratName: "Points(outlier)",
        Thresholds: map[string]float32{
            "points": -.5,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: 500,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 2000,
    }
    rPickerP := analysis.PropSelector{
        StratName: "Rebounds(outlier)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": -.5,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: 500,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 1000,
    }
    aPickerP := analysis.PropSelector{
        StratName: "Assists(outlier)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": -.5,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: 500,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 1000,
    }
    tPickerP := analysis.PropSelector{
        StratName: "Threes(outlier)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": 100,
            "threes": -.5,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: 500,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 1000,
    }
    fpPickerP := analysis.PropSelector{
        StratName: "Points(weighted)",
        Thresholds: map[string]float32{
            "points": .5,
            "rebounds": 100,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: 100,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 100,
    }
    frPickerP := analysis.PropSelector{
        StratName: "Rebounds(weighted)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": .2,
            "assists": 100,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: 100,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 100,
    }
    faPickerP := analysis.PropSelector{
        StratName: "Assists(weighted)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": .2,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: 100,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 100,
    }
    ftPickerP := analysis.PropSelector{
        StratName: "Threes(weighted)",
        Thresholds: map[string]float32{
            "points": 100,
            "rebounds": 100,
            "assists": 100,
            "threes": .3,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: false,
        MinOdds: 100,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 100,
    }
    fsPickerP := analysis.PropSelector{
        StratName: "Final Points(Percent)",
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": 1000,
            "assists": 1000,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: -135,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 1000,
        MaxUnder: 0,
        TotalMax: 100,
    }
    foPickerP := analysis.PropSelector{
        StratName: "Final Points(Raw)",
        Thresholds: map[string]float32{
            "points": 2.5,
            "rebounds": 1000,
            "assists": 1000,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MinOdds: -135,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    fPickerP := analysis.PropSelector{
        StratName: "Final Rebounds",
        Thresholds: map[string]float32{
            "points": 1000,
            "rebounds": 1,
            "assists": 1000,
        },
        TresholdType: analysis.Raw,
        RequireOutlier: true,
        MinOdds: -135,
        MinGames: 10,
        MinMinutes: 0,
        BetSize: 100,
        MaxOver: 5,
        MaxUnder: 0,
        TotalMax: 100,
    }
    tfPicker := analysis.PropSelector{
        StratName: "Final Threes",
        Thresholds: map[string]float32{
            "points": 1000,
            "rebounds": 1000,
            "assists": 1000,
            "threes": .6,
        },
        TresholdType: analysis.Percent,
        RequireOutlier: true,
        MinOdds: -135,
        MinGames: 10,
        MinMinutes: 0,
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
            {PropSelector: tPicker, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: pPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: rPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: aPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: tPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fpPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: frPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: faPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: ftPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fsPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: foPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: fPickerP, BacktestResult: &backtesting.BacktestResult{}},
            {PropSelector: tfPicker, BacktestResult: &backtesting.BacktestResult{}},
        },
    }
    b.RunBacktest()
}
