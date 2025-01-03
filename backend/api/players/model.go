package players

import "time"

type Player struct {
    Index       string    `json:"index"`
    Sport       string    `json:"sport"`
    Name        string    `json:"name"`
}

type Roster struct {
    Starters        []string
    Bench           []string
    Out             []string
}

type PlayerGame struct {
    PlayerIndex     string      `json:"player_index"`
    Game            int         `json:"game"`
    TeamIndex       string      `json:"team_index"`
    Minutes         float32     `json:"minutes"`
    Points          int         `json:"points"`
    Rebounds        int         `json:"rebounds"`
    Assists         int         `json:"assists"`
    Usg             float32     `json:"usg"`
    Ortg            int         `json:"drtg"`
    Drtg            int         `json:"ortg"`
} 

type PIPFactor struct {
    PlayerIndex     string      `json:"player_index"`
    OtherIndex      string      `json:"other_index"`
    Relationship    string      `json:"relationship"`
    Averages        *NBAAvg
}

type NBAPIPPrediction struct {
    PlayerIndex     string      `json:"player_index"`
    Date            time.Time   `json:"date"`
    Version         int         `json:"version"`
    NumGames        int         `json:"num_games"`
    Minutes         float32     `json:"minutes"`
    Points          float32     `json:"points"`
    Rebounds        float32     `json:"rebounds"`
    Assists         float32     `json:"assists"`
    Usg             float32     `json:"usg"`
    Ortg            float32     `json:"drtg"`
    Drtg            float32     `json:"ortg"`
}

type PlayerRoster struct {
    Id              int         `json:"id"`
    Sport           string      `json:"sport"`
    PlayerIndex     string      `json:"player_index"`
    TeamIndex       string      `json:"team_index"`
    Status          string      `json:"status"`
    AvgMins         float32     `json:"avg_minutes" db:"avg_minutes"`
}

func CurrNBAPIPPredVersion() int {
    return 1
}
