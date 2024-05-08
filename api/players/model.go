package players

type Player struct {
  Index      string    `json:"index"`
  Sport      string    `json:"sport"`
  Name       string    `json:"name"`
}

type PlayerGame struct {
  PlayerIndex      string    `json:"player_index"`
  Game             int       `json:"game"`
  TeamIndex        string    `json:"team_index"`
  Minutes          float32   `json:"minutes"`
  Points           int       `json:"points"`
  Rebounds         int       `json:"rebounds"`
  Assists          int       `json:"assists"`
  Usg              float32   `json:"usg"`
  Ortg             int       `json:"drtg"`
  Drtg             int       `json:"ortg"`
}
