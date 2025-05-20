package players

func getStatPchange(controlStat float32, newStat float32) float32 {
    return (newStat - controlStat) / controlStat
}

type PlayerAvg interface {
    IsValid() bool
    AddAvg(PlayerAvg) PlayerAvg
    CompareAvg(PlayerAvg) PlayerAvg
    PredictStats(PlayerAvg) PlayerAvg
    ConvertToPer() PlayerAvg
    ConvertToStats() PlayerAvg
    GetStats() map[string]float32
}

type NBAAvg struct {
    NumGames     int         `json:"num_minutes"`
    Minutes      float32     `json:"avg_minutes"`
    Points       float32     `json:"avg_points"`
    Rebounds     float32     `json:"avg_rebounds"`
    Assists      float32     `json:"avg_assists"`
    Threes       float32     `json:"avg_threes"`
    Usg          float32     `json:"avg_usg"`
    Ortg         float32     `json:"avg_drtg"`
    Drtg         float32     `json:"avg_ortg"`
}

func (n NBAAvg) IsValid() bool {
    return n.NumGames > 0
}

func (n NBAAvg) GetStats() map[string]float32 {
    return map[string]float32{
        "minutes": n.Minutes,
        "points": n.Points,
        "rebounds": n.Rebounds,
        "assists": n.Assists,
        "threes": n.Threes,
        "usg": n.Usg,
        "ortg": n.Ortg,
        "drtg": n.Drtg,
    }
}

func (n NBAAvg) AddAvg(a PlayerAvg) PlayerAvg {
    if !a.IsValid(){
        return n
    }
    nba := a.(NBAAvg)
    total_games := float32(n.NumGames + nba.NumGames)
    return NBAAvg{
        NumGames: n.NumGames + nba.NumGames,
        Minutes: (n.Minutes * float32(n.NumGames) + nba.Minutes * float32(nba.NumGames)) / total_games,
        Points: (n.Points * float32(n.NumGames) + nba.Points * float32(nba.NumGames)) / total_games,
        Rebounds: (n.Rebounds * float32(n.NumGames) + nba.Rebounds * float32(nba.NumGames)) / total_games,
        Assists: (n.Assists * float32(n.NumGames) + nba.Assists * float32(nba.NumGames)) / total_games,
        Threes: (n.Threes * float32(n.NumGames) + nba.Threes * float32(nba.NumGames)) / total_games,
        Usg: (n.Usg * float32(n.NumGames) + nba.Usg * float32(nba.NumGames)) / total_games,
        Ortg: (n.Ortg * float32(n.NumGames) + nba.Ortg * float32(nba.NumGames)) / total_games,
        Drtg: (n.Drtg * float32(n.NumGames) + nba.Drtg * float32(nba.NumGames)) / total_games,
    }
}

func (n NBAAvg) CompareAvg(controlAvg PlayerAvg) PlayerAvg {
    if !n.IsValid() {
        return n
    }
    nbaControl := controlAvg.(NBAAvg)
    return NBAAvg{
        NumGames: n.NumGames,
        Minutes: getStatPchange(nbaControl.Minutes, n.Minutes),
        Points: getStatPchange(nbaControl.Points, n.Points),
        Rebounds: getStatPchange(nbaControl.Rebounds, n.Rebounds),
        Assists: getStatPchange(nbaControl.Assists, n.Assists),
        Threes: getStatPchange(nbaControl.Threes, n.Threes),
        Usg: getStatPchange(nbaControl.Usg, n.Usg),
        Ortg: getStatPchange(nbaControl.Ortg, n.Ortg),
        Drtg: getStatPchange(nbaControl.Drtg, n.Drtg),
    }
}

func (n NBAAvg) ConvertToPer() PlayerAvg { 
    if n.IsValid() {
        return NBAAvg{
            NumGames: n.NumGames,
            Minutes: n.Minutes,
            Points: n.Points / n.Minutes,
            Rebounds: n.Rebounds / n.Minutes,
            Assists: n.Assists / n.Minutes,
            Threes: n.Threes / n.Minutes,
            Usg: n.Usg / n.Minutes,
            Ortg: n.Ortg / n.Minutes,
            Drtg: n.Drtg / n.Minutes,
        }
    } else {
        return n
    }
}

func (n NBAAvg) ConvertToStats() PlayerAvg { 
    if n.IsValid() {
        return NBAAvg{
            NumGames: n.NumGames,
            Minutes: n.Minutes,
            Points: n.Points * n.Minutes,
            Rebounds: n.Rebounds * n.Minutes,
            Assists: n.Assists * n.Minutes,
            Threes: n.Threes * n.Minutes,
            Usg: n.Usg * n.Minutes,
            Ortg: n.Ortg * n.Minutes,
            Drtg: n.Drtg * n.Minutes,
        }
    } else {
        return n
    }
}

