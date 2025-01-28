package odds

import "time"

type PlayerLine struct {
    Id              int       `json:"id"`
    Sport           string    `json:"sport"`
    PlayerIndex     string    `json:"player_index"`
    Timestamp       time.Time `json:"timestamp"`
    Stat            string    `json:"stat"`
    Side            string    `json:"side"`
    Type            string    `json:"type"`
    Line            float32   `json:"line"`
    Odds            int       `json:"odds"`
    Link            string    `json:"link"`
}
