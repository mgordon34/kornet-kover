package players

type Player struct {
    Index       string    `json:"index"`
    Sport       string    `json:"sport"`
    Name        string    `json:"name"`
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

type NBAAvg struct {
    NumGames     int         `json:"num_minutes"`
    Minutes      float32     `json:"avg_minutes"`
    Points       float32     `json:"avg_points"`
    Rebounds     float32     `json:"avg_rebounds"`
    Assists      float32     `json:"avg_assists"`
    Usg          float32     `json:"avg_usg"`
    Ortg         float32     `json:"avg_drtg"`
    Drtg         float32     `json:"avg_ortg"`
}

type PIPFactor struct {
    PlayerIndex     string      `json:"player_index"`
    OtherIndex      string      `json:"other_index"`
    Relationship    string      `json:"relationship"`
    Averages        *NBAAvg
}
