package backtesting

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
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
    date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.UTC().Location())
    log.Printf("Running for date %v", date)
    todayGames, err := games.GetGamesForDate(date)
    if err != nil {
        log.Fatal("Error getting games for date: ", err)
    }
    for _, game := range todayGames {
        log.Printf("Game %v: %v vs. %v", game.Id, game.HomeIndex, game.AwayIndex)
    }
}
