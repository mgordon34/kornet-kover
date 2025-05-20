package players

import "time"

type Player struct {
    Index       string              `json:"index"`
    Sport       string              `json:"sport"` 
    Name        string              `json:"name"`
    Details     map[string]string   `json:"details,omitempty"`
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
    Threes          int         `json:"threes"`
    Usg             float32     `json:"usg"`
    Ortg            int         `json:"drtg"`
    Drtg            int         `json:"ortg"`
} 

type MLBPlayerGameBatting struct {
    PlayerIndex     string      `json:"player_index"`
    Game            int         `json:"game"`
    TeamIndex       string      `json:"team_index"`
    AtBats          int         `json:"at_bats"`
    Runs            int         `json:"runs"`
    Hits            int         `json:"hits"`
    RBIs            int         `json:"rbis"`
    HomeRuns        int         `json:"home_runs"`
    Walks           int         `json:"walks"`
    Strikeouts      int         `json:"strikeouts"`
    PAs             int         `json:"pas"`
    Pitches         int         `json:"pitches"`
    Strikes         int         `json:"strikes"`
    BA              float32     `json:"ba"`
    OBP             float32     `json:"obp"`
    SLG             float32     `json:"slg"`
    OPS             float32     `json:"ops"`
    WPA             float32     `json:"wpa"`
    Details         string      `json:"details"`
}

type MLBPlayerGamePitching struct {
    PlayerIndex     string      `json:"player_index"`
    Game            int         `json:"game"`
    TeamIndex       string      `json:"team_index"`
    Innings         float32     `json:"innings"`
    Hits            int         `json:"hits"`
    Runs            int         `json:"runs"`
    EarnedRuns      int         `json:"earned_runs"`
    Walks           int         `json:"walks"`
    Strikeouts      int         `json:"strikeouts"`
    HomeRuns        int         `json:"home_runs"`
    ERA             float32     `json:"era"`
    BattersFaced    int         `json:"batters_faced"`
    WPA             float32     `json:"wpa"`
}

type MLBPlayByPlay struct {
    BatterIndex     string      `json:"batter_index"`
    PitcherIndex    string      `json:"pitcher_index"`
    Game            int         `json:"game"`
    Inning          int         `json:"inning"`
    Outs            int         `json:"outs"`
    Appearance      int         `json:"appearance"`
    Pitches         int         `json:"pitches"`
    Result          string      `json:"result"`
    RawResult       string      `json:"raw_result"`
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
    Threes          float32     `json:"threes"`
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
