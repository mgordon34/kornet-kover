package odds

import "time"

type PlayerLine struct {
    Id              int       `json:"id"`
    Sport           string    `json:"sport"`
    PlayerIndex     string    `json:"player_index"`
    Timestamp       time.Time `json:"timestamp"`
    Type            string    `json:"type"`
    Stat            string    `json:"stat"`
    Side            string    `json:"side"`
    Line            float32   `json:"line"`
    Odds            int       `json:"odds"`
    Link            string    `json:"link"`
}
