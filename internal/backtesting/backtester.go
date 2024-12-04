package backtesting

import (
    "log"
    "time"
)

type Backtester struct {
    StartDate           time.Time
    EndDate             time.Time
}

func (b Backtester) RunBacktest() {
    for d := b.StartDate; d.After(b.EndDate) == false; d = d.AddDate(0, 0, 1) {
        log.Printf("Running for date %v", d)
    }
}
