package backtesting

import (
	"log"
	"math"
	"strconv"
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type Strategy struct {
    analysis.PropSelector
    *BacktestResult
}

type BacktestResult struct {
    Bets                []*analysis.PropPick
    StartDate           time.Time
    EndDate             time.Time
    Wins                int
    Losses              int
    Profit              float32
}

func (b *BacktestResult) addResult(pick analysis.PropPick, result players.PlayerAvg) {
    if result == nil {
        log.Printf("Skipping result for %s, no stats found", pick.Analysis.PlayerIndex)
        return
    }
    b.Bets = append(b.Bets, &pick)
    actualValue := result.GetStats()[pick.Stat]

    if pick.Side == "Over" && actualValue > pick.GetLine().Line || pick.Side == "Under" && actualValue < pick.GetLine().Line {
        pick.Result = "Win"
        b.Wins++
        b.Profit += calculateProfit(pick.BetSize, pick.GetLine().Odds)
        log.Printf("Bet is win. line %v vs actual %v. Profits $%.2f", pick.GetLine().Line, actualValue, calculateProfit(pick.BetSize, pick.GetLine().Odds))
    } else {
        pick.Result = "Loss"
        b.Losses++
        b.Profit -= pick.BetSize
        log.Printf("Bet is loss. line %v vs actual %v", pick.GetLine().Line, actualValue)
    }
}

func (b BacktestResult) printResults(name string) {
    log.Println("------------------------------------------")
    log.Printf("Strategy: %s", name)
    log.Printf("%v Bets with %.2f%% winrate. Profits: $%.2f", b.Losses + b.Wins, (float32(b.Wins)/float32(b.Losses + b.Wins))*100, b.Profit)
    log.Println("------------------------------------------")
}

func (b BacktestResult) resultBreakdown() {
    log.Println("------------------------------------------")
    pBrackets := map[float64][]analysis.PropPick{
        -.5: []analysis.PropPick{},
        -.4: []analysis.PropPick{},
        -.3: []analysis.PropPick{},
        -.2: []analysis.PropPick{},
        -.1: []analysis.PropPick{},
        0: []analysis.PropPick{},
        .1: []analysis.PropPick{},
        .2: []analysis.PropPick{},
        .3: []analysis.PropPick{},
        .4: []analysis.PropPick{},
        .5: []analysis.PropPick{},
        .6: []analysis.PropPick{},
        .7: []analysis.PropPick{},
        .8: []analysis.PropPick{},
        .9: []analysis.PropPick{},
        1: []analysis.PropPick{},
    }
    pKeys := []float64{-.5,-.4,-.3,-.2,-.1,0,.1,.2,.3,.4,.5,.6,.7,.8,.9,1}
    rBrackets := map[float64][]analysis.PropPick{
        -5: []analysis.PropPick{},
        -4: []analysis.PropPick{},
        -3: []analysis.PropPick{},
        -2: []analysis.PropPick{},
        -1: []analysis.PropPick{},
        0: []analysis.PropPick{},
        .5: []analysis.PropPick{},
        1: []analysis.PropPick{},
        1.5: []analysis.PropPick{},
        2: []analysis.PropPick{},
        2.5: []analysis.PropPick{},
        3: []analysis.PropPick{},
        3.5: []analysis.PropPick{},
        4: []analysis.PropPick{},
        4.5: []analysis.PropPick{},
        5: []analysis.PropPick{},
        8: []analysis.PropPick{},
        10: []analysis.PropPick{},
        15: []analysis.PropPick{},
    }
    rKeys := []float64{-5,-4,-3,-2,-1,0,.5,1,1.5,2,2.5,3,3.5,4,4.5,5,8,10,15}
    oBrackets := map[int][]analysis.PropPick{
        0: []analysis.PropPick{},
        100: []analysis.PropPick{},
        200: []analysis.PropPick{},
        300: []analysis.PropPick{},
        400: []analysis.PropPick{},
        500: []analysis.PropPick{},
        600: []analysis.PropPick{},
        700: []analysis.PropPick{},
        800: []analysis.PropPick{},
        900: []analysis.PropPick{},
        1000: []analysis.PropPick{},
    }
    oKeys := []int{0,100,200,300,400,500,600,700,800,900,1000}
    lKeys := []float32{0,1,2,3,4,5,6,7,8,9,10,12,14,16,18,20,22,24,26,28,30}
    lBrackets := map[float32][]analysis.PropPick{
        0: []analysis.PropPick{},
        1: []analysis.PropPick{},
        2: []analysis.PropPick{},
        3: []analysis.PropPick{},
        4: []analysis.PropPick{},
        5: []analysis.PropPick{},
        6: []analysis.PropPick{},
        7: []analysis.PropPick{},
        8: []analysis.PropPick{},
        9: []analysis.PropPick{},
        10: []analysis.PropPick{},
        12: []analysis.PropPick{},
        14: []analysis.PropPick{},
        16: []analysis.PropPick{},
        18: []analysis.PropPick{},
        20: []analysis.PropPick{},
        22: []analysis.PropPick{},
        24: []analysis.PropPick{},
        26: []analysis.PropPick{},
        28: []analysis.PropPick{},
        30: []analysis.PropPick{},
    }
    for _, pick := range b.Bets {
        for key := range pBrackets {
            if float64(pick.PDiff) > key {
                pBrackets[key] = append(pBrackets[key], *pick)
            }
        }
        for key := range rBrackets {
            if float64(pick.Diff) > key {
                rBrackets[key] = append(rBrackets[key], *pick)
            }
        }
        for key := range oBrackets {
            if pick.GetLine().Odds >= key && pick.GetLine().Odds < (key + 100) {
                oBrackets[key] = append(oBrackets[key], *pick)
            }
			if key == 1000 && pick.GetLine().Odds >= 1100 {
                oBrackets[key] = append(oBrackets[key], *pick)
			}
			if key == 0 && pick.GetLine().Odds < 0 {
                oBrackets[key] = append(oBrackets[key], *pick)
			}
        }
        for key := range lBrackets {
            if pick.GetLine().Line < key {
                lBrackets[key] = append(lBrackets[key], *pick)
            }
        }
    }

    for _, key := range pKeys {
        var wins, profit float32
        for _, bet := range pBrackets[key] {
            if bet.Result == "Win" {
                wins++
                profit += calculateProfit(bet.BetSize, bet.GetLine().Odds)
            } else {
                profit -= bet.BetSize
            }
        }
        log.Printf("%v: %v winrate and $%.2f profit[%v]", key, wins/float32(len(pBrackets[key])), profit, len(pBrackets[key]))
    }
    log.Println("------------------------------------------")
    for _, key := range rKeys {
        var wins, profit float32
        for _, bet := range rBrackets[key] {
            if bet.Result == "Win" {
                wins++
                profit += calculateProfit(bet.BetSize, bet.GetLine().Odds)
            } else {
                profit -= bet.BetSize
            }
        }
        log.Printf("%v: %v winrate and $%.2f profit[%v]", key, wins/float32(len(rBrackets[key])), profit, len(rBrackets[key]))
    }
    log.Println("------------------------------------------")
    for _, key := range lKeys {
        var wins, profit float32
        for _, bet := range lBrackets[key] {
            if bet.Result == "Win" {
                wins++
                profit += calculateProfit(bet.BetSize, bet.GetLine().Odds)
            } else {
                profit -= bet.BetSize
            }
        }
        log.Printf("%v: %v winrate and $%.2f profit[%v]", key, wins/float32(len(lBrackets[key])), profit, len(lBrackets[key]))
    }
    log.Println("------------------------------------------")
    for _, key := range oKeys {
        var wins, profit float32
        for _, bet := range oBrackets[key] {
            if bet.Result == "Win" {
                wins++
                profit += calculateProfit(bet.BetSize, bet.GetLine().Odds)
            } else {
                profit -= bet.BetSize
            }
        }
        log.Printf("%v: %v winrate and $%.2f profit[%v]", key, wins/float32(len(oBrackets[key])), profit, len(oBrackets[key]))
    }
    log.Println("------------------------------------------")
}

func calculateProfit(betSize float32, odds int) float32 {
    if odds < 0 {
        return float32((100 / math.Abs(float64(odds))) * float64(betSize))
    } else {
        return (float32(odds) / 100) * betSize
    }
}

type Backtester struct {
    StartDate           time.Time
    EndDate             time.Time
    Strategies          []Strategy
}

func (b Backtester) RunBacktest() {
    for d := b.StartDate; d.After(b.EndDate) == false; d = d.AddDate(0, 0, 1) {
        b.backtestDate(d)
    }

    for _, strategy := range b.Strategies {
        strategy.printResults(strategy.StratName)
        strategy.resultBreakdown()
    }
}

func (b Backtester) backtestDate(date time.Time) {
    date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
    log.Printf("Running for date %v", date)

    todayGames, err := games.GetGamesForDate(sports.NBA, date)
    if err != nil {
        log.Fatal("Error getting games for date: ", err)
    }
    if len(todayGames) == 0 {
        log.Printf("No games for %v", date)
        return
    }

    var strs []string
    for _, game := range todayGames {
        strs = append(strs, strconv.FormatInt(int64(game.Id), 10))
    }
    statMap, err := players.GetPlayerStatsForGames(strs)
    if err != nil {
        log.Fatal("Error getting historical stats: ", err)
    }

    // todaysOdds, err := odds.GetPlayerOddsForDate(date, []string{"points", "rebounds", "assists", "threes"})
    todaysOdds, err := odds.GetAlternatePlayerOddsForDate(date, []string{"points", "rebounds", "assists", "threes"})
    if err != nil {
        log.Fatal("Error getting historical odds: ", err)
    }
    if len(todaysOdds) == 0 {
        log.Printf("No player odds for %v", date)
        return
    }

    var results []analysis.Analysis
    for _, game := range todayGames {
        log.Printf("Analyzing %v vs. %v", game.HomeIndex, game.AwayIndex)
        playerMap, err := players.GetPlayersForGame(game.Id, game.HomeIndex)
        if err != nil {
            log.Fatal("Error getting players for game: ", err)
        }
        // TODO: make this more intelligent by getting player's avg minutes for this point in the season
        homeRoster := convertPlayerMaptoPlayerRosters(playerMap["home"][:8])
        awayRoster := convertPlayerMaptoPlayerRosters(playerMap["away"][:8])

        results = append(results, analysis.RunAnalysisOnGame(homeRoster, awayRoster, date, false, true)...)
        results = append(results, analysis.RunAnalysisOnGame(awayRoster, homeRoster, date, false, true)...)
    }

    var picks []analysis.PropPick
    for _, strategy := range b.Strategies {
        picks, _ = strategy.PickAlternateProps(todaysOdds, results, date, false)

        for _, pick := range picks {
            log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f Odds: %v/%v", pick.Analysis.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.GetLine().Line, pick.Diff, pick.Over.Odds, pick.Under.Odds)
            strategy.addResult(pick, statMap[pick.Analysis.PlayerIndex])
        }
    }
}

func convertPlayerMaptoPlayerRosters(p []players.Player) []players.PlayerRoster {
    var playerRosters []players.PlayerRoster
    for _, player := range p {
        playerRosters = append(playerRosters, players.PlayerRoster{
            PlayerIndex: player.Index,
            Status: "Available",
            AvgMins: 21,
        })
    }

    return playerRosters
}

func convertPlayerstoIndex(players []players.Player) []string {
    var indexes []string
    for _, player := range players {
        indexes = append(indexes, player.Index)
    }

    return indexes
}
