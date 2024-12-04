package backtesting

import (
	"log"
	"time"

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
    log.Printf("Running for date %v", date)
}
