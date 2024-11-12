package games

import "time"

type Game struct {
    Id         int       `json:"id"`
    Sport      string    `json:"sport"`
    HomeIndex  string    `json:"home_index"`
    AwayIndex  string    `json:"away_index"`
    HomeScore  int       `json:"home_score"`
    AwayScore  int       `json:"away_score"`
    Date       time.Time `json:"date"`
}
