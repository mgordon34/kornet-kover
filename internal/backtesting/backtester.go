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
        log.Printf("Skipping result for %s, no stats found", pick.PlayerIndex)
        return
    }
    b.Bets = append(b.Bets, &pick)
    actualValue := result.GetStats()[pick.Stat]

    if pick.Side == "Over" && actualValue > pick.Over.Line || pick.Side == "Under" && actualValue < pick.Under.Line {
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

func (b BacktestResult) printResults() {
    log.Println("------------------------------------------")
    log.Printf("%v Bets with %.2f%% winrate. Profits: $%.2f", b.Losses + b.Wins, (float32(b.Wins)/float32(b.Losses + b.Wins))*100, b.Profit)
    log.Println("------------------------------------------")
}

func (b BacktestResult) resultBreakdown() {
    log.Println("------------------------------------------")
    brackets := map[float32][]analysis.PropPick{
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
    keys := []float32{0,.1,.2,.3,.4,.5,.6,.7,.8,.9,1}
    for _, pick := range b.Bets {
        for key := range brackets {
            if pick.PDiff > key {
                brackets[key] = append(brackets[key], *pick)
            }
        }
    }

    for _, key := range keys {
        var wins, profit float32
        for _, bet := range brackets[key] {
            if bet.Result == "Win" {
                wins++
                profit += calculateProfit(bet.BetSize, bet.GetLine().Odds)
            } else {
                profit -= bet.BetSize
            }
        }
        log.Printf("%v: %v winrate and $%.2f profit", key, wins/float32(len(brackets[key])), profit)
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
        strategy.printResults()
        strategy.resultBreakdown()
    }
}

func (b Backtester) backtestDate(date time.Time) {
    date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
    log.Printf("Running for date %v", date)

    todayGames, err := games.GetGamesForDate(date)
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

    todaysOdds, err := odds.GetPlayerOddsForDate(date, []string{"points", "rebounds", "assists"})
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
        homeRoster := players.Roster{
            Starters: convertPlayerstoIndex(playerMap["home"][:5]),
        }
        awayRoster := players.Roster{
            Starters: convertPlayerstoIndex(playerMap["away"][:5]),
        }
        results = append(results, analysis.RunAnalysisOnGame(homeRoster, awayRoster, date)...)
        results = append(results, analysis.RunAnalysisOnGame(awayRoster, homeRoster, date)...)
    }

    var picks []analysis.PropPick
    for _, strategy := range b.Strategies {
        picks, _ = strategy.PickProps(todaysOdds, results)

        for _, pick := range picks {
            log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f Odds: %v/%v", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff, pick.Over.Odds, pick.Under.Odds)
            strategy.addResult(pick, statMap[pick.PlayerIndex])
        }
    }
}

func convertPlayerstoIndex(players []players.Player) []string {
    var indexes []string
    for _, player := range players {
        indexes = append(indexes, player.Index)
    }

    return indexes
}
