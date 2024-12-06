package backtesting

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/analysis"
)

type Backtester struct {
    StartDate           time.Time
    EndDate             time.Time
    Strategies          []analysis.PropSelector
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

    for index, odds := range todaysOdds["points"] {
        log.Printf("%v odds: %v", index, odds)
    }

    var picks []analysis.PropPick
    for _, strategy := range b.Strategies {
        log.Println("Running results against strategy...")
        picks, _ = strategy.PickProps(todaysOdds, results)
    }

    for _, pick := range picks {
        log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff)
    }
}

func convertPlayerstoIndex(players []players.Player) []string {
    var indexes []string
    for _, player := range players {
        indexes = append(indexes, player.Index)
    }

    return indexes
}