func (n NBAAvg) PredictStats(pipFactor PlayerAvg) PlayerAvg {
    nbaPip := pipFactor.(NBAAvg)
    predictedMinutes := n.Minutes + n.Minutes * nbaPip.Minutes

    return NBAAvg{
        NumGames: nbaPip.NumGames,
        Minutes: predictedMinutes,
        Points: (n.Points + n.Points * nbaPip.Points) * predictedMinutes,
        Rebounds: (n.Rebounds + n.Rebounds * nbaPip.Rebounds) * predictedMinutes,
        Assists: (n.Assists + n.Assists * nbaPip.Assists) * predictedMinutes,
        Threes: (n.Threes + n.Threes * nbaPip.Threes) * predictedMinutes,
        Usg: (n.Usg + n.Usg * nbaPip.Usg) * predictedMinutes,
        Ortg: (n.Ortg + n.Ortg * nbaPip.Ortg) * predictedMinutes,
        Drtg: (n.Drtg + n.Drtg * nbaPip.Drtg) * predictedMinutes,
    }
}

type MLBBattingAvg struct {
    NumGames        int         `json:"num_games"`
    AtBats          float32     `json:"avg_at_bats"`
    Runs            float32     `json:"avg_runs"`
    Hits            float32     `json:"avg_hits"`
    RBIs            float32     `json:"avg_rbis"`
    HomeRuns        float32     `json:"avg_home_runs"`
    Walks           float32     `json:"avg_walks"`
    Strikeouts      float32     `json:"avg_strikeouts"`
    PAs             float32     `json:"avg_pas"`
    Pitches         float32     `json:"avg_pitches"`
    Strikes         float32     `json:"avg_strikes"`
    BA              float32     `json:"avg_ba"`
    OBP             float32     `json:"avg_obp"`
    SLG             float32     `json:"avg_slg"`
    OPS             float32     `json:"avg_ops"`
    WPA             float32     `json:"avg_wpa"`
}

func (m MLBBattingAvg) IsValid() bool {
    return m.NumGames > 0
}

func (m MLBBattingAvg) GetStats() map[string]float32 {
    return map[string]float32{
        "at_bats": m.AtBats,
        "runs": m.Runs,
        "hits": m.Hits,
        "rbis": m.RBIs,
        "home_runs": m.HomeRuns,
        "walks": m.Walks,
        "strikeouts": m.Strikeouts,
        "pas": m.PAs,
        "pitches": m.Pitches,
        "strikes": m.Strikes,
        "ba": m.BA,
        "obp": m.OBP,
        "slg": m.SLG,
        "ops": m.OPS,
        "wpa": m.WPA,
    }
}

func (m MLBBattingAvg) AddAvg(a PlayerAvg) PlayerAvg {
    if !a.IsValid() {
        return m
    }
    mlb := a.(MLBBattingAvg)
    total_games := float32(m.NumGames + mlb.NumGames)
    return MLBBattingAvg{
        NumGames: m.NumGames + mlb.NumGames,
        AtBats: (m.AtBats * float32(m.NumGames) + mlb.AtBats * float32(mlb.NumGames)) / total_games,
        Runs: (m.Runs * float32(m.NumGames) + mlb.Runs * float32(mlb.NumGames)) / total_games,
        Hits: (m.Hits * float32(m.NumGames) + mlb.Hits * float32(mlb.NumGames)) / total_games,
        RBIs: (m.RBIs * float32(m.NumGames) + mlb.RBIs * float32(mlb.NumGames)) / total_games,
        HomeRuns: (m.HomeRuns * float32(m.NumGames) + mlb.HomeRuns * float32(mlb.NumGames)) / total_games,
        Walks: (m.Walks * float32(m.NumGames) + mlb.Walks * float32(mlb.NumGames)) / total_games,
        Strikeouts: (m.Strikeouts * float32(m.NumGames) + mlb.Strikeouts * float32(mlb.NumGames)) / total_games,
        PAs: (m.PAs * float32(m.NumGames) + mlb.PAs * float32(mlb.NumGames)) / total_games,
        Pitches: (m.Pitches * float32(m.NumGames) + mlb.Pitches * float32(mlb.NumGames)) / total_games,
        Strikes: (m.Strikes * float32(m.NumGames) + mlb.Strikes * float32(mlb.NumGames)) / total_games,
        BA: (m.BA * float32(m.NumGames) + mlb.BA * float32(mlb.NumGames)) / total_games,
        OBP: (m.OBP * float32(m.NumGames) + mlb.OBP * float32(mlb.NumGames)) / total_games,
        SLG: (m.SLG * float32(m.NumGames) + mlb.SLG * float32(mlb.NumGames)) / total_games,
        OPS: (m.OPS * float32(m.NumGames) + mlb.OPS * float32(mlb.NumGames)) / total_games,
        WPA: (m.WPA * float32(m.NumGames) + mlb.WPA * float32(mlb.NumGames)) / total_games,
    }
}

