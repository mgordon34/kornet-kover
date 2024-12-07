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
    BacktestResult
}

type BacktestResult struct {
    Strategy            *analysis.PropSelector
    StartDate           time.Time
    EndDate             time.Time
    NumBets             int
    Wins                int
    Losses              int
    Profit              float32
}

func (b BacktestResult) addResult(pick analysis.PropPick, result players.PlayerAvg) {
    if result == nil {
        log.Printf("Skipping result for %s, no stats found", pick.PlayerIndex)
    }
    actualValue := result.GetStats()[pick.Stat]
    var odds int
    var line float32
    if pick.Side == "Over" {
        odds = pick.Over.Odds
        line = pick.Over.Line
    } else {
        odds = pick.Under.Odds
        line = pick.Under.Line
    }

    if pick.Side == "Over" && actualValue > pick.Over.Line || pick.Side == "Under" && actualValue < pick.Under.Line {
        b.Wins++
        b.Profit += calculateProfit(pick.BetSize, odds)
        log.Printf("Bet is win. line %v vs actual %v. Profits $%.2f", line, actualValue, calculateProfit(pick.BetSize, odds))
    } else {
        b.Losses++
        b.Profit -= pick.BetSize
        log.Printf("Bet is loss. line %v vs actual %v", line, actualValue)
    }
    b.NumBets++
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
}

func (b Backtester) backtestDate(date time.Time) {
    date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
    log.Printf("Running for date %v", date)

    todayGames, err := games.GetGamesForDate(date)
    if err != nil {
        log.Fatal("Error getting games for date: ", err)
    }

    var strs []string
    for _, game := range todayGames {
        strs = append(strs, strconv.FormatInt(int64(game.Id), 10))
    }
    statMap, err := players.GetPlayerStatsForGames(strs)
    if err != nil {
        log.Fatal("Error getting historical stats: ", err)
    }
    for player, performance := range statMap {
        log.Printf("%v: %v", player, performance)
    }

    todaysOdds, err := odds.GetPlayerOddsForDate(date, []string{"points", "rebounds", "assists"})
    if err != nil {
        log.Fatal("Error getting historical odds: ", err)
    }

    var results []analysis.Analysis
    todayGames = []games.Game{todayGames[0]}
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
        results = append(results, analysis.RunAnalysisOnGame(homeRoster, awayRoster)...)
        results = append(results, analysis.RunAnalysisOnGame(awayRoster, homeRoster)...)
    }

    var picks []analysis.PropPick
    for _, strategy := range b.Strategies {
        log.Println("Running results against strategy...")
        log.Printf("BacktestResult games: %v", strategy.NumBets)
        picks, _ = strategy.PickProps(todaysOdds, results)

        for _, pick := range picks {
            log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff)
            strategy.BacktestResult.addResult(pick, statMap[pick.PlayerIndex])
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
