package odds

import "time"

type PlayerOdds struct {
    PlayerIndex     string    `json:"player_index"`
    Date            time.Time `json:"date"`
    Stat            string    `json:"stat"`
    Line            float32   `json:"line"`
    OverOdds        int       `json:"over_odds"`
    UnderOdds       int       `json:"under_odds"`
}