func (m MLBBattingAvg) CompareAvg(controlAvg PlayerAvg) PlayerAvg {
    if !m.IsValid() {
        return m
    }
    mlbControl := controlAvg.(MLBBattingAvg)
    return MLBBattingAvg{
        NumGames: m.NumGames,
        AtBats: getStatPchange(mlbControl.AtBats, m.AtBats),
        Runs: getStatPchange(mlbControl.Runs, m.Runs),
        Hits: getStatPchange(mlbControl.Hits, m.Hits),
        RBIs: getStatPchange(mlbControl.RBIs, m.RBIs),
        HomeRuns: getStatPchange(mlbControl.HomeRuns, m.HomeRuns),
        Walks: getStatPchange(mlbControl.Walks, m.Walks),
        Strikeouts: getStatPchange(mlbControl.Strikeouts, m.Strikeouts),
        PAs: getStatPchange(mlbControl.PAs, m.PAs),
        Pitches: getStatPchange(mlbControl.Pitches, m.Pitches),
        Strikes: getStatPchange(mlbControl.Strikes, m.Strikes),
        OBP: getStatPchange(mlbControl.OBP, m.OBP),
        SLG: getStatPchange(mlbControl.SLG, m.SLG),
        OPS: getStatPchange(mlbControl.OPS, m.OPS),
        WPA: getStatPchange(mlbControl.WPA, m.WPA),
    }
}

func (m MLBBattingAvg) ConvertToPer() PlayerAvg {
    if m.IsValid() {
        // For baseball, "per" is usually per plate appearance (PA)
        return MLBBattingAvg{
            NumGames: m.NumGames,
            PAs: m.PAs,
            AtBats: m.AtBats / m.PAs,
            Runs: m.Runs / m.PAs,
            Hits: m.Hits / m.PAs,
            RBIs: m.RBIs / m.PAs,
            HomeRuns: m.HomeRuns / m.PAs,
            Walks: m.Walks / m.PAs,
            Strikeouts: m.Strikeouts / m.PAs,
            Pitches: m.Pitches / m.PAs,
            Strikes: m.Strikes / m.PAs,
            OBP: m.OBP,        // OBP, SLG, OPS are already rate stats
            SLG: m.SLG,        // so we don't need to convert them
            OPS: m.OPS,
            WPA: m.WPA / m.PAs,
        }
    } else {
        return m
    }
}

func (m MLBBattingAvg) ConvertToStats() PlayerAvg {
    if m.IsValid() {
        // Convert from per-PA back to raw stats
        return MLBBattingAvg{
            NumGames: m.NumGames,
            AtBats: m.AtBats * m.PAs,
            Runs: m.Runs * m.PAs,
            Hits: m.Hits * m.PAs,
            RBIs: m.RBIs * m.PAs,
            HomeRuns: m.HomeRuns * m.PAs,
            Walks: m.Walks * m.PAs,
            Strikeouts: m.Strikeouts * m.PAs,
            PAs: m.PAs,
            Pitches: m.Pitches * m.PAs,
            Strikes: m.Strikes * m.PAs,
            OBP: m.OBP,        // These stay the same as they are rate stats
            SLG: m.SLG,
            OPS: m.OPS,
            WPA: m.WPA * m.PAs,
        }
    } else {
        return m
    }
}

func (m MLBBattingAvg) PredictStats(pipFactor PlayerAvg) PlayerAvg {
    mlbPip := pipFactor.(MLBBattingAvg)
    predictedPAs := m.PAs + m.PAs * mlbPip.PAs

    return MLBBattingAvg{
        NumGames: mlbPip.NumGames,
        PAs: predictedPAs,
        AtBats: (m.AtBats + m.AtBats * mlbPip.AtBats) * predictedPAs,
        Runs: (m.Runs + m.Runs * mlbPip.Runs) * predictedPAs,
        Hits: (m.Hits + m.Hits * mlbPip.Hits) * predictedPAs,
        RBIs: (m.RBIs + m.RBIs * mlbPip.RBIs) * predictedPAs,
        HomeRuns: (m.HomeRuns + m.HomeRuns * mlbPip.HomeRuns) * predictedPAs,
        Walks: (m.Walks + m.Walks * mlbPip.Walks) * predictedPAs,
        Strikeouts: (m.Strikeouts + m.Strikeouts * mlbPip.Strikeouts) * predictedPAs,
        Pitches: (m.Pitches + m.Pitches * mlbPip.Pitches) * predictedPAs,
        Strikes: (m.Strikes + m.Strikes * mlbPip.Strikes) * predictedPAs,
        OBP: m.OBP + m.OBP * mlbPip.OBP,
        SLG: m.SLG + m.SLG * mlbPip.SLG,
        OPS: m.OPS + m.OPS * mlbPip.OPS,
        WPA: (m.WPA + m.WPA * mlbPip.WPA) * predictedPAs,
    }
}
