package games

import "time"

type Game struct {
  Sport      string    `json:"sport"`
  HomeIndex  string    `json:"home_index"`
  AwayIndex  string    `json:"away_index"`
  HomeScore  string    `json:"home_score"`
  AwayScore  string    `json:"away_score"`
  Date       time.Time `json:"date"`
}
